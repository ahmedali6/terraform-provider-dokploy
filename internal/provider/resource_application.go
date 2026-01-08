package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

type ApplicationResource struct {
	client *client.DokployClient
}

type ApplicationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	ServerID      types.String `tfsdk:"server_id"`

	// Source type
	SourceType types.String `tfsdk:"source_type"`

	// Git provider settings (for source_type = "git")
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	CustomGitBuildPath types.String `tfsdk:"custom_git_build_path"`
	EnableSubmodules   types.Bool   `tfsdk:"enable_submodules"`
	WatchPaths         types.List   `tfsdk:"watch_paths"`
	CleanCache         types.Bool   `tfsdk:"clean_cache"`

	// GitHub provider settings (for source_type = "github")
	Repository  types.String `tfsdk:"repository"`
	Branch      types.String `tfsdk:"branch"`
	Owner       types.String `tfsdk:"owner"`
	BuildPath   types.String `tfsdk:"build_path"`
	GithubId    types.String `tfsdk:"github_id"`
	TriggerType types.String `tfsdk:"trigger_type"`

	// GitLab provider settings (for source_type = "gitlab")
	GitlabId            types.String `tfsdk:"gitlab_id"`
	GitlabProjectId     types.Int64  `tfsdk:"gitlab_project_id"`
	GitlabRepository    types.String `tfsdk:"gitlab_repository"`
	GitlabOwner         types.String `tfsdk:"gitlab_owner"`
	GitlabBranch        types.String `tfsdk:"gitlab_branch"`
	GitlabBuildPath     types.String `tfsdk:"gitlab_build_path"`
	GitlabPathNamespace types.String `tfsdk:"gitlab_path_namespace"`

	// Bitbucket provider settings (for source_type = "bitbucket")
	BitbucketId         types.String `tfsdk:"bitbucket_id"`
	BitbucketRepository types.String `tfsdk:"bitbucket_repository"`
	BitbucketOwner      types.String `tfsdk:"bitbucket_owner"`
	BitbucketBranch     types.String `tfsdk:"bitbucket_branch"`
	BitbucketBuildPath  types.String `tfsdk:"bitbucket_build_path"`

	// Gitea provider settings (for source_type = "gitea")
	GiteaId         types.String `tfsdk:"gitea_id"`
	GiteaRepository types.String `tfsdk:"gitea_repository"`
	GiteaOwner      types.String `tfsdk:"gitea_owner"`
	GiteaBranch     types.String `tfsdk:"gitea_branch"`
	GiteaBuildPath  types.String `tfsdk:"gitea_build_path"`

	// Docker provider settings (for source_type = "docker")
	DockerImage types.String `tfsdk:"docker_image"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	RegistryUrl types.String `tfsdk:"registry_url"`
	RegistryId  types.String `tfsdk:"registry_id"`

	// Build type settings
	BuildType         types.String `tfsdk:"build_type"`
	DockerfilePath    types.String `tfsdk:"dockerfile_path"`
	DockerContextPath types.String `tfsdk:"docker_context_path"`
	DockerBuildStage  types.String `tfsdk:"docker_build_stage"`
	PublishDirectory  types.String `tfsdk:"publish_directory"`
	Dockerfile        types.String `tfsdk:"dockerfile"`
	DropBuildPath     types.String `tfsdk:"drop_build_path"`
	HerokuVersion     types.String `tfsdk:"heroku_version"`
	RailpackVersion   types.String `tfsdk:"railpack_version"`
	IsStaticSpa       types.Bool   `tfsdk:"is_static_spa"`

	// Environment settings
	Env           types.String `tfsdk:"env"`
	BuildArgs     types.String `tfsdk:"build_args"`
	BuildSecrets  types.String `tfsdk:"build_secrets"`
	CreateEnvFile types.Bool   `tfsdk:"create_env_file"`

	// Runtime configuration
	AutoDeploy        types.Bool   `tfsdk:"auto_deploy"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	MemoryLimit       types.Int64  `tfsdk:"memory_limit"`
	MemoryReservation types.Int64  `tfsdk:"memory_reservation"`
	CpuLimit          types.Int64  `tfsdk:"cpu_limit"`
	CpuReservation    types.Int64  `tfsdk:"cpu_reservation"`
	Command           types.String `tfsdk:"command"`
	Args              types.String `tfsdk:"args"`

	// Preview deployments
	IsPreviewDeploymentsActive            types.Bool   `tfsdk:"preview_deployments_enabled"`
	PreviewEnv                            types.String `tfsdk:"preview_env"`
	PreviewBuildArgs                      types.String `tfsdk:"preview_build_args"`
	PreviewBuildSecrets                   types.String `tfsdk:"preview_build_secrets"`
	PreviewLabels                         types.List   `tfsdk:"preview_labels"`
	PreviewWildcard                       types.String `tfsdk:"preview_wildcard"`
	PreviewPort                           types.Int64  `tfsdk:"preview_port"`
	PreviewHttps                          types.Bool   `tfsdk:"preview_https"`
	PreviewPath                           types.String `tfsdk:"preview_path"`
	PreviewCertificateType                types.String `tfsdk:"preview_certificate_type"`
	PreviewCustomCertResolver             types.String `tfsdk:"preview_custom_cert_resolver"`
	PreviewLimit                          types.Int64  `tfsdk:"preview_limit"`
	PreviewRequireCollaboratorPermissions types.Bool   `tfsdk:"preview_require_collaborator_permissions"`

	// Rollback configuration
	RollbackActive     types.Bool   `tfsdk:"rollback_active"`
	RollbackRegistryId types.String `tfsdk:"rollback_registry_id"`

	// Build server configuration
	BuildServerId   types.String `tfsdk:"build_server_id"`
	BuildRegistryId types.String `tfsdk:"build_registry_id"`

	// Display settings
	Title    types.String `tfsdk:"title"`
	Subtitle types.String `tfsdk:"subtitle"`
	Enabled  types.Bool   `tfsdk:"enabled"`

	// Deployment options
	DeployOnCreate types.Bool `tfsdk:"deploy_on_create"`

	// Application status (computed)
	ApplicationStatus types.String `tfsdk:"application_status"`

	// Docker Swarm configuration (stored as JSON strings)
	HealthCheckSwarm     types.String `tfsdk:"health_check_swarm"`
	RestartPolicySwarm   types.String `tfsdk:"restart_policy_swarm"`
	PlacementSwarm       types.String `tfsdk:"placement_swarm"`
	UpdateConfigSwarm    types.String `tfsdk:"update_config_swarm"`
	RollbackConfigSwarm  types.String `tfsdk:"rollback_config_swarm"`
	ModeSwarm            types.String `tfsdk:"mode_swarm"`
	LabelsSwarm          types.String `tfsdk:"labels_swarm"`
	NetworkSwarm         types.String `tfsdk:"network_swarm"`
	StopGracePeriodSwarm types.Int64  `tfsdk:"stop_grace_period_swarm"`
	EndpointSpecSwarm    types.String `tfsdk:"endpoint_spec_swarm"`
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy application. Supports multiple source types including GitHub, GitLab, Bitbucket, Gitea, custom Git repositories, and Docker images.",
		Attributes: map[string]schema.Attribute{
			// Core attributes
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The environment ID this application belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the application.",
			},
			"app_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The app name used for Docker container naming. Auto-generated if not specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the application.",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID to deploy the application to. If not specified, deploys to the default server.",
			},

			// Source type
			"source_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The source type for the application: github, gitlab, bitbucket, gitea, git, docker, or drop.",
				Validators: []validator.String{
					stringvalidator.OneOf("github", "gitlab", "bitbucket", "gitea", "git", "docker", "drop"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Custom Git provider settings (source_type = "git")
			"custom_git_url": schema.StringAttribute{
				Optional:    true,
				Description: "Custom Git repository URL (for source_type 'git').",
			},
			"custom_git_branch": schema.StringAttribute{
				Optional:    true,
				Description: "Branch to use for custom Git repository.",
			},
			"custom_git_ssh_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "SSH key ID for accessing the custom Git repository.",
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
			"clean_cache": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Clean cache before building.",
				Default:     booldefault.StaticBool(false),
			},
			"watch_paths": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Paths to watch for changes to trigger deployments.",
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
			"build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path within the repository for GitHub source.",
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

			// Docker provider settings (source_type = "docker")
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Description: "Docker image to use (for source_type 'docker'). Example: 'nginx:alpine'.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for Docker registry authentication.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for Docker registry authentication.",
			},
			"registry_url": schema.StringAttribute{
				Optional:    true,
				Description: "Docker registry URL. Leave empty for Docker Hub.",
			},
			"registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "Registry ID from Dokploy registry management.",
			},

			// Build type settings
			"build_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Build type: dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, or railpack.",
				Validators: []validator.String{
					stringvalidator.OneOf("dockerfile", "heroku_buildpacks", "paketo_buildpacks", "nixpacks", "static", "railpack"),
				},
				Default: stringdefault.StaticString("nixpacks"),
			},
			"dockerfile_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path to the Dockerfile (relative to build path).",
				Default:     stringdefault.StaticString("./Dockerfile"),
			},
			"docker_context_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker build context path.",
				Default:     stringdefault.StaticString("."),
			},
			"docker_build_stage": schema.StringAttribute{
				Optional:    true,
				Description: "Target stage for multi-stage Docker builds.",
			},
			"publish_directory": schema.StringAttribute{
				Optional:    true,
				Description: "Publish directory for static builds.",
			},
			"dockerfile": schema.StringAttribute{
				Optional:    true,
				Description: "Raw Dockerfile content (for 'drop' source type or inline Dockerfile).",
			},
			"drop_build_path": schema.StringAttribute{
				Optional:    true,
				Description: "Build path for 'drop' source type deployments.",
			},
			"heroku_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Heroku buildpack version (for heroku_buildpacks build type).",
			},
			"railpack_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Railpack version (for railpack build type).",
			},
			"is_static_spa": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the static build is a Single Page Application.",
			},

			// Environment settings
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables in KEY=VALUE format, one per line.",
			},
			"build_args": schema.StringAttribute{
				Optional:    true,
				Description: "Build arguments in KEY=VALUE format, one per line.",
			},
			"build_secrets": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Build secrets in KEY=VALUE format, one per line.",
			},
			"create_env_file": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Create a .env file in the container.",
			},

			// Runtime configuration
			"auto_deploy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable automatic deployment on Git push.",
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of container replicas to run.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"memory_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Memory limit in bytes. Example: 536870912 (512MB).",
			},
			"memory_reservation": schema.Int64Attribute{
				Optional:    true,
				Description: "Memory reservation (soft limit) in bytes.",
			},
			"cpu_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "CPU limit in nanocores. Example: 1000000000 (1 CPU).",
			},
			"cpu_reservation": schema.Int64Attribute{
				Optional:    true,
				Description: "CPU reservation in nanocores.",
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run (overrides Dockerfile CMD).",
			},
			"args": schema.StringAttribute{
				Optional:    true,
				Description: "Arguments to pass to the command.",
			},

			// Preview deployments
			"preview_deployments_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable preview deployments for pull requests.",
			},
			"preview_env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables for preview deployments.",
			},
			"preview_build_args": schema.StringAttribute{
				Optional:    true,
				Description: "Build arguments for preview deployments.",
			},
			"preview_build_secrets": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Build secrets for preview deployments in KEY=VALUE format.",
			},
			"preview_labels": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Labels for preview deployments.",
			},
			"preview_wildcard": schema.StringAttribute{
				Optional:    true,
				Description: "Wildcard domain for preview deployments (e.g., '*.preview.example.com').",
			},
			"preview_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Port for preview deployment containers.",
			},
			"preview_https": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable HTTPS for preview deployments.",
			},
			"preview_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path prefix for preview deployment URLs.",
			},
			"preview_certificate_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Certificate type for preview deployments: letsencrypt, none.",
				Validators: []validator.String{
					stringvalidator.OneOf("letsencrypt", "none"),
				},
			},
			"preview_custom_cert_resolver": schema.StringAttribute{
				Optional:    true,
				Description: "Custom certificate resolver for preview deployments.",
			},
			"preview_limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of concurrent preview deployments.",
			},
			"preview_require_collaborator_permissions": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Require collaborator permissions to create preview deployments.",
			},

			// Rollback configuration
			"rollback_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable rollback capability.",
			},
			"rollback_registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "Registry ID to use for rollback images.",
			},

			// Build server configuration
			"build_server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Build server ID for remote builds.",
			},
			"build_registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "Registry ID to push build images to.",
			},

			// Display settings
			"title": schema.StringAttribute{
				Optional:    true,
				Description: "Display title for the application in the UI.",
			},
			"subtitle": schema.StringAttribute{
				Optional:    true,
				Description: "Display subtitle for the application in the UI.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the application is enabled.",
			},

			// Deployment options
			"deploy_on_create": schema.BoolAttribute{
				Optional:    true,
				Description: "Trigger a deployment after creating the application.",
			},

			// Application status (computed)
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the application: idle, running, done, error.",
			},

			// Docker Swarm configuration
			"health_check_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Health check configuration for Docker Swarm mode (JSON format).",
			},
			"restart_policy_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Restart policy configuration for Docker Swarm mode (JSON format).",
			},
			"placement_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Placement constraints for Docker Swarm mode (JSON format).",
			},
			"update_config_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Update configuration for Docker Swarm mode (JSON format).",
			},
			"rollback_config_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Rollback configuration for Docker Swarm mode (JSON format).",
			},
			"mode_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Service mode for Docker Swarm: replicated or global (JSON format).",
			},
			"labels_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Labels for Docker Swarm service (JSON format).",
			},
			"network_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Network configuration for Docker Swarm mode (JSON array format).",
			},
			"stop_grace_period_swarm": schema.Int64Attribute{
				Optional:    true,
				Description: "Stop grace period in nanoseconds for Docker Swarm mode.",
			},
			"endpoint_spec_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Endpoint specification for Docker Swarm mode (JSON format).",
			},
		},
	}
}

