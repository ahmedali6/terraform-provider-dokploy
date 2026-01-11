package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DockerContainersDataSource{}

func NewDockerContainersDataSource() datasource.DataSource {
	return &DockerContainersDataSource{}
}

type DockerContainersDataSource struct {
	client *client.DokployClient
}

type DockerContainersDataSourceModel struct {
	ServerID   types.String           `tfsdk:"server_id"`
	AppName    types.String           `tfsdk:"app_name"`
	AppType    types.String           `tfsdk:"app_type"`
	LabelType  types.String           `tfsdk:"label_type"`
	Containers []DockerContainerModel `tfsdk:"containers"`
}

type DockerContainerModel struct {
	ContainerID types.String `tfsdk:"container_id"`
	Name        types.String `tfsdk:"name"`
	Image       types.String `tfsdk:"image"`
	Ports       types.String `tfsdk:"ports"`
	State       types.String `tfsdk:"state"`
	Status      types.String `tfsdk:"status"`
}

func (d *DockerContainersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_containers"
}

func (d *DockerContainersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Docker containers from Dokploy. Supports filtering by server, app name, and labels. Note: When using app_name with app_type or label_type filters, only container_id, name, and state fields are returned; image, ports, and status will be null.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter containers by remote server ID. Leave empty for the local server.",
			},
			"app_name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter containers by Dokploy app name pattern. Used with app_type or label_type filters.",
			},
			"app_type": schema.StringAttribute{
				Optional:    true,
				Description: "App type filter when using app_name: 'application' or 'compose'. Triggers name pattern matching. Cannot be used together with label_type.",
			},
			"label_type": schema.StringAttribute{
				Optional:    true,
				Description: "Label type filter when using app_name: 'standalone' or 'swarm'. Triggers label-based filtering. Cannot be used together with app_type.",
			},
			"containers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Docker containers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_id": schema.StringAttribute{
							Computed:    true,
							Description: "Short container ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Container name.",
						},
						"image": schema.StringAttribute{
							Computed:    true,
							Description: "Container image. Note: Only available when listing all containers (no app_name filter).",
						},
						"ports": schema.StringAttribute{
							Computed:    true,
							Description: "Exposed ports. Note: Only available when listing all containers (no app_name filter).",
						},
						"state": schema.StringAttribute{
							Computed:    true,
							Description: "Container state (running, exited, etc.).",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable status (e.g., 'Up 5 hours'). Note: Only available when listing all containers (no app_name filter).",
						},
					},
				},
			},
		},
	}
}

func (d *DockerContainersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DockerContainersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DockerContainersDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := ""
	if !data.ServerID.IsNull() {
		serverID = data.ServerID.ValueString()
	}

	appName := ""
	if !data.AppName.IsNull() {
		appName = data.AppName.ValueString()
	}

	appType := ""
	if !data.AppType.IsNull() {
		appType = data.AppType.ValueString()
	}

	labelType := ""
	if !data.LabelType.IsNull() {
		labelType = data.LabelType.ValueString()
	}

	// Validate conflicting filter options
	if appType != "" && labelType != "" {
		resp.Diagnostics.AddError(
			"Conflicting Filter Options",
			"Cannot specify both app_type and label_type. Use app_type for name pattern matching or label_type for label-based filtering, but not both.",
		)
		return
	}

	// Warn if app_type or label_type is set without app_name
	if appName == "" && (appType != "" || labelType != "") {
		resp.Diagnostics.AddWarning(
			"Filter Option Ignored",
			"app_type and label_type require app_name to be set. These options will be ignored.",
		)
	}

	// Determine which API to call based on filters
	if appName != "" && labelType != "" {
		// Use label-based filtering
		containers, err := d.client.ListDockerContainersByAppLabel(appName, labelType, serverID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to List Docker Containers by Label", err.Error())
			return
		}
		data.Containers = make([]DockerContainerModel, len(containers))
		for i, c := range containers {
			data.Containers[i] = DockerContainerModel{
				ContainerID: types.StringValue(c.ContainerID),
				Name:        types.StringValue(c.Name),
				Image:       types.StringNull(),
				Ports:       types.StringNull(),
				State:       types.StringValue(c.State),
				Status:      types.StringNull(),
			}
		}
	} else if appName != "" {
		// Use name pattern matching
		containers, err := d.client.ListDockerContainersByAppNameMatch(appName, appType, serverID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to List Docker Containers by App Name", err.Error())
			return
		}
		data.Containers = make([]DockerContainerModel, len(containers))
		for i, c := range containers {
			data.Containers[i] = DockerContainerModel{
				ContainerID: types.StringValue(c.ContainerID),
				Name:        types.StringValue(c.Name),
				Image:       types.StringNull(),
				Ports:       types.StringNull(),
				State:       types.StringValue(c.State),
				Status:      types.StringNull(),
			}
		}
	} else {
		// List all containers
		containers, err := d.client.ListDockerContainers(serverID)
		if err != nil {
			resp.Diagnostics.AddError("Unable to List Docker Containers", err.Error())
			return
		}
		data.Containers = make([]DockerContainerModel, len(containers))
		for i, c := range containers {
			data.Containers[i] = DockerContainerModel{
				ContainerID: types.StringValue(c.ContainerID),
				Name:        types.StringValue(c.Name),
				Image:       types.StringValue(c.Image),
				Ports:       types.StringValue(c.Ports),
				State:       types.StringValue(c.State),
				Status:      types.StringValue(c.Status),
			}
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
