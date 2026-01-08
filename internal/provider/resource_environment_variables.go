package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &EnvironmentVariablesResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariablesResource{}

func NewEnvironmentVariablesResource() resource.Resource {
	return &EnvironmentVariablesResource{}
}

type EnvironmentVariablesResource struct {
	client *client.DokployClient
}

type EnvironmentVariablesResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Variables     types.Map    `tfsdk:"variables"`
	CreateEnvFile types.Bool   `tfsdk:"create_env_file"`
}

func (r *EnvironmentVariablesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variables"
}

func (r *EnvironmentVariablesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages all environment variables for a Dokploy application as a single resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"application_id": schema.StringAttribute{
				Required: true,
			},
			"variables": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Sensitive:   true,
			},
			"create_env_file": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

func (r *EnvironmentVariablesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *EnvironmentVariablesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentVariablesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envMap := make(map[string]string)
	diags = plan.Variables.ElementsAs(ctx, &envMap, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateApplicationEnv(plan.ApplicationID.ValueString(), func(m map[string]string) {
		for k, v := range envMap {
			m[k] = v
		}
	}, plan.CreateEnvFile.ValueBoolPointer())

	if err != nil {
		resp.Diagnostics.AddError("Error creating environment variables", err.Error())
		return
	}

	plan.ID = plan.ApplicationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentVariablesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(state.ApplicationID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	envMap := client.ParseEnv(app.Env)
	state.Variables, diags = types.MapValueFrom(ctx, types.StringType, envMap)
	resp.Diagnostics.Append(diags...)

	// The CreateEnvFile attribute is not stored in the API, so we keep the configured value.
	// If it's not configured, Terraform will use the default.

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EnvironmentVariablesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envMap := make(map[string]string)
	diags = plan.Variables.ElementsAs(ctx, &envMap, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateApplicationEnv(plan.ApplicationID.ValueString(), func(m map[string]string) {
		// Clear existing vars and set new ones
		for k := range m {
			delete(m, k)
		}
		for k, v := range envMap {
			m[k] = v
		}
	}, plan.CreateEnvFile.ValueBoolPointer())

	if err != nil {
		resp.Diagnostics.AddError("Error updating environment variables", err.Error())
		return
	}

	plan.ID = plan.ApplicationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentVariablesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentVariablesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateApplicationEnv(state.ApplicationID.ValueString(), func(m map[string]string) {
		for k := range m {
			delete(m, k)
		}
	}, state.CreateEnvFile.ValueBoolPointer())

	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting environment variables", err.Error())
		return
	}
}

func (r *EnvironmentVariablesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}