func (r *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Infer source type if not specified
	if plan.SourceType.IsUnknown() || plan.SourceType.IsNull() {
		plan.SourceType = inferSourceType(&plan)
	}

	// 1. Create application with minimal required fields
	app := client.Application{
		Name:          plan.Name.ValueString(),
		AppName:       plan.AppName.ValueString(),
		Description:   plan.Description.ValueString(),
		EnvironmentID: plan.EnvironmentID.ValueString(),
		ServerID:      plan.ServerID.ValueString(),
	}

	createdApp, err := r.client.CreateApplication(app)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application", err.Error())
		return
	}

	plan.ID = types.StringValue(createdApp.ID)
	if createdApp.AppName != "" {
		plan.AppName = types.StringValue(createdApp.AppName)
	}

	// 2. Update general settings (sourceType, autoDeploy, replicas, etc.)
	if err := r.updateGeneralSettings(createdApp.ID, &plan); err != nil {
		resp.Diagnostics.AddError("Error updating application general settings", err.Error())
		return
	}

	// 3. Save build type settings if applicable (non-docker source types)
	if plan.SourceType.ValueString() != "docker" {
		if err := r.saveBuildType(createdApp.ID, &plan); err != nil {
			resp.Diagnostics.AddError("Error saving build type", err.Error())
			return
		}
	}

	// 4. Configure source provider based on source_type
	if err := r.saveSourceProvider(createdApp.ID, &plan); err != nil {
		resp.Diagnostics.AddError("Error saving source provider", err.Error())
		return
	}

	// 5. Save environment variables if provided
	if err := r.saveEnvironment(createdApp.ID, &plan); err != nil {
		resp.Diagnostics.AddError("Error saving environment", err.Error())
		return
	}

	// 6. Read back the final state
	finalApp, err := r.client.GetApplication(createdApp.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading application after create", err.Error())
		return
	}

	// Update plan with values from the API
	updatePlanFromApplication(&plan, finalApp)

	// 7. Deploy if requested
	if !plan.DeployOnCreate.IsNull() && plan.DeployOnCreate.ValueBool() {
		err := r.client.DeployApplication(createdApp.ID, plan.ServerID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning("Deployment Trigger Failed", fmt.Sprintf("Application created but deployment failed to trigger: %s", err.Error()))
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	// Update state with values from API
	readApplicationIntoState(&state, app)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	var state ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ID.ValueString()
	plan.ID = state.ID

	// 1. Update general settings
	if err := r.updateGeneralSettings(appID, &plan); err != nil {
		resp.Diagnostics.AddError("Error updating application general settings", err.Error())
		return
	}

	// 2. Update build type if changed (for non-docker source types)
	sourceType := plan.SourceType.ValueString()
	if sourceType != "docker" {
		if err := r.saveBuildType(appID, &plan); err != nil {
			resp.Diagnostics.AddError("Error saving build type", err.Error())
			return
		}
	}

	// 3. Update source provider settings based on source_type
	if err := r.saveSourceProvider(appID, &plan); err != nil {
		resp.Diagnostics.AddError("Error saving source provider", err.Error())
		return
	}

	// 4. Update environment if changed
	if err := r.saveEnvironment(appID, &plan); err != nil {
		resp.Diagnostics.AddError("Error saving environment", err.Error())
		return
	}

	// 5. Read back the final state
	finalApp, err := r.client.GetApplication(appID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading application after update", err.Error())
		return
	}

	// Update plan with values from the API
	updatePlanFromApplication(&plan, finalApp)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(state.ID.ValueString())
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "not_found") || strings.Contains(errStr, "404") {
			// Resource already deleted, that's fine
			return
		}
		resp.Diagnostics.AddError("Error deleting application", err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

func inferSourceType(plan *ApplicationResourceModel) types.String {
	if !plan.DockerImage.IsNull() && !plan.DockerImage.IsUnknown() && plan.DockerImage.ValueString() != "" {
		return types.StringValue("docker")
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

func (r *ApplicationResource) updateGeneralSettings(appID string, plan *ApplicationResourceModel) error {
	generalApp := client.Application{
		ID:         appID,
		Name:       plan.Name.ValueString(),
		AppName:    plan.AppName.ValueString(),
		SourceType: plan.SourceType.ValueString(),
		AutoDeploy: plan.AutoDeploy.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		generalApp.Description = plan.Description.ValueString()
	}
	if !plan.Replicas.IsNull() && !plan.Replicas.IsUnknown() {
		generalApp.Replicas = int(plan.Replicas.ValueInt64())
	}
	if !plan.MemoryLimit.IsNull() && !plan.MemoryLimit.IsUnknown() {
		val := plan.MemoryLimit.ValueInt64()
		generalApp.MemoryLimit = &val
	}
	if !plan.MemoryReservation.IsNull() && !plan.MemoryReservation.IsUnknown() {
		val := plan.MemoryReservation.ValueInt64()
		generalApp.MemoryReservation = &val
	}
	if !plan.CpuLimit.IsNull() && !plan.CpuLimit.IsUnknown() {
		val := plan.CpuLimit.ValueInt64()
		generalApp.CpuLimit = &val
	}
	if !plan.CpuReservation.IsNull() && !plan.CpuReservation.IsUnknown() {
		val := plan.CpuReservation.ValueInt64()
		generalApp.CpuReservation = &val
	}
	if !plan.Command.IsNull() && !plan.Command.IsUnknown() {
		generalApp.Command = plan.Command.ValueString()
	}
	if !plan.Args.IsNull() && !plan.Args.IsUnknown() {
		generalApp.Args = plan.Args.ValueString()
	}

	// Preview deployments
	generalApp.IsPreviewDeploymentsActive = plan.IsPreviewDeploymentsActive.ValueBool()
	if !plan.PreviewEnv.IsNull() && !plan.PreviewEnv.IsUnknown() {
		generalApp.PreviewEnv = plan.PreviewEnv.ValueString()
	}
	if !plan.PreviewBuildArgs.IsNull() && !plan.PreviewBuildArgs.IsUnknown() {
		generalApp.PreviewBuildArgs = plan.PreviewBuildArgs.ValueString()
	}
	if !plan.PreviewWildcard.IsNull() && !plan.PreviewWildcard.IsUnknown() {
		generalApp.PreviewWildcard = plan.PreviewWildcard.ValueString()
	}
	if !plan.PreviewPort.IsNull() && !plan.PreviewPort.IsUnknown() {
		generalApp.PreviewPort = plan.PreviewPort.ValueInt64()
	}
	generalApp.PreviewHttps = plan.PreviewHttps.ValueBool()
	if !plan.PreviewPath.IsNull() && !plan.PreviewPath.IsUnknown() {
		generalApp.PreviewPath = plan.PreviewPath.ValueString()
	}
	if !plan.PreviewCertificateType.IsNull() && !plan.PreviewCertificateType.IsUnknown() {
		generalApp.PreviewCertificateType = plan.PreviewCertificateType.ValueString()
	}
	if !plan.PreviewCustomCertResolver.IsNull() && !plan.PreviewCustomCertResolver.IsUnknown() {
		generalApp.PreviewCustomCertResolver = plan.PreviewCustomCertResolver.ValueString()
	}
	if !plan.PreviewLimit.IsNull() && !plan.PreviewLimit.IsUnknown() {
		generalApp.PreviewLimit = plan.PreviewLimit.ValueInt64()
	}
	generalApp.PreviewRequireCollaboratorPermissions = plan.PreviewRequireCollaboratorPermissions.ValueBool()
	if !plan.PreviewBuildSecrets.IsNull() && !plan.PreviewBuildSecrets.IsUnknown() {
		generalApp.PreviewBuildSecrets = plan.PreviewBuildSecrets.ValueString()
	}

	// Rollback
	generalApp.RollbackActive = plan.RollbackActive.ValueBool()
	if !plan.RollbackRegistryId.IsNull() && !plan.RollbackRegistryId.IsUnknown() {
		generalApp.RollbackRegistryId = plan.RollbackRegistryId.ValueString()
	}

	// Build server
	if !plan.BuildServerId.IsNull() && !plan.BuildServerId.IsUnknown() {
		generalApp.BuildServerId = plan.BuildServerId.ValueString()
	}
	if !plan.BuildRegistryId.IsNull() && !plan.BuildRegistryId.IsUnknown() {
		generalApp.BuildRegistryId = plan.BuildRegistryId.ValueString()
	}

	// Display
	if !plan.Title.IsNull() && !plan.Title.IsUnknown() {
		generalApp.Title = plan.Title.ValueString()
	}
	if !plan.Subtitle.IsNull() && !plan.Subtitle.IsUnknown() {
		generalApp.Subtitle = plan.Subtitle.ValueString()
	}
	generalApp.Enabled = plan.Enabled.ValueBool()

	// Docker Swarm fields - parse JSON strings to maps
	if !plan.HealthCheckSwarm.IsNull() && !plan.HealthCheckSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.HealthCheckSwarm.ValueString()), &m); err == nil {
			generalApp.HealthCheckSwarm = m
		}
	}
	if !plan.RestartPolicySwarm.IsNull() && !plan.RestartPolicySwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.RestartPolicySwarm.ValueString()), &m); err == nil {
			generalApp.RestartPolicySwarm = m
		}
	}
	if !plan.PlacementSwarm.IsNull() && !plan.PlacementSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.PlacementSwarm.ValueString()), &m); err == nil {
			generalApp.PlacementSwarm = m
		}
	}
	if !plan.UpdateConfigSwarm.IsNull() && !plan.UpdateConfigSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.UpdateConfigSwarm.ValueString()), &m); err == nil {
			generalApp.UpdateConfigSwarm = m
		}
	}
	if !plan.RollbackConfigSwarm.IsNull() && !plan.RollbackConfigSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.RollbackConfigSwarm.ValueString()), &m); err == nil {
			generalApp.RollbackConfigSwarm = m
		}
	}
	if !plan.ModeSwarm.IsNull() && !plan.ModeSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.ModeSwarm.ValueString()), &m); err == nil {
			generalApp.ModeSwarm = m
		}
	}
	if !plan.LabelsSwarm.IsNull() && !plan.LabelsSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.LabelsSwarm.ValueString()), &m); err == nil {
			generalApp.LabelsSwarm = m
		}
	}
	if !plan.NetworkSwarm.IsNull() && !plan.NetworkSwarm.IsUnknown() {
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(plan.NetworkSwarm.ValueString()), &arr); err == nil {
			generalApp.NetworkSwarm = arr
		}
	}
	if !plan.StopGracePeriodSwarm.IsNull() && !plan.StopGracePeriodSwarm.IsUnknown() {
		val := plan.StopGracePeriodSwarm.ValueInt64()
		generalApp.StopGracePeriodSwarm = &val
	}
	if !plan.EndpointSpecSwarm.IsNull() && !plan.EndpointSpecSwarm.IsUnknown() {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(plan.EndpointSpecSwarm.ValueString()), &m); err == nil {
			generalApp.EndpointSpecSwarm = m
		}
	}

	_, err := r.client.UpdateApplicationGeneral(generalApp)
	return err
}

