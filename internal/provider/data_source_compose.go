package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ComposeDataSource{}

func NewComposeDataSource() datasource.DataSource {
	return &ComposeDataSource{}
}

type ComposeDataSource struct {
	client *client.DokployClient
}

type ComposeDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ServerID      types.String `tfsdk:"server_id"`

	// Compose file
	ComposeFileContent types.String `tfsdk:"compose_file_content"`
	ComposePath        types.String `tfsdk:"compose_path"`
	ComposeType        types.String `tfsdk:"compose_type"`

	// Source configuration
	SourceType types.String `tfsdk:"source_type"`

	// Custom Git provider settings
	CustomGitUrl       types.String `tfsdk:"custom_git_url"`
	CustomGitBranch    types.String `tfsdk:"custom_git_branch"`
	CustomGitSSHKeyID  types.String `tfsdk:"custom_git_ssh_key_id"`
	CustomGitBuildPath types.String `tfsdk:"custom_git_build_path"`
	EnableSubmodules   types.Bool   `tfsdk:"enable_submodules"`

	// GitHub provider settings
	Repository  types.String `tfsdk:"repository"`
	Branch      types.String `tfsdk:"branch"`
	Owner       types.String `tfsdk:"owner"`
	GithubId    types.String `tfsdk:"github_id"`
	TriggerType types.String `tfsdk:"trigger_type"`

	// GitLab provider settings
	GitlabId            types.String `tfsdk:"gitlab_id"`
	GitlabProjectId     types.Int64  `tfsdk:"gitlab_project_id"`
	GitlabRepository    types.String `tfsdk:"gitlab_repository"`
	GitlabOwner         types.String `tfsdk:"gitlab_owner"`
	GitlabBranch        types.String `tfsdk:"gitlab_branch"`
	GitlabBuildPath     types.String `tfsdk:"gitlab_build_path"`
	GitlabPathNamespace types.String `tfsdk:"gitlab_path_namespace"`

	// Bitbucket provider settings
	BitbucketId         types.String `tfsdk:"bitbucket_id"`
	BitbucketRepository types.String `tfsdk:"bitbucket_repository"`
	BitbucketOwner      types.String `tfsdk:"bitbucket_owner"`
	BitbucketBranch     types.String `tfsdk:"bitbucket_branch"`
	BitbucketBuildPath  types.String `tfsdk:"bitbucket_build_path"`

	// Gitea provider settings
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
}

func (d *ComposeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compose"
}

