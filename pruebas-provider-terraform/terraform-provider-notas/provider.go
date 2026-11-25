package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notasProvider struct{}

type provModelo struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *notasProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "notas" // los recursos se llaman notas_<algo>
	resp.Version = "0.1.0"
}

func (p *notasProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"endpoint": schema.StringAttribute{Optional: true, Description: "URL de notas-api; si falta, NOTAS_ENDPOINT"},
	}}
}

func (p *notasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var m provModelo
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ep := os.Getenv("NOTAS_ENDPOINT")
	if !m.Endpoint.IsNull() {
		ep = m.Endpoint.ValueString()
	}
	if ep == "" {
		resp.Diagnostics.AddError("Falta endpoint", "Pon endpoint en el bloque provider o NOTAS_ENDPOINT.")
		return
	}
	// Lo que se pone aquí les llega a los recursos en su Configure.
	resp.ResourceData = &cliente{base: ep, http: &http.Client{Timeout: 10 * time.Second}}
}

func (p *notasProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{nuevaNota}
}

func (p *notasProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
