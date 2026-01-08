package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &RegistryResource{}
var _ resource.ResourceWithImportState = &RegistryResource{}

func NewRegistryResource() resource.Resource {
	return &RegistryResource{}
}

type RegistryResource struct {
	client *client.DokployClient
}

type RegistryResourceModel struct {
	ID           types.String `tfsdk:"id"`
	RegistryName types.String `tfsdk:"registry_name"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	RegistryUrl  types.String `tfsdk:"registry_url"`
	RegistryType types.String `tfsdk:"registry_type"`
	ImagePrefix  types.String `tfsdk:"image_prefix"`
	ServerID     types.String `tfsdk:"server_id"`
}

func (r *RegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *RegistryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Docker registry configuration in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the registry.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registry_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the registry.",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Username for the registry.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for the registry.",
			},
			"registry_url": schema.StringAttribute{
				Required:    true,
				Description: "URL of the registry (e.g., ghcr.io, docker.io).",
			},
			"registry_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type of registry. Currently only 'cloud' is supported.",
				Default:     stringdefault.StaticString("cloud"),
			},
			"image_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Image prefix for the registry (e.g., ghcr.io/myorg).",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID to associate the registry with (optional).",
			},
		},
	}
}

func (r *RegistryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DokployClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *RegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RegistryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry := client.Registry{
		RegistryName: plan.RegistryName.ValueString(),
		Username:     plan.Username.ValueString(),
		Password:     plan.Password.ValueString(),
		RegistryUrl:  plan.RegistryUrl.ValueString(),
		RegistryType: plan.RegistryType.ValueString(),
		ImagePrefix:  plan.ImagePrefix.ValueString(),
		ServerID:     plan.ServerID.ValueString(),
	}

	createdRegistry, err := r.client.CreateRegistry(registry)
	if err != nil {
		resp.Diagnostics.AddError("Error creating registry", err.Error())
		return
	}

	plan.ID = types.StringValue(createdRegistry.ID)
	plan.RegistryType = types.StringValue(createdRegistry.RegistryType)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.GetRegistry(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading registry", err.Error())
		return
	}

	state.RegistryName = types.StringValue(registry.RegistryName)
	state.Username = types.StringValue(registry.Username)
	// Don't update password from API as it might not be returned
	state.RegistryUrl = types.StringValue(registry.RegistryUrl)
	state.RegistryType = types.StringValue(registry.RegistryType)
	state.ImagePrefix = types.StringValue(registry.ImagePrefix)
	if registry.ServerID != "" {
		state.ServerID = types.StringValue(registry.ServerID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RegistryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry := client.Registry{
		ID:           plan.ID.ValueString(),
		RegistryName: plan.RegistryName.ValueString(),
		Username:     plan.Username.ValueString(),
		Password:     plan.Password.ValueString(),
		RegistryUrl:  plan.RegistryUrl.ValueString(),
		RegistryType: plan.RegistryType.ValueString(),
		ImagePrefix:  plan.ImagePrefix.ValueString(),
		ServerID:     plan.ServerID.ValueString(),
	}

	updatedRegistry, err := r.client.UpdateRegistry(registry)
	if err != nil {
		resp.Diagnostics.AddError("Error updating registry", err.Error())
		return
	}

	plan.RegistryName = types.StringValue(updatedRegistry.RegistryName)
	plan.Username = types.StringValue(updatedRegistry.Username)
	plan.RegistryUrl = types.StringValue(updatedRegistry.RegistryUrl)
	plan.RegistryType = types.StringValue(updatedRegistry.RegistryType)
	plan.ImagePrefix = types.StringValue(updatedRegistry.ImagePrefix)
	if updatedRegistry.ServerID != "" {
		plan.ServerID = types.StringValue(updatedRegistry.ServerID)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRegistry(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting registry", err.Error())
		return
	}
}

func (r *RegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
