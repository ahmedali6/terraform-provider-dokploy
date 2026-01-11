package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DockerContainerDataSource{}

func NewDockerContainerDataSource() datasource.DataSource {
	return &DockerContainerDataSource{}
}

type DockerContainerDataSource struct {
	client *client.DokployClient
}

type DockerContainerDataSourceModel struct {
	ContainerID  types.String `tfsdk:"container_id"`
	ServerID     types.String `tfsdk:"server_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Created      types.String `tfsdk:"created"`
	Path         types.String `tfsdk:"path"`
	Image        types.String `tfsdk:"image"`
	Platform     types.String `tfsdk:"platform"`
	RestartCount types.Int64  `tfsdk:"restart_count"`
	// State fields
	StateStatus     types.String `tfsdk:"state_status"`
	StateRunning    types.Bool   `tfsdk:"state_running"`
	StatePaused     types.Bool   `tfsdk:"state_paused"`
	StateRestarting types.Bool   `tfsdk:"state_restarting"`
	StateDead       types.Bool   `tfsdk:"state_dead"`
	StatePid        types.Int64  `tfsdk:"state_pid"`
	StateExitCode   types.Int64  `tfsdk:"state_exit_code"`
	StateError      types.String `tfsdk:"state_error"`
	StateStartedAt  types.String `tfsdk:"state_started_at"`
	StateFinishedAt types.String `tfsdk:"state_finished_at"`
	// Config fields
	ConfigHostname   types.String `tfsdk:"config_hostname"`
	ConfigUser       types.String `tfsdk:"config_user"`
	ConfigImage      types.String `tfsdk:"config_image"`
	ConfigWorkingDir types.String `tfsdk:"config_working_dir"`
	// Raw JSON output
	ConfigJSON types.String `tfsdk:"config_json"`
}

func (d *DockerContainerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_container"
}

func (d *DockerContainerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches detailed configuration for a specific Docker container from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"container_id": schema.StringAttribute{
				Required:    true,
				Description: "The Docker container ID (short or full).",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID for remote servers. Leave empty for the local server.",
			},
			// Basic info
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Full container ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Container name.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Container creation timestamp.",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "Container path/entrypoint.",
			},
			"image": schema.StringAttribute{
				Computed:    true,
				Description: "Container image hash.",
			},
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "Container platform.",
			},
			"restart_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of times the container has been restarted.",
			},
			// State fields
			"state_status": schema.StringAttribute{
				Computed:    true,
				Description: "Container state status (running, exited, etc.).",
			},
			"state_running": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the container is running.",
			},
			"state_paused": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the container is paused.",
			},
			"state_restarting": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the container is restarting.",
			},
			"state_dead": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the container is dead.",
			},
			"state_pid": schema.Int64Attribute{
				Computed:    true,
				Description: "Container process ID.",
			},
			"state_exit_code": schema.Int64Attribute{
				Computed:    true,
				Description: "Container exit code.",
			},
			"state_error": schema.StringAttribute{
				Computed:    true,
				Description: "Container error message if any.",
			},
			"state_started_at": schema.StringAttribute{
				Computed:    true,
				Description: "Container start timestamp.",
			},
			"state_finished_at": schema.StringAttribute{
				Computed:    true,
				Description: "Container finish timestamp.",
			},
			// Config fields
			"config_hostname": schema.StringAttribute{
				Computed:    true,
				Description: "Container hostname.",
			},
			"config_user": schema.StringAttribute{
				Computed:    true,
				Description: "Container user.",
			},
			"config_image": schema.StringAttribute{
				Computed:    true,
				Description: "Container image name.",
			},
			"config_working_dir": schema.StringAttribute{
				Computed:    true,
				Description: "Container working directory.",
			},
			// Raw JSON
			"config_json": schema.StringAttribute{
				Computed:    true,
				Description: "Full container configuration as raw JSON. Useful for accessing all fields via jsondecode().",
			},
		},
	}
}

func (d *DockerContainerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DockerContainerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DockerContainerDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	containerID := data.ContainerID.ValueString()
	serverID := ""
	if !data.ServerID.IsNull() {
		serverID = data.ServerID.ValueString()
	}

	// Get parsed config
	config, err := d.client.GetDockerContainerConfig(containerID, serverID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Docker Container Config", err.Error())
		return
	}

	// Get raw JSON config
	rawJSON, err := d.client.GetDockerContainerConfigRaw(containerID, serverID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Docker Container Raw Config", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(config.ID)
	data.Name = types.StringValue(config.Name)
	data.Created = types.StringValue(config.Created)
	data.Path = types.StringValue(config.Path)
	data.Image = types.StringValue(config.Image)
	data.Platform = types.StringValue(config.Platform)
	data.RestartCount = types.Int64Value(int64(config.RestartCount))

	// State fields
	data.StateStatus = types.StringValue(config.State.Status)
	data.StateRunning = types.BoolValue(config.State.Running)
	data.StatePaused = types.BoolValue(config.State.Paused)
	data.StateRestarting = types.BoolValue(config.State.Restarting)
	data.StateDead = types.BoolValue(config.State.Dead)
	data.StatePid = types.Int64Value(int64(config.State.Pid))
	data.StateExitCode = types.Int64Value(int64(config.State.ExitCode))
	data.StateError = types.StringValue(config.State.Error)
	data.StateStartedAt = types.StringValue(config.State.StartedAt)
	data.StateFinishedAt = types.StringValue(config.State.FinishedAt)

	// Config fields
	data.ConfigHostname = types.StringValue(config.Config.Hostname)
	data.ConfigUser = types.StringValue(config.Config.User)
	data.ConfigImage = types.StringValue(config.Config.Image)
	data.ConfigWorkingDir = types.StringValue(config.Config.WorkingDir)

	// Raw JSON
	data.ConfigJSON = types.StringValue(rawJSON)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