func (d *ComposeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Dokploy compose stack by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the compose stack.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the compose stack.",
			},
			"app_name": schema.StringAttribute{
				Computed:    true,
				Description: "The app name used for Docker service naming.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the compose stack.",
			},
			"environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The environment ID this compose stack belongs to.",
			},
			"server_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server ID the compose stack is deployed to.",
			},

			// Compose file
			"compose_file_content": schema.StringAttribute{
				Computed:    true,
				Description: "Raw docker-compose.yml content.",
			},
			"compose_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the docker-compose.yml file in the repository.",
			},
			"compose_type": schema.StringAttribute{
				Computed:    true,
				Description: "The compose type: 'docker-compose' or 'stack' for Docker Swarm.",
			},

			// Source type
			"source_type": schema.StringAttribute{
				Computed:    true,
				Description: "The source type: github, gitlab, bitbucket, gitea, git, or raw.",
			},

			// Custom Git provider settings
			"custom_git_url": schema.StringAttribute{
				Computed:    true,
				Description: "Custom Git repository URL.",
			},
			"custom_git_branch": schema.StringAttribute{
				Computed:    true,
				Description: "Branch for custom Git repository.",
			},
			"custom_git_ssh_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "SSH key ID for accessing the custom Git repository.",
			},
			"custom_git_build_path": schema.StringAttribute{
				Computed:    true,
				Description: "Build path within the custom Git repository.",
			},
			"enable_submodules": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Git submodules support is enabled.",
			},

			// GitHub provider settings
			"repository": schema.StringAttribute{
				Computed:    true,
				Description: "GitHub repository name.",
			},
			"branch": schema.StringAttribute{
				Computed:    true,
				Description: "GitHub branch name.",
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "GitHub repository owner.",
			},
			"github_id": schema.StringAttribute{
				Computed:    true,
				Description: "GitHub App installation ID.",
			},
			"trigger_type": schema.StringAttribute{
				Computed:    true,
				Description: "Trigger type for deployments: 'push' or 'tag'.",
			},

			// GitLab provider settings
			"gitlab_id": schema.StringAttribute{
				Computed:    true,
				Description: "GitLab integration ID.",
			},
			"gitlab_project_id": schema.Int64Attribute{
				Computed:    true,
				Description: "GitLab project ID.",
			},
			"gitlab_repository": schema.StringAttribute{
				Computed:    true,
				Description: "GitLab repository name.",
			},
			"gitlab_owner": schema.StringAttribute{
				Computed:    true,
				Description: "GitLab repository owner/group.",
			},
			"gitlab_branch": schema.StringAttribute{
				Computed:    true,
				Description: "GitLab branch to deploy from.",
			},
			"gitlab_build_path": schema.StringAttribute{
				Computed:    true,
				Description: "Build path within the GitLab repository.",
			},
			"gitlab_path_namespace": schema.StringAttribute{
				Computed:    true,
				Description: "GitLab path namespace.",
			},

			// Bitbucket provider settings
			"bitbucket_id": schema.StringAttribute{
				Computed:    true,
				Description: "Bitbucket integration ID.",
			},
			"bitbucket_repository": schema.StringAttribute{
				Computed:    true,
				Description: "Bitbucket repository name.",
			},
			"bitbucket_owner": schema.StringAttribute{
				Computed:    true,
				Description: "Bitbucket repository owner/workspace.",
			},
			"bitbucket_branch": schema.StringAttribute{
				Computed:    true,
				Description: "Bitbucket branch to deploy from.",
			},
			"bitbucket_build_path": schema.StringAttribute{
				Computed:    true,
				Description: "Build path within the Bitbucket repository.",
			},

			// Gitea provider settings
			"gitea_id": schema.StringAttribute{
				Computed:    true,
				Description: "Gitea integration ID.",
			},
			"gitea_repository": schema.StringAttribute{
				Computed:    true,
				Description: "Gitea repository name.",
			},
			"gitea_owner": schema.StringAttribute{
				Computed:    true,
				Description: "Gitea repository owner/organization.",
			},
			"gitea_branch": schema.StringAttribute{
				Computed:    true,
				Description: "Gitea branch to deploy from.",
			},
			"gitea_build_path": schema.StringAttribute{
				Computed:    true,
				Description: "Build path within the Gitea repository.",
			},

			// Environment
			"env": schema.StringAttribute{
				Computed:    true,
				Description: "Environment variables in KEY=VALUE format.",
			},

			// Runtime configuration
			"auto_deploy": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether auto-deploy is enabled.",
			},

			// Advanced configuration
			"command": schema.StringAttribute{
				Computed:    true,
				Description: "Custom command to run for deployment.",
			},
			"suffix": schema.StringAttribute{
				Computed:    true,
				Description: "Suffix added to service names.",
			},
			"randomize": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether service names are randomized.",
			},
			"isolated_deployment": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether isolated deployments are enabled.",
			},
			"isolated_deployments_volume": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether isolated deployment volumes are enabled.",
			},
			"watch_paths": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Paths watched for changes to trigger deployments.",
			},

			// Computed status fields
			"compose_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status: idle, running, done, or error.",
			},
			"refresh_token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook refresh token for triggering deployments.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the compose stack was created.",
			},
		},
	}
}

