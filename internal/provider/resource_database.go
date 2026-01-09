package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	client *client.DokployClient
}

type DatabaseResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	Password        types.String `tfsdk:"password"`
	Version         types.String `tfsdk:"version"`
	InternalPort    types.Int64  `tfsdk:"internal_port"`
	ExternalPort    types.Int64  `tfsdk:"external_port"`
	ApplicationName types.String `tfsdk:"app_name"`
}

func (r *DatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_port": schema.Int64Attribute{
				Computed: true,
			},
			"external_port": schema.Int64Attribute{
				Computed: true,
			},
			"app_name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *DatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct dockerImage from type and version
	dockerImage := plan.Type.ValueString()
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() && plan.Version.ValueString() != "" {
		dockerImage = fmt.Sprintf("%s:%s", plan.Type.ValueString(), plan.Version.ValueString())
	}

	db, err := r.client.CreateDatabase(
		plan.ProjectID.ValueString(),
		plan.EnvironmentID.ValueString(),
		plan.Name.ValueString(),
		plan.Type.ValueString(),
		plan.Password.ValueString(),
		dockerImage,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
		return
	}

	plan.ID = types.StringValue(db.ID)
	plan.InternalPort = types.Int64Value(db.InternalPort)
	plan.ExternalPort = types.Int64Value(db.ExternalPort)
	plan.ApplicationName = types.StringValue(db.AppName)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.GetDatabase(state.ID.ValueString(), state.Type.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database", err.Error())
		return
	}

	state.Name = types.StringValue(db.Name)
	state.Type = types.StringValue(db.Type)
	if db.ProjectID != "" {
		state.ProjectID = types.StringValue(db.ProjectID)
	}
	if db.EnvironmentID != "" {
		state.EnvironmentID = types.StringValue(db.EnvironmentID)
	}
	if db.Version != "" {
		state.Version = types.StringValue(db.Version)
	}
	// InternalPort/ExternalPort mapping
	state.InternalPort = types.Int64Value(db.InternalPort)
	state.ExternalPort = types.Int64Value(db.ExternalPort)
	state.ApplicationName = types.StringValue(db.AppName)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update support
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDatabaseWithType(state.ID.ValueString(), state.Type.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting database", err.Error())
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <type>:<id> or just <id> (will try to detect type)
	importID := req.ID

	var dbType, dbID string

	// Check if format is type:id
	if strings.Contains(importID, ":") {
		parts := strings.SplitN(importID, ":", 2)
		if len(parts) == 2 {
			dbType = parts[0]
			dbID = parts[1]
		}
	} else {
		// Try to detect database type by querying each type
		dbID = importID
		types := []string{"postgres", "mysql", "mariadb", "mongo", "redis"}

		for _, t := range types {
			db, err := r.client.GetDatabase(dbID, t)
			if err == nil && db != nil && db.ID != "" {
				dbType = t
				break
			}
		}

		if dbType == "" {
			resp.Diagnostics.AddError(
				"Unable to detect database type",
				fmt.Sprintf("Could not determine database type for ID %s. Please use format <type>:<id> for import (e.g., postgres:%s)", dbID, dbID),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dbID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), dbType)...)
}
