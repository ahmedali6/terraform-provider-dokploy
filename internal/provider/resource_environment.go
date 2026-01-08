package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &EnvironmentResource{}
var _ resource.ResourceWithImportState = &EnvironmentResource{}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

type EnvironmentResource struct {
	client *client.DokployClient
}

type EnvironmentResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.CreateEnvironment(plan.ProjectID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		// Handle "Already exists" logic
		if strings.Contains(strings.ToLower(err.Error()), "already exists") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			// Fetch project to find the existing environment
			project, pErr := r.client.GetProject(plan.ProjectID.ValueString())
			if pErr != nil {
				resp.Diagnostics.AddError("Error creating environment (and failed to fetch project for recovery)", fmt.Sprintf("Create error: %s. Fetch error: %s", err, pErr))
				return
			}

			// Find env by name
			found := false
			for _, e := range project.Environments {
				if e.Name == plan.Name.ValueString() {
					plan.ID = types.StringValue(e.ID)
					plan.Description = types.StringValue(e.Description)
					found = true
					break
				}
			}

			if !found {
				resp.Diagnostics.AddError("Error creating environment", fmt.Sprintf("API reported duplicate, but could not find environment with name '%s' in project '%s'. Original error: %s", plan.Name.ValueString(), plan.ProjectID.ValueString(), err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Error creating environment", err.Error())
			return
		}
	} else {
		plan.ID = types.StringValue(env.ID)
		plan.Description = types.StringValue(env.Description)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Environments are read via Project
	project, err := r.client.GetProject(state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading parent project", err.Error())
		return
	}

	found := false
	for _, env := range project.Environments {
		if env.ID == state.ID.ValueString() {
			state.Name = types.StringValue(env.Name)
			state.Description = types.StringValue(env.Description)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := client.Environment{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ProjectID:   plan.ProjectID.ValueString(),
	}

	updatedEnv, err := r.client.UpdateEnvironment(env)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment", err.Error())
		return
	}

	plan.Name = types.StringValue(updatedEnv.Name)
	plan.Description = types.StringValue(updatedEnv.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironment(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
		return
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
