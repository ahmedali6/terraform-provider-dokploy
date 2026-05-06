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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	BackupType      types.String `tfsdk:"backup_type"`
	DatabaseID      types.String `tfsdk:"database_id"`
	DatabaseType    types.String `tfsdk:"database_type"`
	ComposeID       types.String `tfsdk:"compose_id"`
	ServiceName     types.String `tfsdk:"service_name"`
	Schedule        types.String `tfsdk:"schedule"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Prefix          types.String `tfsdk:"prefix"`
	Database        types.String `tfsdk:"database"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
	Metadata        types.Map    `tfsdk:"metadata"`
}

func (r *BackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *BackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages automated backups in Dokploy for databases and compose services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the backup.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the backup destination (S3, MinIO, etc.).",
			},
			"backup_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("database"),
				Description: "Type of backup: 'database' for database backups or 'compose' for compose service backups.",
				Validators: []validator.String{
					stringvalidator.OneOf("database", "compose"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the database to backup. Required when backup_type is 'database'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_type": schema.StringAttribute{
				Optional:    true,
				Description: "Type of database: postgres, mysql, mariadb, or mongo. Required when backup_type is 'database'.",
				Validators: []validator.String{
					stringvalidator.OneOf("postgres", "mysql", "mariadb", "mongo"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the compose to backup. Required when backup_type is 'compose'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the service within the compose to backup. Required when backup_type is 'compose'.",
			},
			"schedule": schema.StringAttribute{
				Required:    true,
				Description: "Cron schedule for backups (e.g., '0 2 * * *' for daily at 2 AM).",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the backup schedule is enabled.",
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "Prefix for backup files.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Database name to backup (for database backups) or identifier (for compose backups).",
			},
			"keep_latest_count": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of recent backups to keep (older ones are deleted).",
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Metadata for the backup configuration as key-value pairs.",
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

	backupType := plan.BackupType.ValueString()
	if backupType == "" {
		backupType = "database"
	}

	// Validate required fields based on backup_type
	switch backupType {
	case "database":
		if plan.DatabaseID.IsNull() || plan.DatabaseID.ValueString() == "" {
			resp.Diagnostics.AddError("Missing required field", "database_id is required when backup_type is 'database'")
			return
		}
		if plan.DatabaseType.IsNull() || plan.DatabaseType.ValueString() == "" {
			resp.Diagnostics.AddError("Missing required field", "database_type is required when backup_type is 'database'")
			return
		}
	case "compose":
		if plan.ComposeID.IsNull() || plan.ComposeID.ValueString() == "" {
			resp.Diagnostics.AddError("Missing required field", "compose_id is required when backup_type is 'compose'")
			return
		}
		if plan.ServiceName.IsNull() || plan.ServiceName.ValueString() == "" {
			resp.Diagnostics.AddError("Missing required field", "service_name is required when backup_type is 'compose'")
			return
		}
	}

	backup := client.Backup{
		DestinationID:   plan.DestinationID.ValueString(),
		Schedule:        plan.Schedule.ValueString(),
		Enabled:         plan.Enabled.ValueBool(),
		Prefix:          plan.Prefix.ValueString(),
		Database:        plan.Database.ValueString(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
		BackupType:      backupType,
	}

	// Set metadata if provided
	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		metadataMap := make(map[string]string)
		diags = plan.Metadata.ElementsAs(ctx, &metadataMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		backup.Metadata = metadataMap
	}

	switch backupType {
	case "database":
		backup.DatabaseType = plan.DatabaseType.ValueString()
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
		}
	case "compose":
		backup.ComposeID = plan.ComposeID.ValueString()
		backup.ServiceName = plan.ServiceName.ValueString()
		// Compose backups still require databaseType field in API (use postgres as default)
		if !plan.DatabaseType.IsNull() && plan.DatabaseType.ValueString() != "" {
			backup.DatabaseType = plan.DatabaseType.ValueString()
		} else {
			backup.DatabaseType = "postgres"
		}
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
	plan.BackupType = types.StringValue(createdBackup.BackupType)

	if createdBackup.ServiceName != "" {
		plan.ServiceName = types.StringValue(createdBackup.ServiceName)
	}

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
	state.BackupType = types.StringValue(backup.BackupType)

	// Set database_type for both database and compose backups (API returns it for both)
	if backup.DatabaseType != "" {
		state.DatabaseType = types.StringValue(backup.DatabaseType)
	}

	switch backup.BackupType {
	case "database":
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
	case "compose":
		state.ComposeID = types.StringValue(backup.ComposeID)
		if backup.ServiceName != "" {
			state.ServiceName = types.StringValue(backup.ServiceName)
		}
	}

	// Read metadata if present
	if backup.Metadata != nil && len(backup.Metadata) > 0 {
		state.Metadata, diags = types.MapValueFrom(ctx, types.StringType, backup.Metadata)
		resp.Diagnostics.Append(diags...)
	} else {
		state.Metadata = types.MapNull(types.StringType)
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

	backupType := plan.BackupType.ValueString()
	if backupType == "" {
		backupType = "database"
	}

	backup := client.Backup{
		BackupID:        plan.ID.ValueString(),
		DestinationID:   plan.DestinationID.ValueString(),
		Schedule:        plan.Schedule.ValueString(),
		Enabled:         plan.Enabled.ValueBool(),
		Prefix:          plan.Prefix.ValueString(),
		Database:        plan.Database.ValueString(),
		KeepLatestCount: int(plan.KeepLatestCount.ValueInt64()),
	}

	// Set metadata if provided
	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		metadataMap := make(map[string]string)
		diags = plan.Metadata.ElementsAs(ctx, &metadataMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		backup.Metadata = metadataMap
	}

	// Set database type for the update API
	if !plan.DatabaseType.IsNull() && plan.DatabaseType.ValueString() != "" {
		backup.DatabaseType = plan.DatabaseType.ValueString()
	} else {
		backup.DatabaseType = "postgres" // Default for compose backups
	}

	// Set service name for compose backups
	if backupType == "compose" && !plan.ServiceName.IsNull() {
		backup.ServiceName = plan.ServiceName.ValueString()
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

	if updatedBackup.ServiceName != "" {
		plan.ServiceName = types.StringValue(updatedBackup.ServiceName)
	}

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