func (r *ApplicationResource) saveBuildType(appID string, plan *ApplicationResourceModel) error {
	return r.client.SaveBuildType(
		appID,
		plan.BuildType.ValueString(),
		plan.DockerfilePath.ValueString(),
		plan.DockerContextPath.ValueString(),
		plan.DockerBuildStage.ValueString(),
		plan.PublishDirectory.ValueString(),
	)
}

func (r *ApplicationResource) saveSourceProvider(appID string, plan *ApplicationResourceModel) error {
	sourceType := plan.SourceType.ValueString()

	switch sourceType {
	case "github":
		input := client.SaveGithubProviderInput{
			ApplicationID:    appID,
			Repository:       plan.Repository.ValueString(),
			Branch:           plan.Branch.ValueString(),
			Owner:            plan.Owner.ValueString(),
			BuildPath:        plan.BuildPath.ValueString(),
			GithubId:         plan.GithubId.ValueString(),
			EnableSubmodules: plan.EnableSubmodules.ValueBool(),
			TriggerType:      plan.TriggerType.ValueString(),
		}
		return r.client.SaveGithubProvider(input)

	case "gitlab":
		input := client.SaveGitlabProviderInput{
			ApplicationID:       appID,
			GitlabId:            plan.GitlabId.ValueString(),
			GitlabProjectId:     plan.GitlabProjectId.ValueInt64(),
			GitlabRepository:    plan.GitlabRepository.ValueString(),
			GitlabOwner:         plan.GitlabOwner.ValueString(),
			GitlabBranch:        plan.GitlabBranch.ValueString(),
			GitlabBuildPath:     plan.GitlabBuildPath.ValueString(),
			GitlabPathNamespace: plan.GitlabPathNamespace.ValueString(),
			EnableSubmodules:    plan.EnableSubmodules.ValueBool(),
		}
		return r.client.SaveGitlabProvider(input)

	case "bitbucket":
		input := client.SaveBitbucketProviderInput{
			ApplicationID:       appID,
			BitbucketId:         plan.BitbucketId.ValueString(),
			BitbucketRepository: plan.BitbucketRepository.ValueString(),
			BitbucketOwner:      plan.BitbucketOwner.ValueString(),
			BitbucketBranch:     plan.BitbucketBranch.ValueString(),
			BitbucketBuildPath:  plan.BitbucketBuildPath.ValueString(),
			EnableSubmodules:    plan.EnableSubmodules.ValueBool(),
		}
		return r.client.SaveBitbucketProvider(input)

	case "gitea":
		input := client.SaveGiteaProviderInput{
			ApplicationID:    appID,
			GiteaId:          plan.GiteaId.ValueString(),
			GiteaRepository:  plan.GiteaRepository.ValueString(),
			GiteaOwner:       plan.GiteaOwner.ValueString(),
			GiteaBranch:      plan.GiteaBranch.ValueString(),
			GiteaBuildPath:   plan.GiteaBuildPath.ValueString(),
			EnableSubmodules: plan.EnableSubmodules.ValueBool(),
		}
		return r.client.SaveGiteaProvider(input)

	case "git":
		input := client.SaveGitProviderInput{
			ApplicationID:      appID,
			CustomGitUrl:       plan.CustomGitUrl.ValueString(),
			CustomGitBranch:    plan.CustomGitBranch.ValueString(),
			CustomGitBuildPath: plan.CustomGitBuildPath.ValueString(),
			CustomGitSSHKeyId:  plan.CustomGitSSHKeyID.ValueString(),
			EnableSubmodules:   plan.EnableSubmodules.ValueBool(),
		}
		return r.client.SaveGitProvider(input)

	case "docker":
		input := client.SaveDockerProviderInput{
			ApplicationID: appID,
			DockerImage:   plan.DockerImage.ValueString(),
			Username:      plan.Username.ValueString(),
			Password:      plan.Password.ValueString(),
			RegistryUrl:   plan.RegistryUrl.ValueString(),
			RegistryId:    plan.RegistryId.ValueString(),
		}
		return r.client.SaveDockerProvider(input)
	}

	return nil
}

