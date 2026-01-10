package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	ServerID      types.String `tfsdk:"server_id"`

	// Compose file
	ComposeFileContent types.String `tfsdk:"compose_file_content"`
	ComposePath        types.String `tfsdk:"compose_path"`
	ComposeType        types.String `tfsdk:"compose_type"`

	// Source configuration
	SourceType types.String `tfsdk:"source_type"`

	// Custom Git provider settings (source_type = "git")
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	CustomGitBuildPath types.String `tfsdk:"custom_git_build_path"`
	EnableSubmodules   types.Bool   `tfsdk:"enable_submodules"`

	// GitHub provider settings (source_type = "github")
	Repository  types.String `tfsdk:"repository"`
	Branch      types.String `tfsdk:"branch"`
	Owner       types.String `tfsdk:"owner"`
	GithubId    types.String `tfsdk:"github_id"`
	TriggerType types.String `tfsdk:"trigger_type"`

	// GitLab provider settings (source_type = "gitlab")
	GitlabId            types.String `tfsdk:"gitlab_id"`
	GitlabProjectId     types.Int64  `tfsdk:"gitlab_project_id"`
	GitlabRepository    types.String `tfsdk:"gitlab_repository"`
	GitlabOwner         types.String `tfsdk:"gitlab_owner"`
	GitlabBranch        types.String `tfsdk:"gitlab_branch"`
	GitlabBuildPath     types.String `tfsdk:"gitlab_build_path"`
	GitlabPathNamespace types.String `tfsdk:"gitlab_path_namespace"`

	// Bitbucket provider settings (source_type = "bitbucket")
	BitbucketId         types.String `tfsdk:"bitbucket_id"`
	BitbucketRepository types.String `tfsdk:"bitbucket_repository"`
	BitbucketOwner      types.String `tfsdk:"bitbucket_owner"`
	BitbucketBranch     types.String `tfsdk:"bitbucket_branch"`
	BitbucketBuildPath  types.String `tfsdk:"bitbucket_build_path"`

	// Gitea provider settings (source_type = "gitea")
	GiteaId         types.String `tfsdk:"gitea_id"`
	GiteaRepository types.String `tfsdk:"gitea_repository"`
	GiteaOwner      types.String `tfsdk:"gitea_owner"`
	GiteaBranch     types.String `tfsdk:"gitea_branch"`
	GiteaBuildPath  types.String `tfsdk:"gitea_build_path"`

	// Environment
	Env types.String `tfsdk:"env"`

	// Runtime configuration
	AutoDeploy types.Bool `tfsdk:"auto_deploy"`

	// Advanced configuration
	Command                   types.String `tfsdk:"command"`
	Suffix                    types.String `tfsdk:"suffix"`
	Randomize                 types.Bool   `tfsdk:"randomize"`
	IsolatedDeployment        types.Bool   `tfsdk:"isolated_deployment"`
	IsolatedDeploymentsVolume types.Bool   `tfsdk:"isolated_deployments_volume"`
	WatchPaths                types.List   `tfsdk:"watch_paths"`

	// Computed status
	ComposeStatus types.String `tfsdk:"compose_status"`
	RefreshToken  types.String `tfsdk:"refresh_token"`
	CreatedAt     types.String `tfsdk:"created_at"`

	// Deployment options
	DeployOnCreate types.Bool `tfsdk:"deploy_on_create"`
}

func (r *ComposeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compose"
}

