// Lección 13 · El proyecto: inventario, la auditoría de una cuenta en varias regiones.
//
// Todo lo del curso en una herramienta:
//   - dice cuenta y regiones antes de nada, y se niega si la cuenta no es la esperada (1);
//   - recorre todas las páginas y cuenta instancias, no reservas (2);
//   - «sin dueño» mirado en Go, porque no es un filtro (3);
//   - grupos abiertos al mundo en puertos de administración (5);
//   - lo que cobra sin que nadie lo use: volúmenes sueltos, IPs sin asociar, NAT gateways (9);
//   - solo LEE: no para ni borra nada (8).
//
// Registro a stderr, informe en texto o JSON, y códigos de salida:
// 0 todo bien · 1 hay hallazgos · 2 no pude ni mirar.
//
//	go run ./13_proyecto -regiones us-east-1,us-west-2 -cuenta 123456789012
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Hallazgo struct {
	Region  string `json:"region"`
	Tipo    string `json:"tipo"`
	Recurso string `json:"recurso"`
	Detalle string `json:"detalle"`
}

type Informe struct {
	Cuenta     string         `json:"cuenta"`
	Instancias map[string]int `json:"instancias_por_region"`
	Hallazgos  []Hallazgo     `json:"hallazgos"`
}

var puertosAdmin = map[int32]string{22: "ssh", 3389: "rdp", 5432: "postgres", 3306: "mysql"}

func revisarRegion(ctx context.Context, region string, inf *Informe) error {
	c := comun.EC2(region)
	add := func(tipo, rec, det string) {
		inf.Hallazgos = append(inf.Hallazgos, Hallazgo{region, tipo, rec, det})
	}
	var n int
	p := ec2.NewDescribeInstancesPaginator(c, &ec2.DescribeInstancesInput{})
	for p.HasMorePages() {
		pg, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, r := range pg.Reservations {
			for _, i := range r.Instances {
				n++
				if comun.Etiqueta(i.Tags, "dueno") == "-" {
					add("sin-dueno", aws.ToString(i.InstanceId), "Name="+comun.Etiqueta(i.Tags, "Name"))
				}
			}
		}
	}
	inf.Instancias[region] = n

	sgs, err := c.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return err
	}
	for _, g := range sgs.SecurityGroups {
		for _, perm := range g.IpPermissions {
			for _, r := range perm.IpRanges {
				if nombre, ok := puertosAdmin[aws.ToInt32(perm.FromPort)]; ok && aws.ToString(r.CidrIp) == "0.0.0.0/0" {
					add("abierto-al-mundo", aws.ToString(g.GroupId),
						fmt.Sprintf("%s: %d (%s) desde 0.0.0.0/0", aws.ToString(g.GroupName), aws.ToInt32(perm.FromPort), nombre))
				}
			}
		}
	}

	vols, err := c.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{})
	if err != nil {
		return err
	}
	for _, v := range vols.Volumes {
		if len(v.Attachments) == 0 {
			add("volumen-suelto", aws.ToString(v.VolumeId), fmt.Sprintf("%d GiB, sin instancia", aws.ToInt32(v.Size)))
		}
	}

	ips, err := c.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		return err
	}
	for _, a := range ips.Addresses {
		if a.AssociationId == nil {
			add("ip-sin-usar", aws.ToString(a.AllocationId), aws.ToString(a.PublicIp)+" reservada y sin asociar")
		}
	}

	nats, err := c.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return err
	}
	for _, g := range nats.NatGateways {
		add("nat-gateway", aws.ToString(g.NatGatewayId), "se cobra por hora, pase trafico o no")
	}
	return nil
}

func main() {
	regiones := flag.String("regiones", "us-east-1,us-west-2", "regiones a revisar, separadas por comas")
	cuenta := flag.String("cuenta", "", "cuenta esperada; si no coincide, no miro nada")
	comoJSON := flag.Bool("json", false, "informe en JSON")
	flag.Parse()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}})))
	ctx := context.Background()

	id, err := sts.NewFromConfig(comun.Config("us-east-1")).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		slog.Error("no se quien soy", "err", err)
		os.Exit(2)
	}
	if *cuenta != "" && aws.ToString(id.Account) != *cuenta {
		slog.Error("cuenta inesperada", "esperaba", *cuenta, "es", aws.ToString(id.Account))
		os.Exit(2)
	}
	inf := &Informe{Cuenta: aws.ToString(id.Account), Instancias: map[string]int{}}
	for _, r := range strings.Split(*regiones, ",") {
		if err := revisarRegion(ctx, r, inf); err != nil {
			slog.Error("no puedo revisar", "region", r, "err", err)
			os.Exit(2)
		}
	}

	if *comoJSON {
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(inf)
	} else {
		fmt.Printf("Cuenta %s\n", inf.Cuenta)
		for _, r := range strings.Split(*regiones, ",") {
			fmt.Printf("  %-10s %d instancias\n", r, inf.Instancias[r])
		}
		for _, h := range inf.Hallazgos {
			fmt.Printf("  %-10s %-17s %-22s %s\n", h.Region, h.Tipo, h.Recurso, h.Detalle)
		}
		if len(inf.Hallazgos) == 0 {
			fmt.Println("  sin hallazgos")
		}
	}
	if len(inf.Hallazgos) > 0 {
		slog.Warn("hay hallazgos", "total", len(inf.Hallazgos))
		os.Exit(1)
	}
}
