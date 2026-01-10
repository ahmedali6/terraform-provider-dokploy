package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DeploymentsDataSource{}

func NewDeploymentsDataSource() datasource.DataSource {
	return &DeploymentsDataSource{}
}

type DeploymentsDataSource struct {
	client *client.DokployClient
}

type DeploymentsDataSourceModel struct {
	// Filter parameters - exactly one must be specified
	ApplicationID types.String `tfsdk:"application_id"`
	ComposeID     types.String `tfsdk:"compose_id"`
	ServerID      types.String `tfsdk:"server_id"`

	// Results
	Deployments []DeploymentDataModel `tfsdk:"deployments"`
}

type DeploymentDataModel struct {
	ID                  types.String `tfsdk:"id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	Status              types.String `tfsdk:"status"`
	LogPath             types.String `tfsdk:"log_path"`
	ApplicationID       types.String `tfsdk:"application_id"`
	ComposeID           types.String `tfsdk:"compose_id"`
	ServerID            types.String `tfsdk:"server_id"`
	IsPreviewDeployment types.Bool   `tfsdk:"is_preview_deployment"`
	PreviewDeploymentID types.String `tfsdk:"preview_deployment_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	StartedAt           types.String `tfsdk:"started_at"`
	FinishedAt          types.String `tfsdk:"finished_at"`
	ErrorMessage        types.String `tfsdk:"error_message"`
	ScheduleID          types.String `tfsdk:"schedule_id"`
	BackupID            types.String `tfsdk:"backup_id"`
	RollbackID          types.String `tfsdk:"rollback_id"`
	VolumeBackupID      types.String `tfsdk:"volume_backup_id"`
	BuildServerID       types.String `tfsdk:"build_server_id"`
}

func (d *DeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *DeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches deployment history from Dokploy. Specify exactly one of application_id, compose_id, or server_id to filter deployments.",
		Attributes: map[string]schema.Attribute{
			// Filter parameters
			"application_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter deployments by application ID.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("application_id"),
						path.MatchRoot("compose_id"),
						path.MatchRoot("server_id"),
					),
				},
			},
			"compose_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter deployments by compose ID.",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter deployments by server ID.",
			},

			// Results
			"deployments": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of deployments.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the deployment.",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "Title of the deployment (e.g., 'Manual deployment').",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Description of the deployment.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Status of the deployment: running, done, error.",
						},
						"log_path": schema.StringAttribute{
							Computed:    true,
							Description: "Path to the deployment log file.",
						},
						"application_id": schema.StringAttribute{
							Computed:    true,
							Description: "Application ID if this is an application deployment.",
						},
						"compose_id": schema.StringAttribute{
							Computed:    true,
							Description: "Compose ID if this is a compose deployment.",
						},
						"server_id": schema.StringAttribute{
							Computed:    true,
							Description: "Server ID if this is a server deployment.",
						},
						"is_preview_deployment": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is a preview deployment.",
						},
						"preview_deployment_id": schema.StringAttribute{
							Computed:    true,
							Description: "Preview deployment ID if applicable.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the deployment was created.",
						},
						"started_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the deployment started.",
						},
						"finished_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the deployment finished (null if still running).",
						},
						"error_message": schema.StringAttribute{
							Computed:    true,
							Description: "Error message if the deployment failed.",
						},
						"schedule_id": schema.StringAttribute{
							Computed:    true,
							Description: "Schedule ID if this was a scheduled deployment.",
						},
						"backup_id": schema.StringAttribute{
							Computed:    true,
							Description: "Backup ID if this is a backup deployment.",
						},
						"rollback_id": schema.StringAttribute{
							Computed:    true,
							Description: "Rollback ID if this is a rollback deployment.",
						},
						"volume_backup_id": schema.StringAttribute{
							Computed:    true,
							Description: "Volume backup ID if this is a volume backup deployment.",
						},
						"build_server_id": schema.StringAttribute{
							Computed:    true,
							Description: "Build server ID if a remote build server was used.",
						},
					},
				},
			},
		},
	}
}

func (d *DeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentsDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deployments []client.Deployment
	var err error

	// Determine which filter to use
	if !data.ApplicationID.IsNull() && data.ApplicationID.ValueString() != "" {
		deployments, err = d.client.ListApplicationDeployments(data.ApplicationID.ValueString())
	} else if !data.ComposeID.IsNull() && data.ComposeID.ValueString() != "" {
		deployments, err = d.client.ListComposeDeployments(data.ComposeID.ValueString())
	} else if !data.ServerID.IsNull() && data.ServerID.ValueString() != "" {
		deployments, err = d.client.ListServerDeployments(data.ServerID.ValueString())
	} else {
		resp.Diagnostics.AddError("Missing Filter", "Exactly one of application_id, compose_id, or server_id must be specified.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to List Deployments", err.Error())
		return
	}

	data.Deployments = make([]DeploymentDataModel, len(deployments))
	for i, dep := range deployments {
		data.Deployments[i] = DeploymentDataModel{
			ID:                  types.StringValue(dep.ID),
			Title:               types.StringValue(dep.Title),
			Description:         types.StringValue(dep.Description),
			Status:              types.StringValue(dep.Status),
			LogPath:             types.StringValue(dep.LogPath),
			IsPreviewDeployment: types.BoolValue(dep.IsPreviewDeployment),
			CreatedAt:           types.StringValue(dep.CreatedAt),
			StartedAt:           types.StringValue(dep.StartedAt),
		}

		// Handle nullable pointer fields
		if dep.ApplicationID != nil {
			data.Deployments[i].ApplicationID = types.StringValue(*dep.ApplicationID)
		}
		if dep.ComposeID != nil {
			data.Deployments[i].ComposeID = types.StringValue(*dep.ComposeID)
		}
		if dep.ServerID != nil {
			data.Deployments[i].ServerID = types.StringValue(*dep.ServerID)
		}
		if dep.PreviewDeploymentID != nil {
			data.Deployments[i].PreviewDeploymentID = types.StringValue(*dep.PreviewDeploymentID)
		}
		if dep.FinishedAt != nil {
			data.Deployments[i].FinishedAt = types.StringValue(*dep.FinishedAt)
		}
		if dep.ErrorMessage != nil {
			data.Deployments[i].ErrorMessage = types.StringValue(*dep.ErrorMessage)
		}
		if dep.ScheduleID != nil {
			data.Deployments[i].ScheduleID = types.StringValue(*dep.ScheduleID)
		}
		if dep.BackupID != nil {
			data.Deployments[i].BackupID = types.StringValue(*dep.BackupID)
		}
		if dep.RollbackID != nil {
			data.Deployments[i].RollbackID = types.StringValue(*dep.RollbackID)
		}
		if dep.VolumeBackupID != nil {
			data.Deployments[i].VolumeBackupID = types.StringValue(*dep.VolumeBackupID)
		}
		if dep.BuildServerID != nil {
			data.Deployments[i].BuildServerID = types.StringValue(*dep.BuildServerID)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
