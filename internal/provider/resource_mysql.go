package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MySQLResource{}
var _ resource.ResourceWithImportState = &MySQLResource{}

func NewMySQLResource() resource.Resource {
	return &MySQLResource{}
}

type MySQLResource struct {
	client *client.DokployClient
}

type MySQLResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	AppNamePrefix        types.String `tfsdk:"app_name_prefix"`
	AppName              types.String `tfsdk:"app_name"`
	Description          types.String `tfsdk:"description"`
	DatabaseName         types.String `tfsdk:"database_name"`
	DatabaseUser         types.String `tfsdk:"database_user"`
	DatabasePassword     types.String `tfsdk:"database_password"`
	DatabaseRootPassword types.String `tfsdk:"database_root_password"`
	DockerImage          types.String `tfsdk:"docker_image"`
	Command              types.String `tfsdk:"command"`
	Env                  types.String `tfsdk:"env"`
	MemoryReservation    types.String `tfsdk:"memory_reservation"`
	MemoryLimit          types.String `tfsdk:"memory_limit"`
	CPUReservation       types.String `tfsdk:"cpu_reservation"`
	CPULimit             types.String `tfsdk:"cpu_limit"`
	InternalPort         types.Int64  `tfsdk:"internal_port"`
	ExternalPort         types.Int64  `tfsdk:"external_port"`
	EnvironmentID        types.String `tfsdk:"environment_id"`
	ApplicationStatus    types.String `tfsdk:"application_status"`
	Replicas             types.Int64  `tfsdk:"replicas"`
	ServerID             types.String `tfsdk:"server_id"`
	InternalConnection   types.String `tfsdk:"internal_connection"`
	ExternalConnection   types.String `tfsdk:"external_connection"`
}

func (r *MySQLResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql"
}

func (r *MySQLResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MySQL database instance in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the MySQL instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MySQL instance.",
			},
			"app_name_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix for the MySQL instance. Dokploy will append a random suffix to create the final app_name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_name": schema.StringAttribute{
				Computed:    true,
				Description: "The actual application name used by Dokploy (includes server-generated suffix).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the MySQL instance.",
			},
			"database_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the database to create.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_user": schema.StringAttribute{
				Required:    true,
				Description: "Database user name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for the database user.",
			},
			"database_root_password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Root password for the MySQL instance.",
			},
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use (defaults to mysql:8).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run in the container.",
			},
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables for the container.",
			},
			"memory_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "Memory reservation for the container.",
			},
			"memory_limit": schema.StringAttribute{
				Optional:    true,
				Description: "Memory limit for the container.",
			},
			"cpu_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "CPU reservation for the container.",
			},
			"cpu_limit": schema.StringAttribute{
				Optional:    true,
				Description: "CPU limit for the container.",
			},
			"external_port": schema.Int64Attribute{
				Optional:    true,
				Description: "External port to expose the MySQL instance.",
			},
			"internal_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal port used by the MySQL instance (default: 3306).",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the MySQL instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the MySQL application (idle, running, done, error).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the MySQL instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the MySQL instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_connection": schema.StringAttribute{
				Computed:    true,
				Description: "Internal connection string for the MySQL instance (format: mysql://user:password@app_name:internal_port/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_connection": schema.StringAttribute{
				Computed:    true,
				Description: "External connection string for the MySQL instance (format: mysql://user:password@server_ip:external_port/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MySQLResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *MySQLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MySQLResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mysql := client.MySQL{
		Name:                 plan.Name.ValueString(),
		AppName:              plan.AppNamePrefix.ValueString(),
		Description:          plan.Description.ValueString(),
		DatabaseName:         plan.DatabaseName.ValueString(),
		DatabaseUser:         plan.DatabaseUser.ValueString(),
		DatabasePassword:     plan.DatabasePassword.ValueString(),
		DatabaseRootPassword: plan.DatabaseRootPassword.ValueString(),
		DockerImage:          plan.DockerImage.ValueString(),
		EnvironmentID:        plan.EnvironmentID.ValueString(),
		ServerID:             plan.ServerID.ValueString(),
	}

	createdMySQL, err := r.client.CreateMySQL(mysql)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MySQL instance", err.Error())
		return
	}

	// Check if we need to update with additional fields not supported by create API
	needsUpdate := (!plan.Command.IsNull() && !plan.Command.IsUnknown()) ||
		(!plan.Env.IsNull() && !plan.Env.IsUnknown()) ||
		(!plan.MemoryReservation.IsNull() && !plan.MemoryReservation.IsUnknown()) ||
		(!plan.MemoryLimit.IsNull() && !plan.MemoryLimit.IsUnknown()) ||
		(!plan.CPUReservation.IsNull() && !plan.CPUReservation.IsUnknown()) ||
		(!plan.CPULimit.IsNull() && !plan.CPULimit.IsUnknown()) ||
		(!plan.ExternalPort.IsNull() && !plan.ExternalPort.IsUnknown()) ||
		(!plan.Replicas.IsNull() && !plan.Replicas.IsUnknown())

	if needsUpdate {
		updateMySQL := client.MySQL{
			MySQLID:           createdMySQL.MySQLID,
			Command:           plan.Command.ValueString(),
			Env:               plan.Env.ValueString(),
			MemoryReservation: plan.MemoryReservation.ValueString(),
			MemoryLimit:       plan.MemoryLimit.ValueString(),
			CPUReservation:    plan.CPUReservation.ValueString(),
			CPULimit:          plan.CPULimit.ValueString(),
			ExternalPort:      int(plan.ExternalPort.ValueInt64()),
			Replicas:          int(plan.Replicas.ValueInt64()),
		}

		_, err := r.client.UpdateMySQL(updateMySQL)
		if err != nil {
			resp.Diagnostics.AddError("Error updating MySQL instance after creation", err.Error())
			return
		}

		createdMySQL, err = r.client.GetMySQL(createdMySQL.MySQLID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading MySQL instance after update", err.Error())
			return
		}
	}

	// Set state from created resource
	r.mapMySQLToState(&plan, createdMySQL)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MySQLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MySQLResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mysql, err := r.client.GetMySQL(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MySQL instance", err.Error())
		return
	}

	r.mapMySQLToState(&state, mysql)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MySQLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MySQLResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mysql := client.MySQL{
		MySQLID:              plan.ID.ValueString(),
		Name:                 plan.Name.ValueString(),
		Description:          plan.Description.ValueString(),
		DatabasePassword:     plan.DatabasePassword.ValueString(),
		DatabaseRootPassword: plan.DatabaseRootPassword.ValueString(),
		DockerImage:          plan.DockerImage.ValueString(),
		Command:              plan.Command.ValueString(),
		Env:                  plan.Env.ValueString(),
		MemoryReservation:    plan.MemoryReservation.ValueString(),
		MemoryLimit:          plan.MemoryLimit.ValueString(),
		CPUReservation:       plan.CPUReservation.ValueString(),
		CPULimit:             plan.CPULimit.ValueString(),
		ExternalPort:         int(plan.ExternalPort.ValueInt64()),
		Replicas:             int(plan.Replicas.ValueInt64()),
	}

	_, err := r.client.UpdateMySQL(mysql)
	if err != nil {
		resp.Diagnostics.AddError("Error updating MySQL instance", err.Error())
		return
	}

	// Fetch updated state
	updatedMySQL, err := r.client.GetMySQL(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MySQL instance after update", err.Error())
		return
	}

	r.mapMySQLToState(&plan, updatedMySQL)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MySQLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MySQLResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMySQL(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting MySQL instance", err.Error())
		return
	}
}

