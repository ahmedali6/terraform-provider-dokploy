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

var _ resource.Resource = &MariaDBResource{}
var _ resource.ResourceWithImportState = &MariaDBResource{}

func NewMariaDBResource() resource.Resource {
	return &MariaDBResource{}
}

type MariaDBResource struct {
	client *client.DokployClient
}

type MariaDBResourceModel struct {
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

func (r *MariaDBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mariadb"
}

func (r *MariaDBResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MariaDB database instance in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the MariaDB instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MariaDB instance.",
			},
			"app_name_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix for the MariaDB instance. Dokploy will append a random suffix to create the final app_name.",
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
				Description: "Description of the MariaDB instance.",
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
				Description: "Root password for the MariaDB instance.",
			},
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use (defaults to mariadb:11).",
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
				Description: "External port to expose the MariaDB instance.",
			},
			"internal_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal port used by the MariaDB instance (default: 3306).",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the MariaDB instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the MariaDB application (idle, running, done, error).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the MariaDB instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the MariaDB instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_connection": schema.StringAttribute{
				Computed:    true,
				Description: "Internal connection string for the MariaDB instance (format: mariadb://user:password@app_name:internal_port/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_connection": schema.StringAttribute{
				Computed:    true,
				Description: "External connection string for the MariaDB instance (format: mariadb://user:password@server_ip:external_port/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MariaDBResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MariaDBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MariaDBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mariadb := client.MariaDB{
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

	createdMariaDB, err := r.client.CreateMariaDB(mariadb)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MariaDB instance", err.Error())
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
		updateMariaDB := client.MariaDB{
			MariaDBID:         createdMariaDB.MariaDBID,
			Command:           plan.Command.ValueString(),
			Env:               plan.Env.ValueString(),
			MemoryReservation: plan.MemoryReservation.ValueString(),
			MemoryLimit:       plan.MemoryLimit.ValueString(),
			CPUReservation:    plan.CPUReservation.ValueString(),
			CPULimit:          plan.CPULimit.ValueString(),
			ExternalPort:      int(plan.ExternalPort.ValueInt64()),
			Replicas:          int(plan.Replicas.ValueInt64()),
		}

		_, err := r.client.UpdateMariaDB(updateMariaDB)
		if err != nil {
			resp.Diagnostics.AddError("Error updating MariaDB instance after creation", err.Error())
			return
		}

		createdMariaDB, err = r.client.GetMariaDB(createdMariaDB.MariaDBID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading MariaDB instance after update", err.Error())
			return
		}
	}

	// Set state from created resource
	r.mapMariaDBToState(&plan, createdMariaDB)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MariaDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MariaDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mariadb, err := r.client.GetMariaDB(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MariaDB instance", err.Error())
		return
	}

	r.mapMariaDBToState(&state, mariadb)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MariaDBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MariaDBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mariadb := client.MariaDB{
		MariaDBID:            plan.ID.ValueString(),
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

	_, err := r.client.UpdateMariaDB(mariadb)
	if err != nil {
		resp.Diagnostics.AddError("Error updating MariaDB instance", err.Error())
		return
	}

	// Fetch updated state
	updatedMariaDB, err := r.client.GetMariaDB(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MariaDB instance after update", err.Error())
		return
	}

	r.mapMariaDBToState(&plan, updatedMariaDB)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MariaDBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MariaDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMariaDB(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting MariaDB instance", err.Error())
		return
	}
}

func (r *MariaDBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MariaDBResource) mapMariaDBToState(state *MariaDBResourceModel, mariadb *client.MariaDB) {
	state.ID = types.StringValue(mariadb.MariaDBID)
	state.Name = types.StringValue(mariadb.Name)
	state.AppName = types.StringValue(mariadb.AppName)
	state.EnvironmentID = types.StringValue(mariadb.EnvironmentID)
	state.ApplicationStatus = types.StringValue(mariadb.ApplicationStatus)
	state.DatabaseName = types.StringValue(mariadb.DatabaseName)
	state.DatabaseUser = types.StringValue(mariadb.DatabaseUser)

	if mariadb.DockerImage != "" {
		state.DockerImage = types.StringValue(mariadb.DockerImage)
	}
	if mariadb.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(mariadb.Replicas))
	} else {
		state.Replicas = types.Int64Value(1)
	}

	// Optional fields
	if !state.Description.IsNull() || mariadb.Description != "" {
		state.Description = types.StringValue(mariadb.Description)
	}
	if !state.Command.IsNull() || mariadb.Command != "" {
		state.Command = types.StringValue(mariadb.Command)
	}
	if !state.Env.IsNull() || mariadb.Env != "" {
		state.Env = types.StringValue(mariadb.Env)
	}
	if !state.MemoryReservation.IsNull() || mariadb.MemoryReservation != "" {
		state.MemoryReservation = types.StringValue(mariadb.MemoryReservation)
	}
	if !state.MemoryLimit.IsNull() || mariadb.MemoryLimit != "" {
		state.MemoryLimit = types.StringValue(mariadb.MemoryLimit)
	}
	if !state.CPUReservation.IsNull() || mariadb.CPUReservation != "" {
		state.CPUReservation = types.StringValue(mariadb.CPUReservation)
	}
	if !state.CPULimit.IsNull() || mariadb.CPULimit != "" {
		state.CPULimit = types.StringValue(mariadb.CPULimit)
	}
	if !state.ExternalPort.IsNull() || mariadb.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(mariadb.ExternalPort))
	}
	if !state.ServerID.IsNull() || mariadb.ServerID != "" {
		state.ServerID = types.StringValue(mariadb.ServerID)
	}

	state.DatabasePassword = types.StringValue(mariadb.DatabasePassword)

	internalPort := mariadb.ExternalPort
	if internalPort == 0 {
		internalPort = 3306
	}
	state.InternalPort = types.Int64Value(int64(internalPort))
	state.InternalConnection = types.StringValue(fmt.Sprintf("mariadb://%s:%s@%s:%d/%s", mariadb.DatabaseUser, mariadb.DatabasePassword, mariadb.AppName, int64(internalPort), mariadb.DatabaseName))

	if mariadb.ServerIP != "" && mariadb.ExternalPort > 0 {
		state.ExternalConnection = types.StringValue(fmt.Sprintf("mariadb://%s:%s@%s:%d/%s", mariadb.DatabaseUser, mariadb.DatabasePassword, mariadb.ServerIP, mariadb.ExternalPort, mariadb.DatabaseName))
	} else {
		state.ExternalConnection = types.StringValue("")
	}
}
