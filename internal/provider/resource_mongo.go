package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MongoDBResource{}
var _ resource.ResourceWithImportState = &MongoDBResource{}

func NewMongoDBResource() resource.Resource {
	return &MongoDBResource{}
}

type MongoDBResource struct {
	client *client.DokployClient
}

type MongoDBResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AppNamePrefix      types.String `tfsdk:"app_name_prefix"`
	AppName            types.String `tfsdk:"app_name"`
	Description        types.String `tfsdk:"description"`
	DatabaseUser       types.String `tfsdk:"database_user"`
	DatabasePassword   types.String `tfsdk:"database_password"`
	ReplicaSets        types.Bool   `tfsdk:"replica_sets"`
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

func (r *MongoDBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongo"
}

func (r *MongoDBResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MongoDB database instance in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the MongoDB instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MongoDB instance.",
			},
			"app_name_prefix": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix for the MongoDB instance. Dokploy will append a random suffix to create the final app_name.",
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
				Description: "Description of the MongoDB instance.",
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
			"replica_sets": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable replica sets for the MongoDB instance.",
			},
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use (defaults to mongo:6).",
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
				Description: "External port to expose the MongoDB instance.",
			},
			"internal_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal port used by the MongoDB instance (default: 27017).",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the MongoDB instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the MongoDB application (idle, running, done, error).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the MongoDB instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the MongoDB instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_connection": schema.StringAttribute{
				Computed:    true,
				Description: "Internal connection string for the MongoDB instance (format: mongodb://user:password@app_name:internal_port).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_connection": schema.StringAttribute{
				Computed:    true,
				Description: "External connection string for the MongoDB instance (format: mongodb://user:password@server_ip:external_port).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MongoDBResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MongoDBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MongoDBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mongo := client.MongoDB{
		Name:             plan.Name.ValueString(),
		AppName:          plan.AppNamePrefix.ValueString(),
		Description:      plan.Description.ValueString(),
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
		ReplicaSets:      plan.ReplicaSets.ValueBool(),
		DockerImage:      plan.DockerImage.ValueString(),
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		ServerID:         plan.ServerID.ValueString(),
	}

	createdMongo, err := r.client.CreateMongoDB(mongo)
	if err != nil {
		resp.Diagnostics.AddError("Error creating MongoDB instance", err.Error())
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
		updateMongo := client.MongoDB{
			MongoID:           createdMongo.MongoID,
			Command:           plan.Command.ValueString(),
			Env:               plan.Env.ValueString(),
			MemoryReservation: plan.MemoryReservation.ValueString(),
			MemoryLimit:       plan.MemoryLimit.ValueString(),
			CPUReservation:    plan.CPUReservation.ValueString(),
			CPULimit:          plan.CPULimit.ValueString(),
			ExternalPort:      int(plan.ExternalPort.ValueInt64()),
			Replicas:          int(plan.Replicas.ValueInt64()),
		}

		_, err := r.client.UpdateMongoDB(updateMongo)
		if err != nil {
			resp.Diagnostics.AddError("Error updating MongoDB instance after creation", err.Error())
			return
		}

		createdMongo, err = r.client.GetMongoDB(createdMongo.MongoID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading MongoDB instance after update", err.Error())
			return
		}
	}

	// Set state from created resource
	r.mapMongoDBToState(&plan, createdMongo)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MongoDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mongo, err := r.client.GetMongoDB(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MongoDB instance", err.Error())
		return
	}

	r.mapMongoDBToState(&state, mongo)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoDBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MongoDBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mongo := client.MongoDB{
		MongoID:           plan.ID.ValueString(),
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		DatabasePassword:  plan.DatabasePassword.ValueString(),
		ReplicaSets:       plan.ReplicaSets.ValueBool(),
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

	_, err := r.client.UpdateMongoDB(mongo)
	if err != nil {
		resp.Diagnostics.AddError("Error updating MongoDB instance", err.Error())
		return
	}

	// Fetch updated state
	updatedMongo, err := r.client.GetMongoDB(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading MongoDB instance after update", err.Error())
		return
	}

	r.mapMongoDBToState(&plan, updatedMongo)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoDBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MongoDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMongoDB(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting MongoDB instance", err.Error())
		return
	}
}

func (r *MongoDBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MongoDBResource) mapMongoDBToState(state *MongoDBResourceModel, mongo *client.MongoDB) {
	state.ID = types.StringValue(mongo.MongoID)
	state.Name = types.StringValue(mongo.Name)
	state.AppName = types.StringValue(mongo.AppName)
	state.EnvironmentID = types.StringValue(mongo.EnvironmentID)
	state.ApplicationStatus = types.StringValue(mongo.ApplicationStatus)
	state.DatabaseUser = types.StringValue(mongo.DatabaseUser)
	state.ReplicaSets = types.BoolValue(mongo.ReplicaSets)

	if mongo.DockerImage != "" {
		state.DockerImage = types.StringValue(mongo.DockerImage)
	}
	if mongo.Replicas > 0 {
		state.Replicas = types.Int64Value(int64(mongo.Replicas))
	} else {
		state.Replicas = types.Int64Value(1)
	}

	// Optional fields
	if !state.Description.IsNull() || mongo.Description != "" {
		state.Description = types.StringValue(mongo.Description)
	}
	if !state.Command.IsNull() || mongo.Command != "" {
		state.Command = types.StringValue(mongo.Command)
	}
	if !state.Env.IsNull() || mongo.Env != "" {
		state.Env = types.StringValue(mongo.Env)
	}
	if !state.MemoryReservation.IsNull() || mongo.MemoryReservation != "" {
		state.MemoryReservation = types.StringValue(mongo.MemoryReservation)
	}
	if !state.MemoryLimit.IsNull() || mongo.MemoryLimit != "" {
		state.MemoryLimit = types.StringValue(mongo.MemoryLimit)
	}
	if !state.CPUReservation.IsNull() || mongo.CPUReservation != "" {
		state.CPUReservation = types.StringValue(mongo.CPUReservation)
	}
	if !state.CPULimit.IsNull() || mongo.CPULimit != "" {
		state.CPULimit = types.StringValue(mongo.CPULimit)
	}
	if !state.ExternalPort.IsNull() || mongo.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(mongo.ExternalPort))
	}
	if !state.ServerID.IsNull() || mongo.ServerID != "" {
		state.ServerID = types.StringValue(mongo.ServerID)
	}

	state.DatabasePassword = types.StringValue(mongo.DatabasePassword)

	internalPort := mongo.ExternalPort
	if internalPort == 0 {
		internalPort = 27017
	}
	state.InternalPort = types.Int64Value(int64(internalPort))
	state.InternalConnection = types.StringValue(fmt.Sprintf("mongodb://%s:%s@%s:%d", mongo.DatabaseUser, mongo.DatabasePassword, mongo.AppName, int64(internalPort)))

	if mongo.ServerIP != "" && mongo.ExternalPort > 0 {
		state.ExternalConnection = types.StringValue(fmt.Sprintf("mongodb://%s:%s@%s:%d", mongo.DatabaseUser, mongo.DatabasePassword, mongo.ServerIP, mongo.ExternalPort))
	} else {
		state.ExternalConnection = types.StringValue("")
	}
}
