package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &ComposeResource{}
var _ resource.ResourceWithImportState = &ComposeResource{}

func NewComposeResource() resource.Resource {
	return &ComposeResource{}
}

type ComposeResource struct {
	client *client.DokployClient
}

type ComposeResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	Name               types.String `tfsdk:"name"`
	ComposeFileContent types.String `tfsdk:"compose_file_content"`
	SourceType         types.String `tfsdk:"source_type"`
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	ComposePath        types.String `tfsdk:"compose_path"`
	AutoDeploy         types.Bool   `tfsdk:"auto_deploy"`
	DeployOnCreate     types.Bool   `tfsdk:"deploy_on_create"`
	ServerID           types.String `tfsdk:"server_id"`
}

func (r *ComposeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compose"
}

func (r *ComposeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"compose_file_content": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"source_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_git_url": schema.StringAttribute{
				Optional: true,
			},
			"custom_git_branch": schema.StringAttribute{
				Optional: true,
			},
			"custom_git_ssh_key_id": schema.StringAttribute{
				Optional: true,
			},
			"compose_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"deploy_on_create": schema.BoolAttribute{
				Optional: true,
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID to deploy the compose stack to. If not specified, deploys to the default server.",
			},
		},
	}
}

func (r *ComposeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComposeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ComposeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ComposePath.IsUnknown() || plan.ComposePath.IsNull() {
		plan.ComposePath = types.StringValue("./docker-compose.yml")
	}

	if plan.SourceType.IsUnknown() || plan.SourceType.IsNull() {
		if !plan.CustomGitUrl.IsNull() && !plan.CustomGitUrl.IsUnknown() && plan.CustomGitUrl.ValueString() != "" {
			plan.SourceType = types.StringValue("git")
		} else if !plan.ComposeFileContent.IsNull() && !plan.ComposeFileContent.IsUnknown() && plan.ComposeFileContent.ValueString() != "" {
			plan.SourceType = types.StringValue("raw")
		} else {
			plan.SourceType = types.StringValue("github")
		}
	}

	comp := client.Compose{
		Name:              plan.Name.ValueString(),
		EnvironmentID:     plan.EnvironmentID.ValueString(),
		ComposeFile:       plan.ComposeFileContent.ValueString(),
		SourceType:        plan.SourceType.ValueString(),
		CustomGitUrl:      plan.CustomGitUrl.ValueString(),
		CustomGitBranch:   plan.CustomGitBranch.ValueString(),
		CustomGitSSHKeyId: plan.CustomGitSSHKeyID.ValueString(),
		ComposePath:       plan.ComposePath.ValueString(),
		AutoDeploy:        plan.AutoDeploy.ValueBool(),
		ServerID:          plan.ServerID.ValueString(),
	}

	createdComp, err := r.client.CreateCompose(comp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating compose", err.Error())
		return
	}

	plan.ID = types.StringValue(createdComp.ID)
	plan.SourceType = types.StringValue(createdComp.SourceType)
	plan.ComposePath = types.StringValue(createdComp.ComposePath)
	plan.AutoDeploy = types.BoolValue(createdComp.AutoDeploy)
	if createdComp.ComposeFile != "" {
		plan.ComposeFileContent = types.StringValue(createdComp.ComposeFile)
	} else {
		plan.ComposeFileContent = types.StringNull()
	}

	if !plan.DeployOnCreate.IsNull() && plan.DeployOnCreate.ValueBool() {
		err := r.client.DeployCompose(createdComp.ID, plan.ServerID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning("Deployment Trigger Failed", fmt.Sprintf("Compose stack created but deployment failed to trigger: %s", err.Error()))
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ComposeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	comp, err := r.client.GetCompose(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading compose", err.Error())
		return
	}

	state.Name = types.StringValue(comp.Name)
	state.EnvironmentID = types.StringValue(comp.EnvironmentID)
	state.ComposeFileContent = types.StringValue(comp.ComposeFile)
	state.SourceType = types.StringValue(comp.SourceType)
	state.CustomGitUrl = types.StringValue(comp.CustomGitUrl)
	state.CustomGitBranch = types.StringValue(comp.CustomGitBranch)
	state.CustomGitSSHKeyID = types.StringValue(comp.CustomGitSSHKeyId)
	state.ComposePath = types.StringValue(comp.ComposePath)
	state.AutoDeploy = types.BoolValue(comp.AutoDeploy)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ComposeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	comp := client.Compose{
		ID:                plan.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		EnvironmentID:     plan.EnvironmentID.ValueString(),
		ComposeFile:       plan.ComposeFileContent.ValueString(),
		SourceType:        plan.SourceType.ValueString(),
		CustomGitUrl:      plan.CustomGitUrl.ValueString(),
		CustomGitBranch:   plan.CustomGitBranch.ValueString(),
		CustomGitSSHKeyId: plan.CustomGitSSHKeyID.ValueString(),
		ComposePath:       plan.ComposePath.ValueString(),
		AutoDeploy:        plan.AutoDeploy.ValueBool(),
	}

	updatedComp, err := r.client.UpdateCompose(comp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating compose", err.Error())
		return
	}

	plan.Name = types.StringValue(updatedComp.Name)
	plan.ComposeFileContent = types.StringValue(updatedComp.ComposeFile)
	plan.SourceType = types.StringValue(updatedComp.SourceType)
	plan.AutoDeploy = types.BoolValue(updatedComp.AutoDeploy)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ComposeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCompose(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting compose", err.Error())
		return
	}
}

func (r *ComposeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
