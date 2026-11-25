package main

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// El recurso notas_nota.
type recursoNota struct{ c *cliente }

var (
	_ resource.ResourceWithConfigure   = &recursoNota{}
	_ resource.ResourceWithImportState = &recursoNota{}
)

func nuevaNota() resource.Resource { return &recursoNota{} }

// El modelo: un campo por atributo del esquema, unidos por la etiqueta tfsdk.
type notaModelo struct {
	ID        types.String `tfsdk:"id"`
	Titulo    types.String `tfsdk:"titulo"`
	Texto     types.String `tfsdk:"texto"`
	Etiquetas []string     `tfsdk:"etiquetas"`
	Clave     types.String `tfsdk:"clave"`
	Creada    types.String `tfsdk:"creada"`
	Version   types.Int64  `tfsdk:"version"`
}

func (r *recursoNota) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nota"
}

func (r *recursoNota) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"titulo": schema.StringAttribute{Required: true},
		"texto":  esquemaTexto(),
		// La API guarda las etiquetas ordenadas: el orden no importa, así que es un conjunto.
		"etiquetas": esquemaEtiquetas(),
		// La API la guarda pero nunca la devuelve.
		"clave":  schema.StringAttribute{Optional: true, Sensitive: true},
		"creada": schema.StringAttribute{Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"version": schema.Int64Attribute{Computed: true},
	}}
}

func esquemaTexto() schema.Attribute {
	if trampa("vacio") {
		return schema.StringAttribute{Optional: true}
	}
	// La API devuelve "" cuando no hay texto: con Default "", null y "" no chocan.
	return schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")}
}

func esquemaEtiquetas() schema.Attribute {
	if trampa("lista") {
		return schema.ListAttribute{Optional: true, ElementType: types.StringType}
	}
	return schema.SetAttribute{Optional: true, ElementType: types.StringType}
}

func (r *recursoNota) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.c = req.ProviderData.(*cliente)
	}
}

// Lo que se manda a la API sale del plan.
func haciaAPI(m notaModelo) nota {
	return nota{Titulo: m.Titulo.ValueString(), Texto: m.Texto.ValueString(),
		Etiquetas: m.Etiquetas, Clave: m.Clave.ValueString()}
}

// Lo que vuelve de la API va al estado. La clave no: la API nunca la devuelve.
func desdeAPI(m *notaModelo, n nota) {
	m.ID = types.StringValue(n.ID)
	m.Titulo = types.StringValue(n.Titulo)
	m.Texto = types.StringValue(n.Texto)
	m.Etiquetas = n.Etiquetas
	m.Creada = types.StringValue(n.Creada)
	m.Version = types.Int64Value(n.Version)
}

func (r *recursoNota) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notaModelo
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	n, err := r.c.crear(ctx, haciaAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Creando la nota", err.Error())
		return
	}
	desdeAPI(&plan, n)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *recursoNota) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var estado notaModelo
	resp.Diagnostics.Append(req.State.Get(ctx, &estado)...)
	if resp.Diagnostics.HasError() {
		return
	}
	n, err := r.c.leer(ctx, estado.ID.ValueString())
	if errors.Is(err, errNoExiste) && !trampa("404") {
		resp.State.RemoveResource(ctx) // la borraron a mano: el plan la vuelve a crear
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Leyendo la nota", err.Error())
		return
	}
	desdeAPI(&estado, n)
	if trampa("clave") {
		estado.Clave = types.StringValue(n.Clave) // la API no la devuelve: llega ""
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &estado)...)
}

func (r *recursoNota) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, estado notaModelo
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &estado)...)
	if resp.Diagnostics.HasError() {
		return
	}
	n, err := r.c.cambiar(ctx, estado.ID.ValueString(), haciaAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Cambiando la nota", err.Error())
		return
	}
	desdeAPI(&plan, n)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *recursoNota) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var estado notaModelo
	resp.Diagnostics.Append(req.State.Get(ctx, &estado)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.c.borrar(ctx, estado.ID.ValueString())
	if err != nil && !errors.Is(err, errNoExiste) {
		resp.Diagnostics.AddError("Borrando la nota", err.Error())
	}
}

// terraform import notas_nota.x <id>: el id va al estado y Read rellena lo demás.
func (r *recursoNota) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
