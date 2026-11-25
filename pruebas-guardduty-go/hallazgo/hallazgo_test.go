package hallazgo

import "testing"

func TestParseTipo(t *testing.T) {
	casos := []struct {
		entrada string
		quiere  Tipo
	}{
		{"Recon:EC2/PortProbeUnprotectedPort", Tipo{"Recon", "EC2", "PortProbeUnprotectedPort", "", ""}},
		{"Backdoor:EC2/C&CActivity.B!DNS", Tipo{"Backdoor", "EC2", "C&CActivity", "B", "DNS"}},
		{"CryptoCurrency:EC2/BitcoinTool.B!DNS", Tipo{"CryptoCurrency", "EC2", "BitcoinTool", "B", "DNS"}},
		{"Backdoor:EC2/DenialOfService.UdpOnTcpPorts", Tipo{"Backdoor", "EC2", "DenialOfService", "UdpOnTcpPorts", ""}},
		{"UnauthorizedAccess:IAMUser/InstanceCredentialExfiltration.OutsideAWS",
			Tipo{"UnauthorizedAccess", "IAMUser", "InstanceCredentialExfiltration", "OutsideAWS", ""}},
		{"UnauthorizedAccess:EC2/MaliciousIPCaller.Custom", Tipo{"UnauthorizedAccess", "EC2", "MaliciousIPCaller", "Custom", ""}},
		{"AttackSequence:IAM/CompromisedCredentials", Tipo{"AttackSequence", "IAM", "CompromisedCredentials", "", ""}},
	}
	for _, c := range casos {
		got, err := ParseTipo(c.entrada)
		if err != nil {
			t.Fatalf("%s: %v", c.entrada, err)
		}
		if got != c.quiere {
			t.Errorf("%s:\n got  %+v\n want %+v", c.entrada, got, c.quiere)
		}
	}
}

func TestParseTipoMalo(t *testing.T) {
	for _, s := range []string{"", "Recon", "Recon:EC2", "Recon:/Port", ":EC2/Port"} {
		if _, err := ParseTipo(s); err == nil {
			t.Errorf("%q: esperaba error", s)
		}
	}
}

func TestTramoDe(t *testing.T) {
	casos := map[float64]Tramo{1: Low, 3.9: Low, 4: Medium, 6.9: Medium, 7: High, 8.9: High, 9: Critical, 10: Critical}
	for sev, quiere := range casos {
		got, err := TramoDe(sev)
		if err != nil || got != quiere {
			t.Errorf("TramoDe(%.1f) = %q, %v; quiere %q", sev, got, err, quiere)
		}
	}
	for _, sev := range []float64{0, 0.5, 10.1} {
		if _, err := TramoDe(sev); err == nil {
			t.Errorf("TramoDe(%.1f): esperaba error", sev)
		}
	}
}