func (r *ApplicationResource) saveEnvironment(appID string, plan *ApplicationResourceModel) error {
	// Only save if at least one env field is set or create_env_file is explicitly configured
	if (plan.Env.IsNull() || plan.Env.IsUnknown()) &&
		(plan.BuildArgs.IsNull() || plan.BuildArgs.IsUnknown()) &&
		(plan.BuildSecrets.IsNull() || plan.BuildSecrets.IsUnknown()) &&
		(plan.CreateEnvFile.IsNull() || plan.CreateEnvFile.IsUnknown()) {
		return nil
	}

	createEnvFile := plan.CreateEnvFile.ValueBool()
	input := client.SaveEnvironmentInput{
		ApplicationID: appID,
		Env:           plan.Env.ValueString(),
		BuildArgs:     plan.BuildArgs.ValueString(),
		BuildSecrets:  plan.BuildSecrets.ValueString(),
		CreateEnvFile: &createEnvFile,
	}
	return r.client.SaveEnvironment(input)
}

func updatePlanFromApplication(plan *ApplicationResourceModel, app *client.Application) {
	if app.AppName != "" {
		plan.AppName = types.StringValue(app.AppName)
	}
	if app.SourceType != "" {
		plan.SourceType = types.StringValue(app.SourceType)
	}

	// Update computed fields
	plan.AutoDeploy = types.BoolValue(app.AutoDeploy)
	plan.EnableSubmodules = types.BoolValue(app.EnableSubmodules)

	if app.Replicas > 0 {
		plan.Replicas = types.Int64Value(int64(app.Replicas))
	} else if plan.Replicas.IsUnknown() {
		plan.Replicas = types.Int64Value(1)
	}

	// Branch
	if plan.Branch.IsNull() || plan.Branch.IsUnknown() {
		if app.Branch != "" {
			plan.Branch = types.StringValue(app.Branch)
		}
	}

	// Build type
	if app.BuildType != "" {
		plan.BuildType = types.StringValue(app.BuildType)
	}
	if app.DockerfilePath != "" {
		plan.DockerfilePath = types.StringValue(app.DockerfilePath)
	}
	if app.DockerContextPath != "" {
		plan.DockerContextPath = types.StringValue(app.DockerContextPath)
	}

	// GitHub fields
	if plan.Repository.IsUnknown() && app.Repository != "" {
		plan.Repository = types.StringValue(app.Repository)
	}
	if plan.Owner.IsUnknown() && app.Owner != "" {
		plan.Owner = types.StringValue(app.Owner)
	}
	if plan.GithubId.IsUnknown() && app.GithubId != "" {
		plan.GithubId = types.StringValue(app.GithubId)
	}
	if plan.TriggerType.IsUnknown() && app.TriggerType != "" {
		plan.TriggerType = types.StringValue(app.TriggerType)
	}

	// GitLab fields
	if app.GitlabId != "" {
		plan.GitlabId = types.StringValue(app.GitlabId)
	}
	if app.GitlabProjectId != 0 {
		plan.GitlabProjectId = types.Int64Value(app.GitlabProjectId)
	}
	if app.GitlabRepository != "" {
		plan.GitlabRepository = types.StringValue(app.GitlabRepository)
	}
	if app.GitlabOwner != "" {
		plan.GitlabOwner = types.StringValue(app.GitlabOwner)
	}
	if app.GitlabBranch != "" {
		plan.GitlabBranch = types.StringValue(app.GitlabBranch)
	}
	// Only update build path if plan has a value OR API returns non-default value
	if !plan.GitlabBuildPath.IsNull() || (app.GitlabBuildPath != "" && app.GitlabBuildPath != "/") {
		if app.GitlabBuildPath != "" {
			plan.GitlabBuildPath = types.StringValue(app.GitlabBuildPath)
		}
	}
	if app.GitlabPathNamespace != "" {
		plan.GitlabPathNamespace = types.StringValue(app.GitlabPathNamespace)
	}

	// Bitbucket fields
	if app.BitbucketId != "" {
		plan.BitbucketId = types.StringValue(app.BitbucketId)
	}
	if app.BitbucketRepository != "" {
		plan.BitbucketRepository = types.StringValue(app.BitbucketRepository)
	}
	if app.BitbucketOwner != "" {
		plan.BitbucketOwner = types.StringValue(app.BitbucketOwner)
	}
	if app.BitbucketBranch != "" {
		plan.BitbucketBranch = types.StringValue(app.BitbucketBranch)
	}
	// Only update build path if plan has a value OR API returns non-default value
	if !plan.BitbucketBuildPath.IsNull() || (app.BitbucketBuildPath != "" && app.BitbucketBuildPath != "/") {
		if app.BitbucketBuildPath != "" {
			plan.BitbucketBuildPath = types.StringValue(app.BitbucketBuildPath)
		}
	}

	// Gitea fields
	if app.GiteaId != "" {
		plan.GiteaId = types.StringValue(app.GiteaId)
	}
	if app.GiteaRepository != "" {
		plan.GiteaRepository = types.StringValue(app.GiteaRepository)
	}
	if app.GiteaOwner != "" {
		plan.GiteaOwner = types.StringValue(app.GiteaOwner)
	}
	if app.GiteaBranch != "" {
		plan.GiteaBranch = types.StringValue(app.GiteaBranch)
	}
	// Only update build path if plan has a value OR API returns non-default value
	if !plan.GiteaBuildPath.IsNull() || (app.GiteaBuildPath != "" && app.GiteaBuildPath != "/") {
		if app.GiteaBuildPath != "" {
			plan.GiteaBuildPath = types.StringValue(app.GiteaBuildPath)
		}
	}

	// Custom Git fields
	if app.CustomGitUrl != "" {
		plan.CustomGitUrl = types.StringValue(app.CustomGitUrl)
	}
	if app.CustomGitBranch != "" {
		plan.CustomGitBranch = types.StringValue(app.CustomGitBranch)
	}
	if app.CustomGitSSHKeyId != "" {
		plan.CustomGitSSHKeyID = types.StringValue(app.CustomGitSSHKeyId)
	}
	// Only update build path if plan has a value OR API returns non-default value
	if !plan.CustomGitBuildPath.IsNull() || (app.CustomGitBuildPath != "" && app.CustomGitBuildPath != "/") {
		if app.CustomGitBuildPath != "" {
			plan.CustomGitBuildPath = types.StringValue(app.CustomGitBuildPath)
		}
	}

	// Docker fields
	if app.DockerImage != "" {
		plan.DockerImage = types.StringValue(app.DockerImage)
	}
	if app.RegistryUrl != "" {
		plan.RegistryUrl = types.StringValue(app.RegistryUrl)
	}
	if app.RegistryId != "" {
		plan.RegistryId = types.StringValue(app.RegistryId)
	}

	// Update all computed fields from API response
	plan.CreateEnvFile = types.BoolValue(app.CreateEnvFile)
	plan.Enabled = types.BoolValue(app.Enabled)
	plan.HerokuVersion = types.StringValue(app.HerokuVersion)
	plan.RailpackVersion = types.StringValue(app.RailpackVersion)
	plan.IsStaticSpa = types.BoolValue(app.IsStaticSpa)
	plan.CleanCache = types.BoolValue(app.CleanCache)

	// Preview deployment computed fields
	plan.IsPreviewDeploymentsActive = types.BoolValue(app.IsPreviewDeploymentsActive)
	plan.PreviewPort = types.Int64Value(app.PreviewPort)
	plan.PreviewHttps = types.BoolValue(app.PreviewHttps)
	plan.PreviewPath = types.StringValue(app.PreviewPath)
	plan.PreviewCertificateType = types.StringValue(app.PreviewCertificateType)
	plan.PreviewLimit = types.Int64Value(app.PreviewLimit)
	plan.PreviewRequireCollaboratorPermissions = types.BoolValue(app.PreviewRequireCollaboratorPermissions)

	// Rollback computed field
	plan.RollbackActive = types.BoolValue(app.RollbackActive)

	// New fields: Build type
	if app.Dockerfile != "" {
		plan.Dockerfile = types.StringValue(app.Dockerfile)
	}
	if app.DropBuildPath != "" {
		plan.DropBuildPath = types.StringValue(app.DropBuildPath)
	}

	// New fields: Preview
	if app.PreviewBuildSecrets != "" {
		plan.PreviewBuildSecrets = types.StringValue(app.PreviewBuildSecrets)
	}
	if app.PreviewCustomCertResolver != "" {
		plan.PreviewCustomCertResolver = types.StringValue(app.PreviewCustomCertResolver)
	}

	// Application status (computed)
	plan.ApplicationStatus = types.StringValue(app.ApplicationStatus)

	// Docker Swarm fields - convert maps to JSON strings
	if app.HealthCheckSwarm != nil {
		if jsonBytes, err := json.Marshal(app.HealthCheckSwarm); err == nil {
			plan.HealthCheckSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.RestartPolicySwarm != nil {
		if jsonBytes, err := json.Marshal(app.RestartPolicySwarm); err == nil {
			plan.RestartPolicySwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.PlacementSwarm != nil {
		if jsonBytes, err := json.Marshal(app.PlacementSwarm); err == nil {
			plan.PlacementSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.UpdateConfigSwarm != nil {
		if jsonBytes, err := json.Marshal(app.UpdateConfigSwarm); err == nil {
			plan.UpdateConfigSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.RollbackConfigSwarm != nil {
		if jsonBytes, err := json.Marshal(app.RollbackConfigSwarm); err == nil {
			plan.RollbackConfigSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.ModeSwarm != nil {
		if jsonBytes, err := json.Marshal(app.ModeSwarm); err == nil {
			plan.ModeSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.LabelsSwarm != nil {
		if jsonBytes, err := json.Marshal(app.LabelsSwarm); err == nil {
			plan.LabelsSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.NetworkSwarm != nil {
		if jsonBytes, err := json.Marshal(app.NetworkSwarm); err == nil {
			plan.NetworkSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.StopGracePeriodSwarm != nil {
		plan.StopGracePeriodSwarm = types.Int64Value(*app.StopGracePeriodSwarm)
	}
	if app.EndpointSpecSwarm != nil {
		if jsonBytes, err := json.Marshal(app.EndpointSpecSwarm); err == nil {
			plan.EndpointSpecSwarm = types.StringValue(string(jsonBytes))
		}
	}
}

func readApplicationIntoState(state *ApplicationResourceModel, app *client.Application) {
	state.Name = types.StringValue(app.Name)

	if app.EnvironmentID != "" {
		state.EnvironmentID = types.StringValue(app.EnvironmentID)
	}
	if app.AppName != "" {
		state.AppName = types.StringValue(app.AppName)
	}
	if app.Description != "" {
		state.Description = types.StringValue(app.Description)
	}
	if app.ServerID != "" {
		state.ServerID = types.StringValue(app.ServerID)
	}
	if app.SourceType != "" {
		state.SourceType = types.StringValue(app.SourceType)
	}

	// Git provider fields
	if app.CustomGitUrl != "" {
		state.CustomGitUrl = types.StringValue(app.CustomGitUrl)
	}
	if app.CustomGitBranch != "" {
		state.CustomGitBranch = types.StringValue(app.CustomGitBranch)
	}
	if app.CustomGitSSHKeyId != "" {
		state.CustomGitSSHKeyID = types.StringValue(app.CustomGitSSHKeyId)
	}
	// Only update build path if state has a value OR API returns non-default value
	if !state.CustomGitBuildPath.IsNull() || (app.CustomGitBuildPath != "" && app.CustomGitBuildPath != "/") {
		if app.CustomGitBuildPath != "" {
			state.CustomGitBuildPath = types.StringValue(app.CustomGitBuildPath)
		}
	}
	state.EnableSubmodules = types.BoolValue(app.EnableSubmodules)
	state.CleanCache = types.BoolValue(app.CleanCache)
	// Note: WatchPaths is stored as a JSON string in the API, needs to be parsed for TF list

	// GitHub provider fields
	if app.Repository != "" {
		state.Repository = types.StringValue(app.Repository)
	}
	if app.Branch != "" {
		state.Branch = types.StringValue(app.Branch)
	}
	if app.Owner != "" {
		state.Owner = types.StringValue(app.Owner)
	}
	// Only update build path if state has a value OR API returns non-default value
	if !state.BuildPath.IsNull() || (app.BuildPath != "" && app.BuildPath != "/") {
		if app.BuildPath != "" {
			state.BuildPath = types.StringValue(app.BuildPath)
		}
	}
	if app.GithubId != "" {
		state.GithubId = types.StringValue(app.GithubId)
	}
	if app.TriggerType != "" {
		state.TriggerType = types.StringValue(app.TriggerType)
	}

	// GitLab provider fields
	if app.GitlabId != "" {
		state.GitlabId = types.StringValue(app.GitlabId)
	}
	if app.GitlabProjectId != 0 {
		state.GitlabProjectId = types.Int64Value(app.GitlabProjectId)
	}
	if app.GitlabRepository != "" {
		state.GitlabRepository = types.StringValue(app.GitlabRepository)
	}
	if app.GitlabOwner != "" {
		state.GitlabOwner = types.StringValue(app.GitlabOwner)
	}
	if app.GitlabBranch != "" {
		state.GitlabBranch = types.StringValue(app.GitlabBranch)
	}
	// Only update build path if state has a value OR API returns non-default value
	if !state.GitlabBuildPath.IsNull() || (app.GitlabBuildPath != "" && app.GitlabBuildPath != "/") {
		if app.GitlabBuildPath != "" {
			state.GitlabBuildPath = types.StringValue(app.GitlabBuildPath)
		}
	}
	if app.GitlabPathNamespace != "" {
		state.GitlabPathNamespace = types.StringValue(app.GitlabPathNamespace)
	}

	// Bitbucket provider fields
	if app.BitbucketId != "" {
		state.BitbucketId = types.StringValue(app.BitbucketId)
	}
	if app.BitbucketRepository != "" {
		state.BitbucketRepository = types.StringValue(app.BitbucketRepository)
	}
	if app.BitbucketOwner != "" {
		state.BitbucketOwner = types.StringValue(app.BitbucketOwner)
	}
	if app.BitbucketBranch != "" {
		state.BitbucketBranch = types.StringValue(app.BitbucketBranch)
	}
	// Only update build path if state has a value OR API returns non-default value
	if !state.BitbucketBuildPath.IsNull() || (app.BitbucketBuildPath != "" && app.BitbucketBuildPath != "/") {
		if app.BitbucketBuildPath != "" {
			state.BitbucketBuildPath = types.StringValue(app.BitbucketBuildPath)
		}
	}

	// Gitea provider fields
	if app.GiteaId != "" {
		state.GiteaId = types.StringValue(app.GiteaId)
	}
	if app.GiteaRepository != "" {
		state.GiteaRepository = types.StringValue(app.GiteaRepository)
	}
	if app.GiteaOwner != "" {
		state.GiteaOwner = types.StringValue(app.GiteaOwner)
	}
	if app.GiteaBranch != "" {
		state.GiteaBranch = types.StringValue(app.GiteaBranch)
	}
	// Only update build path if state has a value OR API returns non-default value
	if !state.GiteaBuildPath.IsNull() || (app.GiteaBuildPath != "" && app.GiteaBuildPath != "/") {
		if app.GiteaBuildPath != "" {
			state.GiteaBuildPath = types.StringValue(app.GiteaBuildPath)
		}
	}

	// Docker provider fields
	if app.DockerImage != "" {
		state.DockerImage = types.StringValue(app.DockerImage)
	}
	if app.Username != "" {
		state.Username = types.StringValue(app.Username)
	}
	if app.RegistryUrl != "" {
		state.RegistryUrl = types.StringValue(app.RegistryUrl)
	}
	if app.RegistryId != "" {
		state.RegistryId = types.StringValue(app.RegistryId)
	}

	// Build type fields
	if app.BuildType != "" {
		state.BuildType = types.StringValue(app.BuildType)
	}
	if app.DockerfilePath != "" {
		state.DockerfilePath = types.StringValue(app.DockerfilePath)
	}
	if app.DockerContextPath != "" {
		state.DockerContextPath = types.StringValue(app.DockerContextPath)
	}
	if app.DockerBuildStage != "" {
		state.DockerBuildStage = types.StringValue(app.DockerBuildStage)
	}
	if app.PublishDirectory != "" {
		state.PublishDirectory = types.StringValue(app.PublishDirectory)
	}
	// Always set computed fields from API
	state.HerokuVersion = types.StringValue(app.HerokuVersion)
	state.RailpackVersion = types.StringValue(app.RailpackVersion)
	state.IsStaticSpa = types.BoolValue(app.IsStaticSpa)

	// Environment fields - only update if they were set in config
	if !state.Env.IsNull() {
		if app.Env != "" {
			state.Env = types.StringValue(app.Env)
		}
	}
	if !state.BuildArgs.IsNull() {
		if app.BuildArgs != "" {
			state.BuildArgs = types.StringValue(app.BuildArgs)
		}
	}
	state.CreateEnvFile = types.BoolValue(app.CreateEnvFile)

	// Runtime configuration
	state.AutoDeploy = types.BoolValue(app.AutoDeploy)
	if app.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(app.Replicas))
	}
	if app.MemoryLimit != nil {
		state.MemoryLimit = types.Int64Value(*app.MemoryLimit)
	}
	if app.MemoryReservation != nil {
		state.MemoryReservation = types.Int64Value(*app.MemoryReservation)
	}
	if app.CpuLimit != nil {
		state.CpuLimit = types.Int64Value(*app.CpuLimit)
	}
	if app.CpuReservation != nil {
		state.CpuReservation = types.Int64Value(*app.CpuReservation)
	}
	if app.Command != "" {
		state.Command = types.StringValue(app.Command)
	}
	if app.Args != "" {
		state.Args = types.StringValue(app.Args)
	}

	// Preview deployments - always set computed fields
	state.IsPreviewDeploymentsActive = types.BoolValue(app.IsPreviewDeploymentsActive)
	if app.PreviewEnv != "" {
		state.PreviewEnv = types.StringValue(app.PreviewEnv)
	}
	if app.PreviewBuildArgs != "" {
		state.PreviewBuildArgs = types.StringValue(app.PreviewBuildArgs)
	}
	if app.PreviewWildcard != "" {
		state.PreviewWildcard = types.StringValue(app.PreviewWildcard)
	}
	state.PreviewPort = types.Int64Value(app.PreviewPort)
	state.PreviewHttps = types.BoolValue(app.PreviewHttps)
	state.PreviewPath = types.StringValue(app.PreviewPath)
	state.PreviewCertificateType = types.StringValue(app.PreviewCertificateType)
	state.PreviewLimit = types.Int64Value(app.PreviewLimit)
	state.PreviewRequireCollaboratorPermissions = types.BoolValue(app.PreviewRequireCollaboratorPermissions)

	// Rollback
	state.RollbackActive = types.BoolValue(app.RollbackActive)
	if app.RollbackRegistryId != "" {
		state.RollbackRegistryId = types.StringValue(app.RollbackRegistryId)
	}

	// Build server
	if app.BuildServerId != "" {
		state.BuildServerId = types.StringValue(app.BuildServerId)
	}
	if app.BuildRegistryId != "" {
		state.BuildRegistryId = types.StringValue(app.BuildRegistryId)
	}

	// Display
	if app.Title != "" {
		state.Title = types.StringValue(app.Title)
	}
	if app.Subtitle != "" {
		state.Subtitle = types.StringValue(app.Subtitle)
	}
	state.Enabled = types.BoolValue(app.Enabled)

	// New fields: Build type
	if app.Dockerfile != "" {
		state.Dockerfile = types.StringValue(app.Dockerfile)
	}
	if app.DropBuildPath != "" {
		state.DropBuildPath = types.StringValue(app.DropBuildPath)
	}

	// New fields: Preview
	if app.PreviewBuildSecrets != "" {
		state.PreviewBuildSecrets = types.StringValue(app.PreviewBuildSecrets)
	}
	if app.PreviewCustomCertResolver != "" {
		state.PreviewCustomCertResolver = types.StringValue(app.PreviewCustomCertResolver)
	}

	// Application status (computed)
	state.ApplicationStatus = types.StringValue(app.ApplicationStatus)

	// Docker Swarm fields - convert maps to JSON strings
	if app.HealthCheckSwarm != nil {
		if jsonBytes, err := json.Marshal(app.HealthCheckSwarm); err == nil {
			state.HealthCheckSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.RestartPolicySwarm != nil {
		if jsonBytes, err := json.Marshal(app.RestartPolicySwarm); err == nil {
			state.RestartPolicySwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.PlacementSwarm != nil {
		if jsonBytes, err := json.Marshal(app.PlacementSwarm); err == nil {
			state.PlacementSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.UpdateConfigSwarm != nil {
		if jsonBytes, err := json.Marshal(app.UpdateConfigSwarm); err == nil {
			state.UpdateConfigSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.RollbackConfigSwarm != nil {
		if jsonBytes, err := json.Marshal(app.RollbackConfigSwarm); err == nil {
			state.RollbackConfigSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.ModeSwarm != nil {
		if jsonBytes, err := json.Marshal(app.ModeSwarm); err == nil {
			state.ModeSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.LabelsSwarm != nil {
		if jsonBytes, err := json.Marshal(app.LabelsSwarm); err == nil {
			state.LabelsSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.NetworkSwarm != nil {
		if jsonBytes, err := json.Marshal(app.NetworkSwarm); err == nil {
			state.NetworkSwarm = types.StringValue(string(jsonBytes))
		}
	}
	if app.StopGracePeriodSwarm != nil {
		state.StopGracePeriodSwarm = types.Int64Value(*app.StopGracePeriodSwarm)
	}
	if app.EndpointSpecSwarm != nil {
		if jsonBytes, err := json.Marshal(app.EndpointSpecSwarm); err == nil {
			state.EndpointSpecSwarm = types.StringValue(string(jsonBytes))
		}
	}
}
