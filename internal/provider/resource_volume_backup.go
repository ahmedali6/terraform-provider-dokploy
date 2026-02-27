package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &VolumeBackupResource{}
var _ resource.ResourceWithImportState = &VolumeBackupResource{}

func NewVolumeBackupResource() resource.Resource {
	return &VolumeBackupResource{}
}

type VolumeBackupResource struct {
	client *client.DokployClient
}

type VolumeBackupResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	VolumeName      types.String `tfsdk:"volume_name"`
	Prefix          types.String `tfsdk:"prefix"`
	DestinationID   types.String `tfsdk:"destination_id"`
	CronExpression  types.String `tfsdk:"cron_expression"`
	ServiceType     types.String `tfsdk:"service_type"`
	ServiceID       types.String `tfsdk:"service_id"`
	AppName         types.String `tfsdk:"app_name"`
	ServiceName     types.String `tfsdk:"service_name"`
	TurnOff         types.Bool   `tfsdk:"turn_off"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (r *VolumeBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_backup"
}

func (r *VolumeBackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages volume backups in Dokploy for backing up Docker volumes from applications, databases, and compose services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the volume backup.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the volume backup configuration.",
			},
			"volume_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Docker volume to backup.",
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "Prefix for backup files in the destination storage.",
			},
			"destination_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the backup destination (S3, MinIO, etc.).",
			},
			"cron_expression": schema.StringAttribute{
				Required:    true,
				Description: "Cron schedule for backups (e.g., '0 3 * * *' for daily at 3 AM).",
			},
			"service_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of service to backup: application, postgres, mysql, mariadb, mongo, redis, or compose.",
				Validators: []validator.String{
					stringvalidator.OneOf("application", "postgres", "mysql", "mariadb", "mongo", "redis", "compose"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service to backup (application_id, postgres_id, mysql_id, etc. based on service_type).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "Docker app/service name (the appName field from the service resource).",
			},
			"service_name": schema.StringAttribute{
				Optional:    true,
				Description: "Service name within a compose stack. Required when service_type is 'compose'.",
			},
			"turn_off": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to stop the service during backup for data consistency. Default: false.",
			},
			"keep_latest_count": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(5),
				Description: "Number of recent backups to keep. Older backups are automatically deleted. Default: 5.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the backup schedule is enabled. Default: true.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the volume backup was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VolumeBackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *VolumeBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VolumeBackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate compose service_name requirement
	if plan.ServiceType.ValueString() == "compose" && (plan.ServiceName.IsNull() || plan.ServiceName.ValueString() == "") {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"service_name is required when service_type is 'compose'",
		)
		return
	}

	backup := client.VolumeBackup{
		Name:            plan.Name.ValueString(),
		VolumeName:      plan.VolumeName.ValueString(),
		Prefix:          plan.Prefix.ValueString(),
		DestinationID:   plan.DestinationID.ValueString(),
		CronExpression:  plan.CronExpression.ValueString(),
		ServiceType:     plan.ServiceType.ValueString(),
		AppName:         plan.AppName.ValueString(),
		TurnOff:         plan.TurnOff.ValueBool(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
		Enabled:         plan.Enabled.ValueBool(),
	}

	if !plan.ServiceName.IsNull() && plan.ServiceName.ValueString() != "" {
		serviceName := plan.ServiceName.ValueString()
		backup.ServiceName = &serviceName
	}

	// Map service_id to the appropriate field based on service_type
	serviceID := plan.ServiceID.ValueString()
	switch plan.ServiceType.ValueString() {
	case "application":
		backup.ApplicationID = &serviceID
	case "postgres":
		backup.PostgresID = &serviceID
	case "mysql":
		backup.MysqlID = &serviceID
	case "mariadb":
		backup.MariadbID = &serviceID
	case "mongo":
		backup.MongoID = &serviceID
	case "redis":
		backup.RedisID = &serviceID
	case "compose":
		backup.ComposeID = &serviceID
	}

	created, err := r.client.CreateVolumeBackup(backup)
	if err != nil {
		resp.Diagnostics.AddError("Error creating volume backup", err.Error())
		return
	}

	plan.ID = types.StringValue(created.VolumeBackupID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.TurnOff = types.BoolValue(created.TurnOff)
	plan.KeepLatestCount = types.Int64Value(int64(created.KeepLatestCount))
	plan.Enabled = types.BoolValue(created.Enabled)

	if created.ServiceName != nil && *created.ServiceName != "" {
		plan.ServiceName = types.StringValue(*created.ServiceName)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeBackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetVolumeBackup(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading volume backup", err.Error())
		return
	}

	state.Name = types.StringValue(backup.Name)
	state.VolumeName = types.StringValue(backup.VolumeName)
	state.Prefix = types.StringValue(backup.Prefix)
	state.DestinationID = types.StringValue(backup.DestinationID)
	state.CronExpression = types.StringValue(backup.CronExpression)
	state.ServiceType = types.StringValue(backup.ServiceType)
	state.AppName = types.StringValue(backup.AppName)
	state.TurnOff = types.BoolValue(backup.TurnOff)
	state.KeepLatestCount = types.Int64Value(int64(backup.KeepLatestCount))
	state.Enabled = types.BoolValue(backup.Enabled)
	state.CreatedAt = types.StringValue(backup.CreatedAt)

	// Extract service_id from the appropriate field
	switch backup.ServiceType {
	case "application":
		if backup.ApplicationID != nil {
			state.ServiceID = types.StringValue(*backup.ApplicationID)
		}
	case "postgres":
		if backup.PostgresID != nil {
			state.ServiceID = types.StringValue(*backup.PostgresID)
		}
	case "mysql":
		if backup.MysqlID != nil {
			state.ServiceID = types.StringValue(*backup.MysqlID)
		}
	case "mariadb":
		if backup.MariadbID != nil {
			state.ServiceID = types.StringValue(*backup.MariadbID)
		}
	case "mongo":
		if backup.MongoID != nil {
			state.ServiceID = types.StringValue(*backup.MongoID)
		}
	case "redis":
		if backup.RedisID != nil {
			state.ServiceID = types.StringValue(*backup.RedisID)
		}
	case "compose":
		if backup.ComposeID != nil {
			state.ServiceID = types.StringValue(*backup.ComposeID)
		}
	}

	if backup.ServiceName != nil && *backup.ServiceName != "" {
		state.ServiceName = types.StringValue(*backup.ServiceName)
	} else {
		state.ServiceName = types.StringNull()
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeBackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VolumeBackupResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup := client.VolumeBackup{
		VolumeBackupID:  state.ID.ValueString(),
		Name:            plan.Name.ValueString(),
		VolumeName:      plan.VolumeName.ValueString(),
		Prefix:          plan.Prefix.ValueString(),
		DestinationID:   plan.DestinationID.ValueString(),
		CronExpression:  plan.CronExpression.ValueString(),
		ServiceType:     state.ServiceType.ValueString(),
		AppName:         state.AppName.ValueString(),
		TurnOff:         plan.TurnOff.ValueBool(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
		Enabled:         plan.Enabled.ValueBool(),
	}

	if !plan.ServiceName.IsNull() && plan.ServiceName.ValueString() != "" {
		serviceName := plan.ServiceName.ValueString()
		backup.ServiceName = &serviceName
	}

	// Pass the service ID in the correct field for the API
	serviceID := state.ServiceID.ValueString()
	switch state.ServiceType.ValueString() {
	case "application":
		backup.ApplicationID = &serviceID
	case "postgres":
		backup.PostgresID = &serviceID
	case "mysql":
		backup.MysqlID = &serviceID
	case "mariadb":
		backup.MariadbID = &serviceID
	case "mongo":
		backup.MongoID = &serviceID
	case "redis":
		backup.RedisID = &serviceID
	case "compose":
		backup.ComposeID = &serviceID
	default:
		resp.Diagnostics.AddError(
			"Invalid service_type",
			fmt.Sprintf("Unknown service_type %q. Must be one of: application, postgres, mysql, mariadb, mongo, redis, compose", state.ServiceType.ValueString()),
		)
		return
	}

	updated, err := r.client.UpdateVolumeBackup(backup)
	if err != nil {
		resp.Diagnostics.AddError("Error updating volume backup", err.Error())
		return
	}

	plan.ID = state.ID
	plan.CreatedAt = state.CreatedAt
	plan.Name = types.StringValue(updated.Name)
	plan.VolumeName = types.StringValue(updated.VolumeName)
	plan.Prefix = types.StringValue(updated.Prefix)
	plan.CronExpression = types.StringValue(updated.CronExpression)
	plan.TurnOff = types.BoolValue(updated.TurnOff)
	plan.KeepLatestCount = types.Int64Value(int64(updated.KeepLatestCount))
	plan.Enabled = types.BoolValue(updated.Enabled)

	if updated.ServiceName != nil && *updated.ServiceName != "" {
		plan.ServiceName = types.StringValue(*updated.ServiceName)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeBackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVolumeBackup(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting volume backup", err.Error())
		return
	}
}

func (r *VolumeBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
