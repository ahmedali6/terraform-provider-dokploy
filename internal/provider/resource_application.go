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
	ProjectID     types.String `tfsdk:"project_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`

	// Source type
	SourceType types.String `tfsdk:"source_type"`

	// Git provider settings (for source_type = "git")
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	CustomGitBuildPath types.String `tfsdk:"custom_git_build_path"`
	EnableSubmodules   types.Bool   `tfsdk:"enable_submodules"`

	// GitHub provider settings (for source_type = "github")
	Repository  types.String `tfsdk:"repository"`
	Branch      types.String `tfsdk:"branch"`
	Owner       types.String `tfsdk:"owner"`
	BuildPath   types.String `tfsdk:"build_path"`
	GithubId    types.String `tfsdk:"github_id"`
	TriggerType types.String `tfsdk:"trigger_type"`

	// Docker provider settings
	DockerImage types.String `tfsdk:"docker_image"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	RegistryUrl types.String `tfsdk:"registry_url"`

	// Build type settings
	BuildType         types.String `tfsdk:"build_type"`
	DockerfilePath    types.String `tfsdk:"dockerfile_path"`
	DockerContextPath types.String `tfsdk:"docker_context_path"`
	DockerBuildStage  types.String `tfsdk:"docker_build_stage"`
	PublishDirectory  types.String `tfsdk:"publish_directory"`

	// Environment settings
	Env       types.String `tfsdk:"env"`
	BuildArgs types.String `tfsdk:"build_args"`

	// Runtime configuration
	AutoDeploy        types.Bool   `tfsdk:"auto_deploy"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	MemoryLimit       types.Int64  `tfsdk:"memory_limit"`
	MemoryReservation types.Int64  `tfsdk:"memory_reservation"`
	CpuLimit          types.Int64  `tfsdk:"cpu_limit"`
	CpuReservation    types.Int64  `tfsdk:"cpu_reservation"`
	Command           types.String `tfsdk:"command"`

	// Deployment options
	DeployOnCreate types.Bool   `tfsdk:"deploy_on_create"`
	ServerID       types.String `tfsdk:"server_id"`
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID this application belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The environment ID this application belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the application.",
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

			// Source type
			"source_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The source type for the application: github, gitlab, bitbucket, git, docker, or drop.",
			},

			// Git provider settings (for source_type = "git")
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// GitHub provider settings (for source_type = "github")
			"repository": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Repository name for GitHub source (e.g., 'my-repo').",
			},
			"branch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Branch to deploy from.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Repository owner/organization for GitHub source.",
			},
			"build_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Build path within the repository for GitHub source.",
			},
			"github_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "GitHub App installation ID. Required for GitHub source type.",
			},
			"trigger_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Trigger type for deployments: 'push' (default) or 'tag'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Docker provider settings
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Description: "Docker image to use (for source_type 'docker').",
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
				Description: "Docker registry URL.",
			},

			// Build type settings
			"build_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Build type: dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, or railpack.",
			},
			"dockerfile_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path to the Dockerfile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"docker_context_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker build context path.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"docker_build_stage": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Target stage for multi-stage Docker builds.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"publish_directory": schema.StringAttribute{
				Optional:    true,
				Description: "Publish directory for static builds.",
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

			// Runtime configuration
			"auto_deploy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable automatic deployment on Git push.",
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of replicas to run.",
			},
			"memory_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Memory limit in bytes.",
			},
			"memory_reservation": schema.Int64Attribute{
				Optional:    true,
				Description: "Memory reservation in bytes.",
			},
			"cpu_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "CPU limit (in millicores, e.g., 1000 = 1 CPU).",
			},
			"cpu_reservation": schema.Int64Attribute{
				Optional:    true,
				Description: "CPU reservation (in millicores).",
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run.",
			},

			// Deployment options
			"deploy_on_create": schema.BoolAttribute{
				Optional:    true,
				Description: "Trigger a deployment after creating the application.",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID to deploy the application to. If not specified, deploys to the default server.",
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

	// Set sensible defaults
	if plan.Branch.IsUnknown() || plan.Branch.IsNull() {
		plan.Branch = types.StringValue("main")
	}
	if plan.BuildType.IsUnknown() || plan.BuildType.IsNull() {
		plan.BuildType = types.StringValue("nixpacks")
	}
	if plan.DockerfilePath.IsUnknown() || plan.DockerfilePath.IsNull() {
		plan.DockerfilePath = types.StringValue("./Dockerfile")
	}
	if plan.DockerContextPath.IsUnknown() || plan.DockerContextPath.IsNull() {
		plan.DockerContextPath = types.StringValue("/")
	}

	// Default SourceType logic
	if plan.SourceType.IsUnknown() || plan.SourceType.IsNull() {
		if !plan.DockerImage.IsNull() && !plan.DockerImage.IsUnknown() && plan.DockerImage.ValueString() != "" {
			plan.SourceType = types.StringValue("docker")
		} else if !plan.CustomGitUrl.IsNull() && !plan.CustomGitUrl.IsUnknown() && plan.CustomGitUrl.ValueString() != "" {
			plan.SourceType = types.StringValue("git")
		} else {
			plan.SourceType = types.StringValue("github")
		}
	}

	// 1. Create the application with minimal fields
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
	if createdApp.EnvironmentID != "" {
		plan.EnvironmentID = types.StringValue(createdApp.EnvironmentID)
	}
	if createdApp.AppName != "" {
		plan.AppName = types.StringValue(createdApp.AppName)
	}

	// 2. Update general settings (sourceType, autoDeploy, replicas, etc.)
	generalApp := client.Application{
		ID:         createdApp.ID,
		Name:       plan.Name.ValueString(),
		AppName:    plan.AppName.ValueString(),
		SourceType: plan.SourceType.ValueString(),
		AutoDeploy: plan.AutoDeploy.ValueBool(),
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

	_, err = r.client.UpdateApplicationGeneral(generalApp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application general settings", err.Error())
		return
	}

	// 3. Save build type settings if applicable
	if plan.SourceType.ValueString() != "docker" {
		err = r.client.SaveBuildType(
			createdApp.ID,
			plan.BuildType.ValueString(),
			plan.DockerfilePath.ValueString(),
			plan.DockerContextPath.ValueString(),
			plan.DockerBuildStage.ValueString(),
			plan.PublishDirectory.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError("Error saving build type", err.Error())
			return
		}
	}

	// 4. Configure source provider based on source_type
	sourceType := plan.SourceType.ValueString()
	switch sourceType {
	case "github":
		// Save GitHub provider settings
		githubInput := client.SaveGithubProviderInput{
			ApplicationID:    createdApp.ID,
			Repository:       plan.Repository.ValueString(),
			Branch:           plan.Branch.ValueString(),
			Owner:            plan.Owner.ValueString(),
			BuildPath:        plan.BuildPath.ValueString(),
			GithubId:         plan.GithubId.ValueString(),
			EnableSubmodules: plan.EnableSubmodules.ValueBool(),
			TriggerType:      plan.TriggerType.ValueString(),
		}
		err = r.client.SaveGithubProvider(githubInput)
		if err != nil {
			resp.Diagnostics.AddError("Error saving github provider", err.Error())
			return
		}

	case "git":
		// Save Git provider settings
		gitInput := client.SaveGitProviderInput{
			ApplicationID:      createdApp.ID,
			CustomGitUrl:       plan.CustomGitUrl.ValueString(),
			CustomGitBranch:    plan.CustomGitBranch.ValueString(),
			CustomGitBuildPath: plan.CustomGitBuildPath.ValueString(),
			CustomGitSSHKeyId:  plan.CustomGitSSHKeyID.ValueString(),
			EnableSubmodules:   plan.EnableSubmodules.ValueBool(),
		}
		err = r.client.SaveGitProvider(gitInput)
		if err != nil {
			resp.Diagnostics.AddError("Error saving git provider", err.Error())
			return
		}

	case "docker":
		// Save Docker provider settings
		dockerInput := client.SaveDockerProviderInput{
			ApplicationID: createdApp.ID,
			DockerImage:   plan.DockerImage.ValueString(),
			Username:      plan.Username.ValueString(),
			Password:      plan.Password.ValueString(),
			RegistryUrl:   plan.RegistryUrl.ValueString(),
		}
		err = r.client.SaveDockerProvider(dockerInput)
		if err != nil {
			resp.Diagnostics.AddError("Error saving docker provider", err.Error())
			return
		}
	}

	// 5. Save environment variables if provided
	if !plan.Env.IsNull() && !plan.Env.IsUnknown() && plan.Env.ValueString() != "" {
		envInput := client.SaveEnvironmentInput{
			ApplicationID: createdApp.ID,
			Env:           plan.Env.ValueString(),
			BuildArgs:     plan.BuildArgs.ValueString(),
		}
		err = r.client.SaveEnvironment(envInput)
		if err != nil {
			resp.Diagnostics.AddError("Error saving environment", err.Error())
			return
		}
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

// updatePlanFromApplication updates the plan with values from the API response
func updatePlanFromApplication(plan *ApplicationResourceModel, app *client.Application) {
	if app.EnvironmentID != "" {
		plan.EnvironmentID = types.StringValue(app.EnvironmentID)
	}
	if app.AppName != "" {
		plan.AppName = types.StringValue(app.AppName)
	}
	if app.Repository != "" {
		plan.Repository = types.StringValue(app.Repository)
	}
	if app.Branch != "" {
		plan.Branch = types.StringValue(app.Branch)
	}
	if app.Owner != "" {
		plan.Owner = types.StringValue(app.Owner)
	}
	if app.GithubId != "" {
		plan.GithubId = types.StringValue(app.GithubId)
	}
	if app.BuildType != "" {
		plan.BuildType = types.StringValue(app.BuildType)
	}
	if app.SourceType != "" {
		plan.SourceType = types.StringValue(app.SourceType)
	}

	// For computed fields that may be empty, always set them to avoid "unknown after apply" errors
	plan.DockerfilePath = types.StringValue(app.DockerfilePath)
	plan.DockerContextPath = types.StringValue(app.DockerContextPath)
	plan.DockerBuildStage = types.StringValue(app.DockerBuildStage)
	plan.BuildPath = types.StringValue(app.BuildPath)
	plan.TriggerType = types.StringValue(app.TriggerType)

	if app.CustomGitUrl != "" {
		plan.CustomGitUrl = types.StringValue(app.CustomGitUrl)
	}
	if app.CustomGitBranch != "" {
		plan.CustomGitBranch = types.StringValue(app.CustomGitBranch)
	}
	if app.CustomGitSSHKeyId != "" {
		plan.CustomGitSSHKeyID = types.StringValue(app.CustomGitSSHKeyId)
	}
	if app.CustomGitBuildPath != "" {
		plan.CustomGitBuildPath = types.StringValue(app.CustomGitBuildPath)
	}
	if app.DockerImage != "" {
		plan.DockerImage = types.StringValue(app.DockerImage)
	}
	if app.RegistryUrl != "" {
		plan.RegistryUrl = types.StringValue(app.RegistryUrl)
	}

	plan.AutoDeploy = types.BoolValue(app.AutoDeploy)
	plan.EnableSubmodules = types.BoolValue(app.EnableSubmodules)
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
	state.Name = types.StringValue(app.Name)

	if app.ProjectID != "" {
		state.ProjectID = types.StringValue(app.ProjectID)
	}
	if app.EnvironmentID != "" {
		state.EnvironmentID = types.StringValue(app.EnvironmentID)
	}
	if app.AppName != "" {
		state.AppName = types.StringValue(app.AppName)
	}
	if app.Description != "" {
		state.Description = types.StringValue(app.Description)
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
	if app.CustomGitBuildPath != "" {
		state.CustomGitBuildPath = types.StringValue(app.CustomGitBuildPath)
	}
	state.EnableSubmodules = types.BoolValue(app.EnableSubmodules)

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
	if app.BuildPath != "" {
		state.BuildPath = types.StringValue(app.BuildPath)
	}
	if app.GithubId != "" {
		state.GithubId = types.StringValue(app.GithubId)
	}
	if app.TriggerType != "" {
		state.TriggerType = types.StringValue(app.TriggerType)
	}

	// Docker provider fields
	if app.DockerImage != "" {
		state.DockerImage = types.StringValue(app.DockerImage)
	}
	if app.Username != "" {
		state.Username = types.StringValue(app.Username)
	}
	// Don't read password back - it's sensitive
	if app.RegistryUrl != "" {
		state.RegistryUrl = types.StringValue(app.RegistryUrl)
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

	// Environment fields
	if app.Env != "" {
		state.Env = types.StringValue(app.Env)
	}
	if app.BuildArgs != "" {
		state.BuildArgs = types.StringValue(app.BuildArgs)
	}

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

	// Server ID - preserve from state as API may not return it
	if app.ServerID != "" {
		state.ServerID = types.StringValue(app.ServerID)
	}

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

	// Use ID from state (it's the authoritative source during updates)
	appID := state.ID.ValueString()
	plan.ID = state.ID

	// Set sensible defaults for unknown values
	if plan.Branch.IsUnknown() {
		plan.Branch = types.StringValue("main")
	}
	if plan.BuildType.IsUnknown() {
		plan.BuildType = types.StringValue("nixpacks")
	}
	if plan.DockerfilePath.IsUnknown() || plan.DockerfilePath.IsNull() {
		plan.DockerfilePath = types.StringValue("./Dockerfile")
	}
	if plan.DockerContextPath.IsUnknown() || plan.DockerContextPath.IsNull() {
		plan.DockerContextPath = types.StringValue("/")
	}

	// 1. Update general settings
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

	_, err := r.client.UpdateApplicationGeneral(generalApp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application general settings", err.Error())
		return
	}

	// 2. Update build type if changed (for non-docker source types)
	sourceType := plan.SourceType.ValueString()
	if sourceType != "docker" {
		buildTypeChanged := !plan.BuildType.Equal(state.BuildType) ||
			!plan.DockerfilePath.Equal(state.DockerfilePath) ||
			!plan.DockerContextPath.Equal(state.DockerContextPath) ||
			!plan.DockerBuildStage.Equal(state.DockerBuildStage) ||
			!plan.PublishDirectory.Equal(state.PublishDirectory)

		if buildTypeChanged {
			err = r.client.SaveBuildType(
				appID,
				plan.BuildType.ValueString(),
				plan.DockerfilePath.ValueString(),
				plan.DockerContextPath.ValueString(),
				plan.DockerBuildStage.ValueString(),
				plan.PublishDirectory.ValueString(),
			)
			if err != nil {
				resp.Diagnostics.AddError("Error saving build type", err.Error())
				return
			}
		}
	}

	// 3. Update source provider settings based on source_type
	switch sourceType {
	case "github":
		githubChanged := !plan.Repository.Equal(state.Repository) ||
			!plan.Branch.Equal(state.Branch) ||
			!plan.Owner.Equal(state.Owner) ||
			!plan.BuildPath.Equal(state.BuildPath) ||
			!plan.GithubId.Equal(state.GithubId) ||
			!plan.EnableSubmodules.Equal(state.EnableSubmodules) ||
			!plan.TriggerType.Equal(state.TriggerType)

		if githubChanged {
			githubInput := client.SaveGithubProviderInput{
				ApplicationID:    appID,
				Repository:       plan.Repository.ValueString(),
				Branch:           plan.Branch.ValueString(),
				Owner:            plan.Owner.ValueString(),
				BuildPath:        plan.BuildPath.ValueString(),
				GithubId:         plan.GithubId.ValueString(),
				EnableSubmodules: plan.EnableSubmodules.ValueBool(),
				TriggerType:      plan.TriggerType.ValueString(),
			}
			err = r.client.SaveGithubProvider(githubInput)
			if err != nil {
				resp.Diagnostics.AddError("Error saving github provider", err.Error())
				return
			}
		}

	case "git":
		gitChanged := !plan.CustomGitUrl.Equal(state.CustomGitUrl) ||
			!plan.CustomGitBranch.Equal(state.CustomGitBranch) ||
			!plan.CustomGitSSHKeyID.Equal(state.CustomGitSSHKeyID) ||
			!plan.CustomGitBuildPath.Equal(state.CustomGitBuildPath) ||
			!plan.EnableSubmodules.Equal(state.EnableSubmodules)

		if gitChanged {
			gitInput := client.SaveGitProviderInput{
				ApplicationID:      appID,
				CustomGitUrl:       plan.CustomGitUrl.ValueString(),
				CustomGitBranch:    plan.CustomGitBranch.ValueString(),
				CustomGitBuildPath: plan.CustomGitBuildPath.ValueString(),
				CustomGitSSHKeyId:  plan.CustomGitSSHKeyID.ValueString(),
				EnableSubmodules:   plan.EnableSubmodules.ValueBool(),
			}
			err = r.client.SaveGitProvider(gitInput)
			if err != nil {
				resp.Diagnostics.AddError("Error saving git provider", err.Error())
				return
			}
		}

	case "docker":
		dockerChanged := !plan.DockerImage.Equal(state.DockerImage) ||
			!plan.Username.Equal(state.Username) ||
			!plan.Password.Equal(state.Password) ||
			!plan.RegistryUrl.Equal(state.RegistryUrl)

		if dockerChanged {
			dockerInput := client.SaveDockerProviderInput{
				ApplicationID: appID,
				DockerImage:   plan.DockerImage.ValueString(),
				Username:      plan.Username.ValueString(),
				Password:      plan.Password.ValueString(),
				RegistryUrl:   plan.RegistryUrl.ValueString(),
			}
			err = r.client.SaveDockerProvider(dockerInput)
			if err != nil {
				resp.Diagnostics.AddError("Error saving docker provider", err.Error())
				return
			}
		}
	}

	// 4. Update environment if changed
	envChanged := !plan.Env.Equal(state.Env) || !plan.BuildArgs.Equal(state.BuildArgs)
	if envChanged {
		envInput := client.SaveEnvironmentInput{
			ApplicationID: appID,
			Env:           plan.Env.ValueString(),
			BuildArgs:     plan.BuildArgs.ValueString(),
		}
		err = r.client.SaveEnvironment(envInput)
		if err != nil {
			resp.Diagnostics.AddError("Error saving environment", err.Error())
			return
		}
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
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting application", err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
