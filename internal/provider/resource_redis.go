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

var _ resource.Resource = &RedisResource{}
var _ resource.ResourceWithImportState = &RedisResource{}

func NewRedisResource() resource.Resource {
	return &RedisResource{}
}

type RedisResource struct {
	client *client.DokployClient
}

type RedisResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AppNamePrefix      types.String `tfsdk:"app_name_prefix"`
	AppName            types.String `tfsdk:"app_name"`
	Description        types.String `tfsdk:"description"`
	DatabaseUser       types.String `tfsdk:"database_user"`
	DatabasePassword   types.String `tfsdk:"database_password"`
	DockerImage        types.String `tfsdk:"docker_image"`
	Command            types.String `tfsdk:"command"`
	Env                types.String `tfsdk:"env"`
	MemoryReservation  types.String `tfsdk:"memory_reservation"`
	MemoryLimit        types.String `tfsdk:"memory_limit"`
	CPUReservation     types.String `tfsdk:"cpu_reservation"`
	CPULimit           types.String `tfsdk:"cpu_limit"`
	InternalPort       types.Int64  `tfsdk:"internal_port"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	ApplicationStatus  types.String `tfsdk:"application_status"`
	Replicas           types.Int64  `tfsdk:"replicas"`
	ServerID           types.String `tfsdk:"server_id"`
	InternalConnection types.String `tfsdk:"internal_connection"`
	ExternalConnection types.String `tfsdk:"external_connection"`
}

func (r *RedisResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

func (r *RedisResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Redis database instance in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the Redis instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Redis instance.",
			},
			"app_name_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix for the Redis instance. Dokploy will append a random suffix to create the final app_name.",
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
				Description: "Description of the Redis instance.",
			},
			"database_user": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Database user name (defaults to 'default' for Redis).",
			},
			"database_password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for the Redis database.",
			},
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use for Redis (defaults to official Redis image).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run in the Redis container.",
			},
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables for the Redis container.",
			},
			"memory_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "Memory reservation for the Redis container.",
			},
			"memory_limit": schema.StringAttribute{
				Optional:    true,
				Description: "Memory limit for the Redis container.",
			},
			"cpu_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "CPU reservation for the Redis container.",
			},
			"cpu_limit": schema.StringAttribute{
				Optional:    true,
				Description: "CPU limit for the Redis container.",
			},
			"external_port": schema.Int64Attribute{
				Optional:    true,
				Description: "External port to expose the Redis instance.",
			},
			"internal_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal port used by the Redis instance (default: 6379).",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the Redis instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the Redis application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the Redis instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the Redis instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_connection": schema.StringAttribute{
				Computed:    true,
				Description: "Internal connection string for Redis instance (format: redis://user:password@app_name:internal_port).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_connection": schema.StringAttribute{
				Computed:    true,
				Description: "External connection string for Redis instance (format: redis://user:password@server_ip:external_port).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RedisResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RedisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedisResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create with only the fields supported by the create API.
	redis := client.Redis{
		Name:             plan.Name.ValueString(),
		AppName:          plan.AppNamePrefix.ValueString(),
		Description:      plan.Description.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
		DockerImage:      plan.DockerImage.ValueString(),
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		ServerID:         plan.ServerID.ValueString(),
	}

	// Set default user for Redis (Redis doesn't support user management)
	if plan.DatabaseUser.IsNull() || plan.DatabaseUser.IsUnknown() {
		plan.DatabaseUser = types.StringValue("default")
	}

	createdRedis, err := r.client.CreateRedis(redis)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Redis instance", err.Error())
		return
	}

	// Check if we need to update with additional fields not supported by create API.
	// Only trigger update if a field is explicitly set (not null AND not unknown).
	needsUpdate := (!plan.Command.IsNull() && !plan.Command.IsUnknown()) ||
		(!plan.Env.IsNull() && !plan.Env.IsUnknown()) ||
		(!plan.MemoryReservation.IsNull() && !plan.MemoryReservation.IsUnknown()) ||
		(!plan.MemoryLimit.IsNull() && !plan.MemoryLimit.IsUnknown()) ||
		(!plan.CPUReservation.IsNull() && !plan.CPUReservation.IsUnknown()) ||
		(!plan.CPULimit.IsNull() && !plan.CPULimit.IsUnknown()) ||
		(!plan.ExternalPort.IsNull() && !plan.ExternalPort.IsUnknown()) ||
		(!plan.Replicas.IsNull() && !plan.Replicas.IsUnknown())

	if needsUpdate {
		updateRedis := client.Redis{
			RedisID:           createdRedis.RedisID,
			Command:           plan.Command.ValueString(),
			Env:               plan.Env.ValueString(),
			MemoryReservation: plan.MemoryReservation.ValueString(),
			MemoryLimit:       plan.MemoryLimit.ValueString(),
			CPUReservation:    plan.CPUReservation.ValueString(),
			CPULimit:          plan.CPULimit.ValueString(),
			ExternalPort:      int(plan.ExternalPort.ValueInt64()),
			Replicas:          int(plan.Replicas.ValueInt64()),
		}

		_, err := r.client.UpdateRedis(updateRedis)
		if err != nil {
			resp.Diagnostics.AddError("Error updating Redis instance after creation", err.Error())
			return
		}

		// Fetch the updated resource to get the final state.
		createdRedis, err = r.client.GetRedis(createdRedis.RedisID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading Redis instance after update", err.Error())
			return
		}
	}

	// Set required and computed fields.
	plan.ID = types.StringValue(createdRedis.RedisID)
	plan.Name = types.StringValue(createdRedis.Name)
	// Store the server-modified app_name so state matches the remote resource.
	plan.AppName = types.StringValue(createdRedis.AppName)
	plan.EnvironmentID = types.StringValue(createdRedis.EnvironmentID)
	plan.ApplicationStatus = types.StringValue(createdRedis.ApplicationStatus)
	// Redis doesn't support user management, always use "default"
	plan.DatabaseUser = types.StringValue("default")

	// Set computed fields that have defaults.
	if createdRedis.DockerImage != "" {
		plan.DockerImage = types.StringValue(createdRedis.DockerImage)
	}
	if createdRedis.Replicas > 0 {
		plan.Replicas = types.Int64Value(int64(createdRedis.Replicas))
	} else {
		plan.Replicas = types.Int64Value(1) // Default to 1.
	}

	// Only update optional fields if they were set in config or have non-empty values.
	if !plan.Description.IsNull() || createdRedis.Description != "" {
		plan.Description = types.StringValue(createdRedis.Description)
	}
	if !plan.Command.IsNull() || createdRedis.Command != "" {
		plan.Command = types.StringValue(createdRedis.Command)
	}
	if !plan.Env.IsNull() || createdRedis.Env != "" {
		plan.Env = types.StringValue(createdRedis.Env)
	}
	if !plan.MemoryReservation.IsNull() || createdRedis.MemoryReservation != "" {
		plan.MemoryReservation = types.StringValue(createdRedis.MemoryReservation)
	}
	if !plan.MemoryLimit.IsNull() || createdRedis.MemoryLimit != "" {
		plan.MemoryLimit = types.StringValue(createdRedis.MemoryLimit)
	}
	if !plan.CPUReservation.IsNull() || createdRedis.CPUReservation != "" {
		plan.CPUReservation = types.StringValue(createdRedis.CPUReservation)
	}
	if !plan.CPULimit.IsNull() || createdRedis.CPULimit != "" {
		plan.CPULimit = types.StringValue(createdRedis.CPULimit)
	}
	if !plan.ExternalPort.IsNull() || createdRedis.ExternalPort > 0 {
		plan.ExternalPort = types.Int64Value(int64(createdRedis.ExternalPort))
	}
	if !plan.ServerID.IsNull() || createdRedis.ServerID != "" {
		plan.ServerID = types.StringValue(createdRedis.ServerID)
	}

	plan.DatabasePassword = types.StringValue(createdRedis.DatabasePassword)

	internalPort := createdRedis.ExternalPort
	if internalPort == 0 {
		internalPort = 6379
	}
	plan.InternalPort = types.Int64Value(int64(internalPort))
	plan.InternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", createdRedis.DatabaseUser, createdRedis.DatabasePassword, createdRedis.AppName, int64(internalPort)))

	if createdRedis.ServerIP != "" && createdRedis.ExternalPort > 0 {
		plan.ExternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", createdRedis.DatabaseUser, createdRedis.DatabasePassword, createdRedis.ServerIP, createdRedis.ExternalPort))
	} else {
		plan.ExternalConnection = types.StringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedisResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redis, err := r.client.GetRedis(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Redis instance", err.Error())
		return
	}

	// Update required and computed fields.
	// Note: AppNamePrefix is not updated from server - it's user-provided config.
	state.Name = types.StringValue(redis.Name)
	state.AppName = types.StringValue(redis.AppName)
	state.EnvironmentID = types.StringValue(redis.EnvironmentID)
	state.ApplicationStatus = types.StringValue(redis.ApplicationStatus)
	// Redis doesn't support user management, always use "default"
	state.DatabaseUser = types.StringValue("default")

	// Update computed fields.
	if redis.DockerImage != "" {
		state.DockerImage = types.StringValue(redis.DockerImage)
	}
	if redis.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(redis.Replicas))
	}

	// Update optional fields only if they have values in state or from API.
	if !state.Description.IsNull() || redis.Description != "" {
		state.Description = types.StringValue(redis.Description)
	}
	if !state.Command.IsNull() || redis.Command != "" {
		state.Command = types.StringValue(redis.Command)
	}
	if !state.Env.IsNull() || redis.Env != "" {
		state.Env = types.StringValue(redis.Env)
	}
	if !state.MemoryReservation.IsNull() || redis.MemoryReservation != "" {
		state.MemoryReservation = types.StringValue(redis.MemoryReservation)
	}
	if !state.MemoryLimit.IsNull() || redis.MemoryLimit != "" {
		state.MemoryLimit = types.StringValue(redis.MemoryLimit)
	}
	if !state.CPUReservation.IsNull() || redis.CPUReservation != "" {
		state.CPUReservation = types.StringValue(redis.CPUReservation)
	}
	if !state.CPULimit.IsNull() || redis.CPULimit != "" {
		state.CPULimit = types.StringValue(redis.CPULimit)
	}
	if !state.ExternalPort.IsNull() || redis.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(redis.ExternalPort))
	}
	if !state.ServerID.IsNull() || redis.ServerID != "" {
		state.ServerID = types.StringValue(redis.ServerID)
	}

	state.DatabasePassword = types.StringValue(redis.DatabasePassword)

	internalPort := redis.ExternalPort
	if internalPort == 0 {
		internalPort = 6379
	}
	state.InternalPort = types.Int64Value(int64(internalPort))
	state.InternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", redis.DatabaseUser, redis.DatabasePassword, redis.AppName, int64(internalPort)))

	if redis.ServerIP != "" && redis.ExternalPort > 0 {
		state.ExternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", redis.DatabaseUser, redis.DatabasePassword, redis.ServerIP, redis.ExternalPort))
	} else {
		state.ExternalConnection = types.StringValue("")
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedisResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redis := client.Redis{
		RedisID:           plan.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		AppName:           plan.AppName.ValueString(),
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

	updatedRedis, err := r.client.UpdateRedis(redis)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Redis instance", err.Error())
		return
	}

	// Update required and computed fields.
	// Note: AppNamePrefix is not updated - it's user-provided config that triggers replace.
	plan.Name = types.StringValue(updatedRedis.Name)
	// Update the computed app_name from server.
	plan.AppName = types.StringValue(updatedRedis.AppName)
	plan.ApplicationStatus = types.StringValue(updatedRedis.ApplicationStatus)
	plan.DatabaseUser = types.StringValue(updatedRedis.DatabaseUser)

	// Update computed fields.
	if updatedRedis.DockerImage != "" {
		plan.DockerImage = types.StringValue(updatedRedis.DockerImage)
	}
	if updatedRedis.Replicas > 0 {
		plan.Replicas = types.Int64Value(int64(updatedRedis.Replicas))
	}

	// Update optional fields only if they have values in plan or from API.
	if !plan.Description.IsNull() || updatedRedis.Description != "" {
		plan.Description = types.StringValue(updatedRedis.Description)
	}
	if !plan.Command.IsNull() || updatedRedis.Command != "" {
		plan.Command = types.StringValue(updatedRedis.Command)
	}
	if !plan.Env.IsNull() || updatedRedis.Env != "" {
		plan.Env = types.StringValue(updatedRedis.Env)
	}
	if !plan.MemoryReservation.IsNull() || updatedRedis.MemoryReservation != "" {
		plan.MemoryReservation = types.StringValue(updatedRedis.MemoryReservation)
	}
	if !plan.MemoryLimit.IsNull() || updatedRedis.MemoryLimit != "" {
		plan.MemoryLimit = types.StringValue(updatedRedis.MemoryLimit)
	}
	if !plan.CPUReservation.IsNull() || updatedRedis.CPUReservation != "" {
		plan.CPUReservation = types.StringValue(updatedRedis.CPUReservation)
	}
	if !plan.CPULimit.IsNull() || updatedRedis.CPULimit != "" {
		plan.CPULimit = types.StringValue(updatedRedis.CPULimit)
	}
	if !plan.ExternalPort.IsNull() || updatedRedis.ExternalPort > 0 {
		plan.ExternalPort = types.Int64Value(int64(updatedRedis.ExternalPort))
	}

	plan.DatabasePassword = types.StringValue(updatedRedis.DatabasePassword)

	internalPort := updatedRedis.ExternalPort
	if internalPort == 0 {
		internalPort = 6379
	}
	plan.InternalPort = types.Int64Value(int64(internalPort))
	plan.InternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", updatedRedis.DatabaseUser, updatedRedis.DatabasePassword, updatedRedis.AppName, int64(internalPort)))

	if updatedRedis.ServerIP != "" && updatedRedis.ExternalPort > 0 {
		plan.ExternalConnection = types.StringValue(fmt.Sprintf("redis://%s:%s@%s:%d", updatedRedis.DatabaseUser, updatedRedis.DatabasePassword, updatedRedis.ServerIP, updatedRedis.ExternalPort))
	} else {
		plan.ExternalConnection = types.StringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedisResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRedis(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Redis instance", err.Error())
		return
	}
}

func (r *RedisResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
