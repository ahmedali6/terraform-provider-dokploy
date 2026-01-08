package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BackupResource{}
var _ resource.ResourceWithImportState = &BackupResource{}

func NewBackupResource() resource.Resource {
	return &BackupResource{}
}

type BackupResource struct {
	client *client.DokployClient
}

type BackupResourceModel struct {
	ID              types.String `tfsdk:"id"`
	DestinationID   types.String `tfsdk:"destination_id"`
	DatabaseID      types.String `tfsdk:"database_id"`
	DatabaseType    types.String `tfsdk:"database_type"`
	Schedule        types.String `tfsdk:"schedule"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Prefix          types.String `tfsdk:"prefix"`
	Database        types.String `tfsdk:"database"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
}

func (r *BackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *BackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages automated database backups in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the backup",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the backup destination (S3, MinIO, etc.)",
			},
			"database_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the database to backup",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of database: postgres, mysql, mariadb, or mongo",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule": schema.StringAttribute{
				Required:    true,
				Description: "Cron schedule for backups (e.g., '0 2 * * *' for daily at 2 AM)",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the backup schedule is enabled",
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "Prefix for backup files",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("postgres"),
				Description: "Database name to backup",
			},
			"keep_latest_count": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of recent backups to keep (older ones are deleted)",
			},
		},
	}
}

func (r *BackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup := client.Backup{
		DestinationID:   plan.DestinationID.ValueString(),
		Schedule:        plan.Schedule.ValueString(),
		Enabled:         plan.Enabled.ValueBool(),
		Prefix:          plan.Prefix.ValueString(),
		Database:        plan.Database.ValueString(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
		BackupType:      "database",
		DatabaseType:    plan.DatabaseType.ValueString(),
	}

	// Set the appropriate type-specific database ID based on database_type
	databaseID := plan.DatabaseID.ValueString()
	switch plan.DatabaseType.ValueString() {
	case "postgres":
		backup.PostgresID = databaseID
	case "mysql":
		backup.MysqlID = databaseID
	case "mariadb":
		backup.MariadbID = databaseID
	case "mongo":
		backup.MongoID = databaseID
	default:
		resp.Diagnostics.AddError(
			"Invalid database type",
			fmt.Sprintf("database_type must be one of: postgres, mysql, mariadb, mongo. Got: %s", plan.DatabaseType.ValueString()),
		)
		return
	}

	createdBackup, err := r.client.CreateBackup(backup)
	if err != nil {
		resp.Diagnostics.AddError("Error creating backup", err.Error())
		return
	}

	plan.ID = types.StringValue(createdBackup.BackupID)
	plan.Schedule = types.StringValue(createdBackup.Schedule)
	plan.Enabled = types.BoolValue(createdBackup.Enabled)
	plan.Prefix = types.StringValue(createdBackup.Prefix)
	plan.Database = types.StringValue(createdBackup.Database)
	plan.KeepLatestCount = types.Int64Value(int64(createdBackup.KeepLatestCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetBackup(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading backup", err.Error())
		return
	}

	state.DestinationID = types.StringValue(backup.DestinationID)
	state.Schedule = types.StringValue(backup.Schedule)
	state.Enabled = types.BoolValue(backup.Enabled)
	state.Prefix = types.StringValue(backup.Prefix)
	state.Database = types.StringValue(backup.Database)
	state.KeepLatestCount = types.Int64Value(int64(backup.KeepLatestCount))
	state.DatabaseType = types.StringValue(backup.DatabaseType)

	// Extract database_id from the appropriate type-specific field
	switch backup.DatabaseType {
	case "postgres":
		state.DatabaseID = types.StringValue(backup.PostgresID)
	case "mysql":
		state.DatabaseID = types.StringValue(backup.MysqlID)
	case "mariadb":
		state.DatabaseID = types.StringValue(backup.MariadbID)
	case "mongo":
		state.DatabaseID = types.StringValue(backup.MongoID)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup := client.Backup{
		BackupID:        plan.ID.ValueString(),
		DestinationID:   plan.DestinationID.ValueString(),
		Schedule:        plan.Schedule.ValueString(),
		Enabled:         plan.Enabled.ValueBool(),
		Prefix:          plan.Prefix.ValueString(),
		Database:        plan.Database.ValueString(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
		DatabaseType:    plan.DatabaseType.ValueString(),
	}

	updatedBackup, err := r.client.UpdateBackup(backup)
	if err != nil {
		resp.Diagnostics.AddError("Error updating backup", err.Error())
		return
	}

	plan.Schedule = types.StringValue(updatedBackup.Schedule)
	plan.Enabled = types.BoolValue(updatedBackup.Enabled)
	plan.Prefix = types.StringValue(updatedBackup.Prefix)
	plan.Database = types.StringValue(updatedBackup.Database)
	plan.KeepLatestCount = types.Int64Value(int64(updatedBackup.KeepLatestCount))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBackup(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting backup", err.Error())
		return
	}
}

func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
