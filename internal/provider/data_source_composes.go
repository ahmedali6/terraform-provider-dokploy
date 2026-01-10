package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ComposesDataSource{}

func NewComposesDataSource() datasource.DataSource {
	return &ComposesDataSource{}
}

type ComposesDataSource struct {
	client *client.DokployClient
}

type ComposesDataSourceModel struct {
	EnvironmentID types.String       `tfsdk:"environment_id"`
	Composes      []ComposeDataModel `tfsdk:"composes"`
}

type ComposeDataModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ServerID      types.String `tfsdk:"server_id"`

	// Compose file
	ComposePath types.String `tfsdk:"compose_path"`
	ComposeType types.String `tfsdk:"compose_type"`

	// Source configuration
	SourceType types.String `tfsdk:"source_type"`

	// Runtime configuration
	AutoDeploy                types.Bool `tfsdk:"auto_deploy"`
	Randomize                 types.Bool `tfsdk:"randomize"`
	IsolatedDeployment        types.Bool `tfsdk:"isolated_deployment"`
	IsolatedDeploymentsVolume types.Bool `tfsdk:"isolated_deployments_volume"`

	// Status
	ComposeStatus types.String `tfsdk:"compose_status"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (d *ComposesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_composes"
}

func (d *ComposesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Dokploy compose stacks, optionally filtered by environment.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional environment ID to filter compose stacks. If not provided, returns all compose stacks across all environments.",
			},
			"composes": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of compose stacks.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
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
						"compose_path": schema.StringAttribute{
							Computed:    true,
							Description: "Path to the docker-compose.yml file in the repository.",
						},
						"compose_type": schema.StringAttribute{
							Computed:    true,
							Description: "The compose type: 'docker-compose' or 'stack' for Docker Swarm.",
						},
						"source_type": schema.StringAttribute{
							Computed:    true,
							Description: "The source type: github, gitlab, bitbucket, gitea, git, or raw.",
						},
						"auto_deploy": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether auto-deploy is enabled.",
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
						"compose_status": schema.StringAttribute{
							Computed:    true,
							Description: "Current status: idle, running, done, or error.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the compose stack was created.",
						},
					},
				},
			},
		},
	}
}

func (d *ComposesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ComposesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ComposesDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environmentID string
	if !data.EnvironmentID.IsNull() && !data.EnvironmentID.IsUnknown() {
		environmentID = data.EnvironmentID.ValueString()
	}

	composes, err := d.client.ListComposes(environmentID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Composes", err.Error())
		return
	}

	data.Composes = make([]ComposeDataModel, len(composes))
	for i, comp := range composes {
		data.Composes[i] = ComposeDataModel{
			ID:                        types.StringValue(comp.ID),
			Name:                      types.StringValue(comp.Name),
			EnvironmentID:             types.StringValue(comp.EnvironmentID),
			SourceType:                types.StringValue(comp.SourceType),
			AutoDeploy:                types.BoolValue(comp.AutoDeploy),
			Randomize:                 types.BoolValue(comp.Randomize),
			IsolatedDeployment:        types.BoolValue(comp.IsolatedDeployment),
			IsolatedDeploymentsVolume: types.BoolValue(comp.IsolatedDeploymentsVolume),
		}

		// Optional string fields
		if comp.AppName != "" {
			data.Composes[i].AppName = types.StringValue(comp.AppName)
		}
		if comp.Description != "" {
			data.Composes[i].Description = types.StringValue(comp.Description)
		}
		if comp.ServerID != "" {
			data.Composes[i].ServerID = types.StringValue(comp.ServerID)
		}
		if comp.ComposePath != "" {
			data.Composes[i].ComposePath = types.StringValue(comp.ComposePath)
		}
		if comp.ComposeType != "" {
			data.Composes[i].ComposeType = types.StringValue(comp.ComposeType)
		}
		if comp.ComposeStatus != "" {
			data.Composes[i].ComposeStatus = types.StringValue(comp.ComposeStatus)
		}
		if comp.CreatedAt != "" {
			data.Composes[i].CreatedAt = types.StringValue(comp.CreatedAt)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