func (r *MySQLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MySQLResource) mapMySQLToState(state *MySQLResourceModel, mysql *client.MySQL) {
	state.ID = types.StringValue(mysql.MySQLID)
	state.Name = types.StringValue(mysql.Name)
	state.AppName = types.StringValue(mysql.AppName)
	state.EnvironmentID = types.StringValue(mysql.EnvironmentID)
	state.ApplicationStatus = types.StringValue(mysql.ApplicationStatus)
	state.DatabaseName = types.StringValue(mysql.DatabaseName)
	state.DatabaseUser = types.StringValue(mysql.DatabaseUser)

	if mysql.DockerImage != "" {
		state.DockerImage = types.StringValue(mysql.DockerImage)
	}
	if mysql.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(mysql.Replicas))
	} else {
		state.Replicas = types.Int64Value(1)
	}

	// Optional fields
	if !state.Description.IsNull() || mysql.Description != "" {
		state.Description = types.StringValue(mysql.Description)
	}
	if !state.Command.IsNull() || mysql.Command != "" {
		state.Command = types.StringValue(mysql.Command)
	}
	if !state.Env.IsNull() || mysql.Env != "" {
		state.Env = types.StringValue(mysql.Env)
	}
	if !state.MemoryReservation.IsNull() || mysql.MemoryReservation != "" {
		state.MemoryReservation = types.StringValue(mysql.MemoryReservation)
	}
	if !state.MemoryLimit.IsNull() || mysql.MemoryLimit != "" {
		state.MemoryLimit = types.StringValue(mysql.MemoryLimit)
	}
	if !state.CPUReservation.IsNull() || mysql.CPUReservation != "" {
		state.CPUReservation = types.StringValue(mysql.CPUReservation)
	}
	if !state.CPULimit.IsNull() || mysql.CPULimit != "" {
		state.CPULimit = types.StringValue(mysql.CPULimit)
	}
	if !state.ExternalPort.IsNull() || mysql.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(mysql.ExternalPort))
	}
	if !state.ServerID.IsNull() || mysql.ServerID != "" {
		state.ServerID = types.StringValue(mysql.ServerID)
	}

	state.DatabasePassword = types.StringValue(mysql.DatabasePassword)

	internalPort := mysql.ExternalPort
	if internalPort == 0 {
		internalPort = 3306
	}
	state.InternalPort = types.Int64Value(int64(internalPort))
	state.InternalConnection = types.StringValue(fmt.Sprintf("mysql://%s:%s@%s:%d/%s", mysql.DatabaseUser, mysql.DatabasePassword, mysql.AppName, int64(internalPort), mysql.DatabaseName))

	if mysql.ServerIP != "" && mysql.ExternalPort > 0 {
		state.ExternalConnection = types.StringValue(fmt.Sprintf("mysql://%s:%s@%s:%d/%s", mysql.DatabaseUser, mysql.DatabasePassword, mysql.ServerIP, mysql.ExternalPort, mysql.DatabaseName))
	} else {
		state.ExternalConnection = types.StringValue("")
	}
}