func (d *ComposeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ComposeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ComposeDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	comp, err := d.client.GetCompose(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Compose", err.Error())
		return
	}

	// Map compose data
	data.Name = types.StringValue(comp.Name)
	data.EnvironmentID = types.StringValue(comp.EnvironmentID)
	data.SourceType = types.StringValue(comp.SourceType)
	data.AutoDeploy = types.BoolValue(comp.AutoDeploy)
	data.EnableSubmodules = types.BoolValue(comp.EnableSubmodules)
	data.Randomize = types.BoolValue(comp.Randomize)
	data.IsolatedDeployment = types.BoolValue(comp.IsolatedDeployment)
	data.IsolatedDeploymentsVolume = types.BoolValue(comp.IsolatedDeploymentsVolume)

	// Optional string fields
	if comp.AppName != "" {
		data.AppName = types.StringValue(comp.AppName)
	}
	if comp.Description != "" {
		data.Description = types.StringValue(comp.Description)
	}
	if comp.ServerID != "" {
		data.ServerID = types.StringValue(comp.ServerID)
	}
	if comp.ComposeFile != "" {
		data.ComposeFileContent = types.StringValue(comp.ComposeFile)
	}
	if comp.ComposePath != "" {
		data.ComposePath = types.StringValue(comp.ComposePath)
	}
	if comp.ComposeType != "" {
		data.ComposeType = types.StringValue(comp.ComposeType)
	}
	if comp.CustomGitUrl != "" {
		data.CustomGitUrl = types.StringValue(comp.CustomGitUrl)
	}
	if comp.CustomGitBranch != "" {
		data.CustomGitBranch = types.StringValue(comp.CustomGitBranch)
	}
	if comp.CustomGitSSHKeyId != "" {
		data.CustomGitSSHKeyID = types.StringValue(comp.CustomGitSSHKeyId)
	}
	if comp.CustomGitBuildPath != "" {
		data.CustomGitBuildPath = types.StringValue(comp.CustomGitBuildPath)
	}
	if comp.Repository != "" {
		data.Repository = types.StringValue(comp.Repository)
	}
	if comp.Branch != "" {
		data.Branch = types.StringValue(comp.Branch)
	}
	if comp.Owner != "" {
		data.Owner = types.StringValue(comp.Owner)
	}
	if comp.GithubId != "" {
		data.GithubId = types.StringValue(comp.GithubId)
	}
	if comp.TriggerType != "" {
		data.TriggerType = types.StringValue(comp.TriggerType)
	}
	if comp.GitlabId != "" {
		data.GitlabId = types.StringValue(comp.GitlabId)
	}
	if comp.GitlabProjectId != 0 {
		data.GitlabProjectId = types.Int64Value(comp.GitlabProjectId)
	}
	if comp.GitlabRepository != "" {
		data.GitlabRepository = types.StringValue(comp.GitlabRepository)
	}
	if comp.GitlabOwner != "" {
		data.GitlabOwner = types.StringValue(comp.GitlabOwner)
	}
	if comp.GitlabBranch != "" {
		data.GitlabBranch = types.StringValue(comp.GitlabBranch)
	}
	if comp.GitlabBuildPath != "" {
		data.GitlabBuildPath = types.StringValue(comp.GitlabBuildPath)
	}
	if comp.GitlabPathNamespace != "" {
		data.GitlabPathNamespace = types.StringValue(comp.GitlabPathNamespace)
	}
	if comp.BitbucketId != "" {
		data.BitbucketId = types.StringValue(comp.BitbucketId)
	}
	if comp.BitbucketRepository != "" {
		data.BitbucketRepository = types.StringValue(comp.BitbucketRepository)
	}
	if comp.BitbucketOwner != "" {
		data.BitbucketOwner = types.StringValue(comp.BitbucketOwner)
	}
	if comp.BitbucketBranch != "" {
		data.BitbucketBranch = types.StringValue(comp.BitbucketBranch)
	}
	if comp.BitbucketBuildPath != "" {
		data.BitbucketBuildPath = types.StringValue(comp.BitbucketBuildPath)
	}
	if comp.GiteaId != "" {
		data.GiteaId = types.StringValue(comp.GiteaId)
	}
	if comp.GiteaRepository != "" {
		data.GiteaRepository = types.StringValue(comp.GiteaRepository)
	}
	if comp.GiteaOwner != "" {
		data.GiteaOwner = types.StringValue(comp.GiteaOwner)
	}
	if comp.GiteaBranch != "" {
		data.GiteaBranch = types.StringValue(comp.GiteaBranch)
	}
	if comp.GiteaBuildPath != "" {
		data.GiteaBuildPath = types.StringValue(comp.GiteaBuildPath)
	}
	if comp.Env != "" {
		data.Env = types.StringValue(comp.Env)
	}
	if comp.Command != "" {
		data.Command = types.StringValue(comp.Command)
	}
	if comp.Suffix != "" {
		data.Suffix = types.StringValue(comp.Suffix)
	}
	if comp.ComposeStatus != "" {
		data.ComposeStatus = types.StringValue(comp.ComposeStatus)
	}
	if comp.RefreshToken != "" {
		data.RefreshToken = types.StringValue(comp.RefreshToken)
	}
	if comp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(comp.CreatedAt)
	}

	// WatchPaths - convert []string to types.List
	if len(comp.WatchPaths) > 0 {
		watchPathsList, d := types.ListValueFrom(ctx, types.StringType, comp.WatchPaths)
		resp.Diagnostics.Append(d...)
		data.WatchPaths = watchPathsList
	} else {
		data.WatchPaths = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
