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

var _ resource.Resource = &PostgresResource{}
var _ resource.ResourceWithImportState = &PostgresResource{}

func NewPostgresResource() resource.Resource {
	return &PostgresResource{}
}

type PostgresResource struct {
	client *client.DokployClient
}

type PostgresResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AppNamePrefix      types.String `tfsdk:"app_name_prefix"`
	AppName            types.String `tfsdk:"app_name"`
	Description        types.String `tfsdk:"description"`
	DatabaseName       types.String `tfsdk:"database_name"`
	DatabaseUser       types.String `tfsdk:"database_user"`
	DatabasePassword   types.String `tfsdk:"database_password"`
	DockerImage        types.String `tfsdk:"docker_image"`
	Command            types.String `tfsdk:"command"`
	Env                types.String `tfsdk:"env"`
	MemoryReservation  types.String `tfsdk:"memory_reservation"`
	MemoryLimit        types.String `tfsdk:"memory_limit"`
	CPUReservation     types.String `tfsdk:"cpu_reservation"`
	CPULimit           types.String `tfsdk:"cpu_limit"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	ApplicationStatus  types.String `tfsdk:"application_status"`
	Replicas           types.Int64  `tfsdk:"replicas"`
	ServerID           types.String `tfsdk:"server_id"`
	InternalConnection types.String `tfsdk:"internal_connection"`
	ExternalConnection types.String `tfsdk:"external_connection"`
}

func (r *PostgresResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

func (r *PostgresResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PostgreSQL database instance in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the PostgreSQL instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the PostgreSQL instance.",
			},
			"app_name_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix for the PostgreSQL instance. Dokploy will append a random suffix to create the final app_name.",
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
				Description: "Description of the PostgreSQL instance.",
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
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use (defaults to postgres:15).",
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
				Description: "External port to expose the PostgreSQL instance.",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the PostgreSQL instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the PostgreSQL application (idle, running, done, error).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the PostgreSQL instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the PostgreSQL instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_connection": schema.StringAttribute{
				Computed:    true,
				Description: "Internal connection string for the PostgreSQL instance (format: postgres://user:password@app_name/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_connection": schema.StringAttribute{
				Computed:    true,
				Description: "External connection string for the PostgreSQL instance (format: postgres://user:password@server_ip:port/database_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PostgresResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PostgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgres := client.Postgres{
		Name:             plan.Name.ValueString(),
		AppName:          plan.AppNamePrefix.ValueString(),
		Description:      plan.Description.ValueString(),
		DatabaseName:     plan.DatabaseName.ValueString(),
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
		DockerImage:      plan.DockerImage.ValueString(),
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		ServerID:         plan.ServerID.ValueString(),
	}

	createdPostgres, err := r.client.CreatePostgres(postgres)
	if err != nil {
		resp.Diagnostics.AddError("Error creating PostgreSQL instance", err.Error())
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
		updatePostgres := client.Postgres{
			PostgresID:        createdPostgres.PostgresID,
			Command:           plan.Command.ValueString(),
			Env:               plan.Env.ValueString(),
			MemoryReservation: plan.MemoryReservation.ValueString(),
			MemoryLimit:       plan.MemoryLimit.ValueString(),
			CPUReservation:    plan.CPUReservation.ValueString(),
			CPULimit:          plan.CPULimit.ValueString(),
			ExternalPort:      int(plan.ExternalPort.ValueInt64()),
			Replicas:          int(plan.Replicas.ValueInt64()),
		}

		_, err := r.client.UpdatePostgres(updatePostgres)
		if err != nil {
			resp.Diagnostics.AddError("Error updating PostgreSQL instance after creation", err.Error())
			return
		}

		createdPostgres, err = r.client.GetPostgres(createdPostgres.PostgresID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading PostgreSQL instance after update", err.Error())
			return
		}
	}

	// Set state from created resource
	r.mapPostgresToState(&plan, createdPostgres)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgres, err := r.client.GetPostgres(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading PostgreSQL instance", err.Error())
		return
	}

	r.mapPostgresToState(&state, postgres)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PostgresResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgres := client.Postgres{
		PostgresID:        plan.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		DatabasePassword:  plan.DatabasePassword.ValueString(),
		DockerImage:       plan.DockerImage.ValueString(),
		Command:           plan.Command.ValueString(),
		Env:               plan.Env.ValueString(),
		MemoryReservation: plan.MemoryReservation.ValueString(),
		MemoryLimit:       plan.MemoryLimit.ValueString(),
		CPUReservation:    plan.CPUReservation.ValueString(),
		CPULimit:          plan.CPULimit.ValueString(),
		ExternalPort:      int(plan.ExternalPort.ValueInt64()),
		Replicas:          int(plan.Replicas.ValueInt64()),
	}

	_, err := r.client.UpdatePostgres(postgres)
	if err != nil {
		resp.Diagnostics.AddError("Error updating PostgreSQL instance", err.Error())
		return
	}

	// Fetch updated state
	updatedPostgres, err := r.client.GetPostgres(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading PostgreSQL instance after update", err.Error())
		return
	}

	r.mapPostgresToState(&plan, updatedPostgres)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePostgres(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting PostgreSQL instance", err.Error())
		return
	}
}

func (r *PostgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PostgresResource) mapPostgresToState(state *PostgresResourceModel, postgres *client.Postgres) {
	state.ID = types.StringValue(postgres.PostgresID)
	state.Name = types.StringValue(postgres.Name)
	state.AppName = types.StringValue(postgres.AppName)
	state.EnvironmentID = types.StringValue(postgres.EnvironmentID)
	state.ApplicationStatus = types.StringValue(postgres.ApplicationStatus)
	state.DatabaseName = types.StringValue(postgres.DatabaseName)
	state.DatabaseUser = types.StringValue(postgres.DatabaseUser)

	if postgres.DockerImage != "" {
		state.DockerImage = types.StringValue(postgres.DockerImage)
	}
	if postgres.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(postgres.Replicas))
	} else {
		state.Replicas = types.Int64Value(1)
	}

	// Optional fields
	if !state.Description.IsNull() || postgres.Description != "" {
		state.Description = types.StringValue(postgres.Description)
	}
	if !state.Command.IsNull() || postgres.Command != "" {
		state.Command = types.StringValue(postgres.Command)
	}
	if !state.Env.IsNull() || postgres.Env != "" {
		state.Env = types.StringValue(postgres.Env)
	}
	if !state.MemoryReservation.IsNull() || postgres.MemoryReservation != "" {
		state.MemoryReservation = types.StringValue(postgres.MemoryReservation)
	}
	if !state.MemoryLimit.IsNull() || postgres.MemoryLimit != "" {
		state.MemoryLimit = types.StringValue(postgres.MemoryLimit)
	}
	if !state.CPUReservation.IsNull() || postgres.CPUReservation != "" {
		state.CPUReservation = types.StringValue(postgres.CPUReservation)
	}
	if !state.CPULimit.IsNull() || postgres.CPULimit != "" {
		state.CPULimit = types.StringValue(postgres.CPULimit)
	}
	if !state.ExternalPort.IsNull() || postgres.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(postgres.ExternalPort))
	}
	if !state.ServerID.IsNull() || postgres.ServerID != "" {
		state.ServerID = types.StringValue(postgres.ServerID)
	}

	state.DatabasePassword = types.StringValue(postgres.DatabasePassword)
	state.InternalConnection = types.StringValue(fmt.Sprintf("postgres://%s:%s@%s/%s", postgres.DatabaseUser, postgres.DatabasePassword, postgres.AppName, postgres.DatabaseName))

	if postgres.ServerIP != "" && postgres.ExternalPort > 0 {
		state.ExternalConnection = types.StringValue(fmt.Sprintf("postgres://%s:%s@%s:%d/%s", postgres.DatabaseUser, postgres.DatabasePassword, postgres.ServerIP, postgres.ExternalPort, postgres.DatabaseName))
	}
}
