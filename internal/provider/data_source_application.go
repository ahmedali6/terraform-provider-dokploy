package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

type ApplicationDataSource struct {
	client *client.DokployClient
}

type ApplicationDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ServerID      types.String `tfsdk:"server_id"`

	// Source type
	SourceType types.String `tfsdk:"source_type"`

	// Build type
	BuildType types.String `tfsdk:"build_type"`

	// Docker settings
	DockerImage types.String `tfsdk:"docker_image"`
	RegistryUrl types.String `tfsdk:"registry_url"`

	// Git settings
	CustomGitUrl    types.String `tfsdk:"custom_git_url"`
	CustomGitBranch types.String `tfsdk:"custom_git_branch"`

	// GitHub settings
	Repository types.String `tfsdk:"repository"`
	Branch     types.String `tfsdk:"branch"`
	Owner      types.String `tfsdk:"owner"`

	// Runtime
	AutoDeploy        types.Bool   `tfsdk:"auto_deploy"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	ApplicationStatus types.String `tfsdk:"application_status"`

	// Traefik
	TraefikConfig types.String `tfsdk:"traefik_config"`

	// Timestamps
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *ApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Dokploy application by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the application.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the application.",
			},
			"app_name": schema.StringAttribute{
				Computed:    true,
				Description: "The app name used for Docker container naming.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the application.",
			},
			"environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The environment ID this application belongs to.",
			},
			"server_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server ID the application is deployed to.",
			},
			"source_type": schema.StringAttribute{
				Computed:    true,
				Description: "The source type: github, gitlab, bitbucket, gitea, git, docker, or drop.",
			},
			"build_type": schema.StringAttribute{
				Computed:    true,
				Description: "The build type: dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, or railpack.",
			},
			"docker_image": schema.StringAttribute{
				Computed:    true,
				Description: "Docker image name (for docker source type).",
			},
			"registry_url": schema.StringAttribute{
				Computed:    true,
				Description: "Docker registry URL.",
			},
			"custom_git_url": schema.StringAttribute{
				Computed:    true,
				Description: "Custom Git repository URL.",
			},
			"custom_git_branch": schema.StringAttribute{
				Computed:    true,
				Description: "Branch for custom Git repository.",
			},
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
			"auto_deploy": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether auto-deploy is enabled.",
			},
			"replicas": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of replicas.",
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status: idle, running, done, or error.",
			},
			"traefik_config": schema.StringAttribute{
				Computed:    true,
				Description: "Custom Traefik configuration for the application.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the application was created.",
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplication(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Application", err.Error())
		return
	}

	// Map application data
	data.Name = types.StringValue(app.Name)
	data.AppName = types.StringValue(app.AppName)
	data.EnvironmentID = types.StringValue(app.EnvironmentID)
	data.SourceType = types.StringValue(app.SourceType)
	data.BuildType = types.StringValue(app.BuildType)
	data.AutoDeploy = types.BoolValue(app.AutoDeploy)
	data.Replicas = types.Int64Value(int64(app.Replicas))
	data.ApplicationStatus = types.StringValue(app.ApplicationStatus)
	data.CreatedAt = types.StringValue(app.CreatedAt)

	// Optional fields
	if app.Description != "" {
		data.Description = types.StringValue(app.Description)
	}
	if app.ServerID != "" {
		data.ServerID = types.StringValue(app.ServerID)
	}
	if app.DockerImage != "" {
		data.DockerImage = types.StringValue(app.DockerImage)
	}
	if app.RegistryUrl != "" {
		data.RegistryUrl = types.StringValue(app.RegistryUrl)
	}
	if app.CustomGitUrl != "" {
		data.CustomGitUrl = types.StringValue(app.CustomGitUrl)
	}
	if app.CustomGitBranch != "" {
		data.CustomGitBranch = types.StringValue(app.CustomGitBranch)
	}
	if app.Repository != "" {
		data.Repository = types.StringValue(app.Repository)
	}
	if app.Branch != "" {
		data.Branch = types.StringValue(app.Branch)
	}
	if app.Owner != "" {
		data.Owner = types.StringValue(app.Owner)
	}

	// Read traefik config
	traefikConfig, err := d.client.ReadTraefikConfig(data.ID.ValueString())
	if err == nil && traefikConfig != "" {
		data.TraefikConfig = types.StringValue(traefikConfig)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