func (r *ComposeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy compose stack. Supports multiple source types including GitHub, GitLab, Bitbucket, Gitea, custom Git repositories, and raw compose file content.",
		Attributes: map[string]schema.Attribute{
			// Core attributes
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the compose stack.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The environment ID this compose stack belongs to. Can be changed to move the compose to a different environment.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the compose stack.",
			},
			"app_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The app name used for Docker service naming. Auto-generated if not specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the compose stack.",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Server ID to deploy the compose stack to. If not specified, deploys to the default server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Compose file
			"compose_file_content": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Raw docker-compose.yml content (for source_type 'raw').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compose_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path to the docker-compose.yml file in the repository.",
				Default:     stringdefault.StaticString("./docker-compose.yml"),
			},
			"compose_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The compose type: 'docker-compose' (default) or 'stack' for Docker Swarm.",
				Validators: []validator.String{
					stringvalidator.OneOf("docker-compose", "stack"),
				},
				Default: stringdefault.StaticString("docker-compose"),
			},

			// Source type
			"source_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The source type for the compose stack: github, gitlab, bitbucket, gitea, git, or raw.",
				Validators: []validator.String{
					stringvalidator.OneOf("github", "gitlab", "bitbucket", "gitea", "git", "raw"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Custom Git provider settings (source_type = "git")
			"custom_git_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom Git repository URL (for source_type 'git').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_git_branch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Branch to use for custom Git repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_git_ssh_key_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH key ID for accessing the custom Git repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_git_build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path within the custom Git repository.",
			},
			"enable_submodules": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable Git submodules support.",
				Default:     booldefault.StaticBool(false),
			},

			// GitHub provider settings (source_type = "github")
			"repository": schema.StringAttribute{
				Optional:    true,
				Description: "Repository name for GitHub source (e.g., 'my-repo').",
			},
			"branch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Branch to deploy from (GitHub/GitLab/Bitbucket/Gitea).",
				Default:     stringdefault.StaticString("main"),
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Description: "Repository owner/organization for GitHub source.",
			},
			"github_id": schema.StringAttribute{
				Optional:    true,
				Description: "GitHub App installation ID. Required for GitHub source type.",
			},
			"trigger_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Trigger type for deployments: 'push' (default) or 'tag'.",
				Validators: []validator.String{
					stringvalidator.OneOf("push", "tag"),
				},
				Default: stringdefault.StaticString("push"),
			},

			// GitLab provider settings (source_type = "gitlab")
			"gitlab_id": schema.StringAttribute{
				Optional:    true,
				Description: "GitLab integration ID. Required for GitLab source type.",
			},
			"gitlab_project_id": schema.Int64Attribute{
				Optional:    true,
				Description: "GitLab project ID.",
			},
			"gitlab_repository": schema.StringAttribute{
				Optional:    true,
				Description: "GitLab repository name.",
			},
			"gitlab_owner": schema.StringAttribute{
				Optional:    true,
				Description: "GitLab repository owner/group.",
			},
			"gitlab_branch": schema.StringAttribute{
				Optional:    true,
				Description: "GitLab branch to deploy from.",
			},
			"gitlab_build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path within the GitLab repository.",
			},
			"gitlab_path_namespace": schema.StringAttribute{
				Optional:    true,
				Description: "GitLab path namespace (for nested groups).",
			},

			// Bitbucket provider settings (source_type = "bitbucket")
			"bitbucket_id": schema.StringAttribute{
				Optional:    true,
				Description: "Bitbucket integration ID. Required for Bitbucket source type.",
			},
			"bitbucket_repository": schema.StringAttribute{
				Optional:    true,
				Description: "Bitbucket repository name.",
			},
			"bitbucket_owner": schema.StringAttribute{
				Optional:    true,
				Description: "Bitbucket repository owner/workspace.",
			},
			"bitbucket_branch": schema.StringAttribute{
				Optional:    true,
				Description: "Bitbucket branch to deploy from.",
			},
			"bitbucket_build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path within the Bitbucket repository.",
			},

			// Gitea provider settings (source_type = "gitea")
			"gitea_id": schema.StringAttribute{
				Optional:    true,
				Description: "Gitea integration ID. Required for Gitea source type.",
			},
			"gitea_repository": schema.StringAttribute{
				Optional:    true,
				Description: "Gitea repository name.",
			},
			"gitea_owner": schema.StringAttribute{
				Optional:    true,
				Description: "Gitea repository owner/organization.",
			},
			"gitea_branch": schema.StringAttribute{
				Optional:    true,
				Description: "Gitea branch to deploy from.",
			},
			"gitea_build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path within the Gitea repository.",
			},

			// Environment
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables in KEY=VALUE format, one per line.",
			},

			// Runtime configuration
			"auto_deploy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable automatic deployment on Git push. Defaults to API default (typically true).",
			},

			// Advanced configuration
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run for deployment.",
			},
			"suffix": schema.StringAttribute{
				Optional:    true,
				Description: "Suffix to add to service names.",
			},
			"randomize": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Randomize service names.",
				Default:     booldefault.StaticBool(false),
			},
			"isolated_deployment": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable isolated deployments.",
				Default:     booldefault.StaticBool(false),
			},
			"isolated_deployments_volume": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable isolated deployment volumes.",
				Default:     booldefault.StaticBool(false),
			},
			"watch_paths": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Paths to watch for changes to trigger deployments.",
			},

			// Computed status fields
			"compose_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the compose stack: idle, running, done, or error.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"refresh_token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook refresh token for triggering deployments.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the compose stack was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Deployment options
			"deploy_on_create": schema.BoolAttribute{
				Optional:    true,
				Description: "Trigger a deployment after creating the compose stack.",
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

	// Infer source type if not specified
	if plan.SourceType.IsUnknown() || plan.SourceType.IsNull() {
		plan.SourceType = inferComposeSourceType(&plan)
	}

	// Convert WatchPaths from types.List to []string
	var watchPaths []string
	if !plan.WatchPaths.IsNull() && !plan.WatchPaths.IsUnknown() {
		diags = plan.WatchPaths.ElementsAs(ctx, &watchPaths, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
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
		// Advanced configuration
		ComposeType:               plan.ComposeType.ValueString(),
		Command:                   plan.Command.ValueString(),
		Suffix:                    plan.Suffix.ValueString(),
		Randomize:                 plan.Randomize.ValueBool(),
		IsolatedDeployment:        plan.IsolatedDeployment.ValueBool(),
		IsolatedDeploymentsVolume: plan.IsolatedDeploymentsVolume.ValueBool(),
		WatchPaths:                watchPaths,
	}

	// GitHub fields
	if !plan.Repository.IsNull() {
		comp.Repository = plan.Repository.ValueString()
	}
	if !plan.Branch.IsNull() {
		comp.Branch = plan.Branch.ValueString()
	}
	if !plan.Owner.IsNull() {
		comp.Owner = plan.Owner.ValueString()
	}
	if !plan.GithubId.IsNull() {
		comp.GithubId = plan.GithubId.ValueString()
	}

	// GitLab fields
	if !plan.GitlabId.IsNull() {
		comp.GitlabId = plan.GitlabId.ValueString()
	}
	if !plan.GitlabProjectId.IsNull() {
		comp.GitlabProjectId = plan.GitlabProjectId.ValueInt64()
	}
	if !plan.GitlabRepository.IsNull() {
		comp.GitlabRepository = plan.GitlabRepository.ValueString()
	}
	if !plan.GitlabOwner.IsNull() {
		comp.GitlabOwner = plan.GitlabOwner.ValueString()
	}
	if !plan.GitlabBranch.IsNull() {
		comp.GitlabBranch = plan.GitlabBranch.ValueString()
	}
	if !plan.GitlabBuildPath.IsNull() {
		comp.GitlabBuildPath = plan.GitlabBuildPath.ValueString()
	}
	if !plan.GitlabPathNamespace.IsNull() {
		comp.GitlabPathNamespace = plan.GitlabPathNamespace.ValueString()
	}

	// Bitbucket fields
	if !plan.BitbucketId.IsNull() {
		comp.BitbucketId = plan.BitbucketId.ValueString()
	}
	if !plan.BitbucketRepository.IsNull() {
		comp.BitbucketRepository = plan.BitbucketRepository.ValueString()
	}
	if !plan.BitbucketOwner.IsNull() {
		comp.BitbucketOwner = plan.BitbucketOwner.ValueString()
	}
	if !plan.BitbucketBranch.IsNull() {
		comp.BitbucketBranch = plan.BitbucketBranch.ValueString()
	}
	if !plan.BitbucketBuildPath.IsNull() {
		comp.BitbucketBuildPath = plan.BitbucketBuildPath.ValueString()
	}

	// Gitea fields
	if !plan.GiteaId.IsNull() {
		comp.GiteaId = plan.GiteaId.ValueString()
	}
	if !plan.GiteaRepository.IsNull() {
		comp.GiteaRepository = plan.GiteaRepository.ValueString()
	}
	if !plan.GiteaOwner.IsNull() {
		comp.GiteaOwner = plan.GiteaOwner.ValueString()
	}
	if !plan.GiteaBranch.IsNull() {
		comp.GiteaBranch = plan.GiteaBranch.ValueString()
	}
	if !plan.GiteaBuildPath.IsNull() {
		comp.GiteaBuildPath = plan.GiteaBuildPath.ValueString()
	}

	createdComp, err := r.client.CreateCompose(comp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating compose", err.Error())
		return
	}

	// Update plan from created compose
	plan.ID = types.StringValue(createdComp.ID)
	readComposeIntoState(ctx, &plan, createdComp, &resp.Diagnostics)

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

	readComposeIntoState(ctx, &state, comp, &resp.Diagnostics)

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

	var state ComposeResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentChanged := !plan.EnvironmentID.Equal(state.EnvironmentID)

	// Check if environment_id changed - use compose.move API
	if environmentChanged {
		movedComp, err := r.client.MoveCompose(plan.ID.ValueString(), plan.EnvironmentID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error moving compose to new environment", err.Error())
			return
		}

		// Check if only environment_id changed - if so, skip the update call
		onlyEnvironmentChanged := plan.Name.Equal(state.Name) &&
			plan.ComposeFileContent.Equal(state.ComposeFileContent) &&
			plan.SourceType.Equal(state.SourceType) &&
			plan.CustomGitUrl.Equal(state.CustomGitUrl) &&
			plan.CustomGitBranch.Equal(state.CustomGitBranch) &&
			plan.CustomGitSSHKeyID.Equal(state.CustomGitSSHKeyID) &&
			plan.ComposePath.Equal(state.ComposePath) &&
			plan.AutoDeploy.Equal(state.AutoDeploy) &&
			plan.ComposeType.Equal(state.ComposeType) &&
			plan.Command.Equal(state.Command) &&
			plan.Suffix.Equal(state.Suffix) &&
			plan.Randomize.Equal(state.Randomize) &&
			plan.IsolatedDeployment.Equal(state.IsolatedDeployment) &&
			plan.IsolatedDeploymentsVolume.Equal(state.IsolatedDeploymentsVolume) &&
			plan.WatchPaths.Equal(state.WatchPaths)

		if onlyEnvironmentChanged {
			// MoveCompose is sufficient; use returned data to update state
			readComposeIntoState(ctx, &plan, movedComp, &resp.Diagnostics)
			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// Convert WatchPaths from types.List to []string
	var watchPaths []string
	if !plan.WatchPaths.IsNull() && !plan.WatchPaths.IsUnknown() {
		diags = plan.WatchPaths.ElementsAs(ctx, &watchPaths, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
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
		// Advanced configuration
		ComposeType:               plan.ComposeType.ValueString(),
		Command:                   plan.Command.ValueString(),
		Suffix:                    plan.Suffix.ValueString(),
		Randomize:                 plan.Randomize.ValueBool(),
		IsolatedDeployment:        plan.IsolatedDeployment.ValueBool(),
		IsolatedDeploymentsVolume: plan.IsolatedDeploymentsVolume.ValueBool(),
		WatchPaths:                watchPaths,
	}

	// GitHub fields
	if !plan.Repository.IsNull() {
		comp.Repository = plan.Repository.ValueString()
	}
	if !plan.Branch.IsNull() {
		comp.Branch = plan.Branch.ValueString()
	}
	if !plan.Owner.IsNull() {
		comp.Owner = plan.Owner.ValueString()
	}
	if !plan.GithubId.IsNull() {
		comp.GithubId = plan.GithubId.ValueString()
	}

	// GitLab fields
	if !plan.GitlabId.IsNull() {
		comp.GitlabId = plan.GitlabId.ValueString()
	}
	if !plan.GitlabProjectId.IsNull() {
		comp.GitlabProjectId = plan.GitlabProjectId.ValueInt64()
	}
	if !plan.GitlabRepository.IsNull() {
		comp.GitlabRepository = plan.GitlabRepository.ValueString()
	}
	if !plan.GitlabOwner.IsNull() {
		comp.GitlabOwner = plan.GitlabOwner.ValueString()
	}
	if !plan.GitlabBranch.IsNull() {
		comp.GitlabBranch = plan.GitlabBranch.ValueString()
	}
	if !plan.GitlabBuildPath.IsNull() {
		comp.GitlabBuildPath = plan.GitlabBuildPath.ValueString()
	}

	// Bitbucket fields
	if !plan.BitbucketId.IsNull() {
		comp.BitbucketId = plan.BitbucketId.ValueString()
	}
	if !plan.BitbucketRepository.IsNull() {
		comp.BitbucketRepository = plan.BitbucketRepository.ValueString()
	}
	if !plan.BitbucketOwner.IsNull() {
		comp.BitbucketOwner = plan.BitbucketOwner.ValueString()
	}
	if !plan.BitbucketBranch.IsNull() {
		comp.BitbucketBranch = plan.BitbucketBranch.ValueString()
	}

	// Gitea fields
	if !plan.GiteaId.IsNull() {
		comp.GiteaId = plan.GiteaId.ValueString()
	}
	if !plan.GiteaRepository.IsNull() {
		comp.GiteaRepository = plan.GiteaRepository.ValueString()
	}
	if !plan.GiteaOwner.IsNull() {
		comp.GiteaOwner = plan.GiteaOwner.ValueString()
	}
	if !plan.GiteaBranch.IsNull() {
		comp.GiteaBranch = plan.GiteaBranch.ValueString()
	}

	updatedComp, err := r.client.UpdateCompose(comp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating compose", err.Error())
		return
	}

	readComposeIntoState(ctx, &plan, updatedComp, &resp.Diagnostics)

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
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "not_found") || strings.Contains(errStr, "404") {
			// Resource already deleted, that's fine
			return
		}
		resp.Diagnostics.AddError("Error deleting compose", err.Error())
		return
	}
}

func (r *ComposeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

func inferComposeSourceType(plan *ComposeResourceModel) types.String {
	if !plan.ComposeFileContent.IsNull() && !plan.ComposeFileContent.IsUnknown() && plan.ComposeFileContent.ValueString() != "" {
		return types.StringValue("raw")
	}
	if !plan.CustomGitUrl.IsNull() && !plan.CustomGitUrl.IsUnknown() && plan.CustomGitUrl.ValueString() != "" {
		return types.StringValue("git")
	}
	if !plan.GitlabId.IsNull() && !plan.GitlabId.IsUnknown() && plan.GitlabId.ValueString() != "" {
		return types.StringValue("gitlab")
	}
	if !plan.BitbucketId.IsNull() && !plan.BitbucketId.IsUnknown() && plan.BitbucketId.ValueString() != "" {
		return types.StringValue("bitbucket")
	}
	if !plan.GiteaId.IsNull() && !plan.GiteaId.IsUnknown() && plan.GiteaId.ValueString() != "" {
		return types.StringValue("gitea")
	}
	return types.StringValue("github")
}

func readComposeIntoState(ctx context.Context, state *ComposeResourceModel, comp *client.Compose, diags *diag.Diagnostics) {
	state.Name = types.StringValue(comp.Name)

	if comp.EnvironmentID != "" {
		state.EnvironmentID = types.StringValue(comp.EnvironmentID)
	}
	if comp.AppName != "" {
		state.AppName = types.StringValue(comp.AppName)
	}
	if comp.Description != "" {
		state.Description = types.StringValue(comp.Description)
	}
	if comp.ServerID != "" {
		state.ServerID = types.StringValue(comp.ServerID)
	} else if state.ServerID.IsUnknown() {
		state.ServerID = types.StringNull()
	}

	// Compose file
	if comp.ComposeFile != "" {
		state.ComposeFileContent = types.StringValue(comp.ComposeFile)
	} else if state.ComposeFileContent.IsUnknown() {
		state.ComposeFileContent = types.StringNull()
	}
	if comp.ComposePath != "" {
		state.ComposePath = types.StringValue(comp.ComposePath)
	}
	if comp.ComposeType != "" {
		state.ComposeType = types.StringValue(comp.ComposeType)
	}

	// Source type
	if comp.SourceType != "" {
		state.SourceType = types.StringValue(comp.SourceType)
	}

	// Custom Git fields
	if comp.CustomGitUrl != "" {
		state.CustomGitUrl = types.StringValue(comp.CustomGitUrl)
	} else if state.CustomGitUrl.IsUnknown() {
		state.CustomGitUrl = types.StringNull()
	}
	if comp.CustomGitBranch != "" {
		state.CustomGitBranch = types.StringValue(comp.CustomGitBranch)
	} else if state.CustomGitBranch.IsUnknown() {
		state.CustomGitBranch = types.StringNull()
	}
	if comp.CustomGitSSHKeyId != "" {
		state.CustomGitSSHKeyID = types.StringValue(comp.CustomGitSSHKeyId)
	} else if state.CustomGitSSHKeyID.IsUnknown() {
		state.CustomGitSSHKeyID = types.StringNull()
	}
	if comp.CustomGitBuildPath != "" {
		state.CustomGitBuildPath = types.StringValue(comp.CustomGitBuildPath)
	}
	state.EnableSubmodules = types.BoolValue(comp.EnableSubmodules)

	// GitHub fields
	if comp.Repository != "" {
		state.Repository = types.StringValue(comp.Repository)
	}
	if comp.Branch != "" {
		state.Branch = types.StringValue(comp.Branch)
	}
	if comp.Owner != "" {
		state.Owner = types.StringValue(comp.Owner)
	}
	if comp.GithubId != "" {
		state.GithubId = types.StringValue(comp.GithubId)
	}
	if comp.TriggerType != "" {
		state.TriggerType = types.StringValue(comp.TriggerType)
	}

	// GitLab fields
	if comp.GitlabId != "" {
		state.GitlabId = types.StringValue(comp.GitlabId)
	}
	if comp.GitlabProjectId != 0 {
		state.GitlabProjectId = types.Int64Value(comp.GitlabProjectId)
	}
	if comp.GitlabRepository != "" {
		state.GitlabRepository = types.StringValue(comp.GitlabRepository)
	}
	if comp.GitlabOwner != "" {
		state.GitlabOwner = types.StringValue(comp.GitlabOwner)
	}
	if comp.GitlabBranch != "" {
		state.GitlabBranch = types.StringValue(comp.GitlabBranch)
	}
	if comp.GitlabBuildPath != "" {
		state.GitlabBuildPath = types.StringValue(comp.GitlabBuildPath)
	}
	if comp.GitlabPathNamespace != "" {
		state.GitlabPathNamespace = types.StringValue(comp.GitlabPathNamespace)
	}

	// Bitbucket fields
	if comp.BitbucketId != "" {
		state.BitbucketId = types.StringValue(comp.BitbucketId)
	}
	if comp.BitbucketRepository != "" {
		state.BitbucketRepository = types.StringValue(comp.BitbucketRepository)
	}
	if comp.BitbucketOwner != "" {
		state.BitbucketOwner = types.StringValue(comp.BitbucketOwner)
	}
	if comp.BitbucketBranch != "" {
		state.BitbucketBranch = types.StringValue(comp.BitbucketBranch)
	}
	if comp.BitbucketBuildPath != "" {
		state.BitbucketBuildPath = types.StringValue(comp.BitbucketBuildPath)
	}

	// Gitea fields
	if comp.GiteaId != "" {
		state.GiteaId = types.StringValue(comp.GiteaId)
	}
	if comp.GiteaRepository != "" {
		state.GiteaRepository = types.StringValue(comp.GiteaRepository)
	}
	if comp.GiteaOwner != "" {
		state.GiteaOwner = types.StringValue(comp.GiteaOwner)
	}
	if comp.GiteaBranch != "" {
		state.GiteaBranch = types.StringValue(comp.GiteaBranch)
	}
	if comp.GiteaBuildPath != "" {
		state.GiteaBuildPath = types.StringValue(comp.GiteaBuildPath)
	}

	// Environment
	if comp.Env != "" {
		state.Env = types.StringValue(comp.Env)
	}

	// Runtime
	state.AutoDeploy = types.BoolValue(comp.AutoDeploy)

	// Advanced configuration
	if comp.Command != "" {
		state.Command = types.StringValue(comp.Command)
	}
	if comp.Suffix != "" {
		state.Suffix = types.StringValue(comp.Suffix)
	}
	state.Randomize = types.BoolValue(comp.Randomize)
	state.IsolatedDeployment = types.BoolValue(comp.IsolatedDeployment)
	state.IsolatedDeploymentsVolume = types.BoolValue(comp.IsolatedDeploymentsVolume)

	// WatchPaths - convert []string to types.List
	if len(comp.WatchPaths) > 0 {
		watchPathsList, d := types.ListValueFrom(ctx, types.StringType, comp.WatchPaths)
		diags.Append(d...)
		state.WatchPaths = watchPathsList
	} else {
		// Set to null when API returns empty array to handle drift properly
		state.WatchPaths = types.ListNull(types.StringType)
	}

	// Computed status fields
	if comp.ComposeStatus != "" {
		state.ComposeStatus = types.StringValue(comp.ComposeStatus)
	} else if state.ComposeStatus.IsUnknown() {
		state.ComposeStatus = types.StringNull()
	}
	if comp.RefreshToken != "" {
		state.RefreshToken = types.StringValue(comp.RefreshToken)
	} else if state.RefreshToken.IsUnknown() {
		state.RefreshToken = types.StringNull()
	}
	if comp.CreatedAt != "" {
		state.CreatedAt = types.StringValue(comp.CreatedAt)
	} else if state.CreatedAt.IsUnknown() {
		state.CreatedAt = types.StringNull()
	}
}
