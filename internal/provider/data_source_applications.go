package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ApplicationsDataSource{}

func NewApplicationsDataSource() datasource.DataSource {
	return &ApplicationsDataSource{}
}

type ApplicationsDataSource struct {
	client *client.DokployClient
}

type ApplicationsDataSourceModel struct {
	EnvironmentID types.String           `tfsdk:"environment_id"`
	Applications  []ApplicationDataModel `tfsdk:"applications"`
}

type ApplicationDataModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	AppName           types.String `tfsdk:"app_name"`
	Description       types.String `tfsdk:"description"`
	EnvironmentID     types.String `tfsdk:"environment_id"`
	ServerID          types.String `tfsdk:"server_id"`
	SourceType        types.String `tfsdk:"source_type"`
	BuildType         types.String `tfsdk:"build_type"`
	AutoDeploy        types.Bool   `tfsdk:"auto_deploy"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	ApplicationStatus types.String `tfsdk:"application_status"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (d *ApplicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

func (d *ApplicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Dokploy applications, optionally filtered by environment.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional environment ID to filter applications. If not provided, returns all applications across all environments.",
			},
			"applications": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of applications.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
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
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the application was created.",
						},
					},
				},
			},
		},
	}
}

func (d *ApplicationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationsDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apps []client.Application
	var err error

	// If environment_id is provided, filter by environment
	if !data.EnvironmentID.IsNull() && !data.EnvironmentID.IsUnknown() && data.EnvironmentID.ValueString() != "" {
		apps, err = d.client.ListApplicationsByEnvironment(data.EnvironmentID.ValueString())
	} else {
		apps, err = d.client.ListApplications()
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to List Applications", err.Error())
		return
	}

	data.Applications = make([]ApplicationDataModel, len(apps))
	for i, app := range apps {
		data.Applications[i] = ApplicationDataModel{
			ID:                types.StringValue(app.ID),
			Name:              types.StringValue(app.Name),
			AppName:           types.StringValue(app.AppName),
			EnvironmentID:     types.StringValue(app.EnvironmentID),
			SourceType:        types.StringValue(app.SourceType),
			BuildType:         types.StringValue(app.BuildType),
			AutoDeploy:        types.BoolValue(app.AutoDeploy),
			Replicas:          types.Int64Value(int64(app.Replicas)),
			ApplicationStatus: types.StringValue(app.ApplicationStatus),
			CreatedAt:         types.StringValue(app.CreatedAt),
		}

		// Optional fields
		if app.Description != "" {
			data.Applications[i].Description = types.StringValue(app.Description)
		}
		if app.ServerID != "" {
			data.Applications[i].ServerID = types.StringValue(app.ServerID)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
