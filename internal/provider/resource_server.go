package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

type ServerResource struct {
	client *client.DokployClient
}

type ServerResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	IPAddress    types.String `tfsdk:"ip_address"`
	Port         types.Int64  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	SSHKeyID     types.String `tfsdk:"ssh_key_id"`
	ServerType   types.String `tfsdk:"server_type"`
	ServerStatus types.String `tfsdk:"server_status"`
	Command      types.String `tfsdk:"command"`
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a remote server for Dokploy deployments or builds.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the server.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the server.",
			},
			"ip_address": schema.StringAttribute{
				Required:    true,
				Description: "IP address of the server.",
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "SSH port of the server.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "SSH username for connecting to the server.",
			},
			"ssh_key_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the SSH key to use for authentication.",
			},
			"server_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of server: 'deploy' or 'build'.",
				Validators: []validator.String{
					stringvalidator.OneOf("deploy", "build"),
				},
			},
			"server_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the server.",
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom command to run on the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create with only the fields supported by the create API.
	// Note: command is NOT accepted by server.create, only by server.update.
	server := client.Server{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IPAddress:   plan.IPAddress.ValueString(),
		Port:        int(plan.Port.ValueInt64()),
		Username:    plan.Username.ValueString(),
		SSHKeyID:    plan.SSHKeyID.ValueString(),
		ServerType:  plan.ServerType.ValueString(),
	}

	createdServer, err := r.client.CreateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	// Check if we need to update with command field (not supported by create API).
	if !plan.Command.IsNull() && !plan.Command.IsUnknown() && plan.Command.ValueString() != "" {
		updateServer := client.Server{
			ID:          createdServer.ID,
			Name:        createdServer.Name,
			Description: createdServer.Description,
			IPAddress:   createdServer.IPAddress,
			Port:        createdServer.Port,
			Username:    createdServer.Username,
			SSHKeyID:    createdServer.SSHKeyID,
			ServerType:  createdServer.ServerType,
			Command:     plan.Command.ValueString(),
		}

		updatedServer, err := r.client.UpdateServer(updateServer)
		if err != nil {
			resp.Diagnostics.AddError("Error updating server command after creation", err.Error())
			return
		}
		createdServer = updatedServer
	}

	plan.ID = types.StringValue(createdServer.ID)
	plan.Name = types.StringValue(createdServer.Name)
	plan.Description = types.StringValue(createdServer.Description)
	plan.IPAddress = types.StringValue(createdServer.IPAddress)
	plan.Port = types.Int64Value(int64(createdServer.Port))
	plan.Username = types.StringValue(createdServer.Username)
	plan.SSHKeyID = types.StringValue(createdServer.SSHKeyID)
	plan.ServerType = types.StringValue(createdServer.ServerType)
	plan.ServerStatus = types.StringValue(createdServer.ServerStatus)
	plan.Command = types.StringValue(createdServer.Command)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	state.Name = types.StringValue(server.Name)
	state.Description = types.StringValue(server.Description)
	state.IPAddress = types.StringValue(server.IPAddress)
	state.Port = types.Int64Value(int64(server.Port))
	state.Username = types.StringValue(server.Username)
	state.SSHKeyID = types.StringValue(server.SSHKeyID)
	state.ServerType = types.StringValue(server.ServerType)
	state.ServerStatus = types.StringValue(server.ServerStatus)
	state.Command = types.StringValue(server.Command)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := client.Server{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IPAddress:   plan.IPAddress.ValueString(),
		Port:        int(plan.Port.ValueInt64()),
		Username:    plan.Username.ValueString(),
		SSHKeyID:    plan.SSHKeyID.ValueString(),
		ServerType:  plan.ServerType.ValueString(),
		Command:     plan.Command.ValueString(),
	}

	updatedServer, err := r.client.UpdateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server", err.Error())
		return
	}

	plan.Name = types.StringValue(updatedServer.Name)
	plan.Description = types.StringValue(updatedServer.Description)
	plan.IPAddress = types.StringValue(updatedServer.IPAddress)
	plan.Port = types.Int64Value(int64(updatedServer.Port))
	plan.Username = types.StringValue(updatedServer.Username)
	plan.SSHKeyID = types.StringValue(updatedServer.SSHKeyID)
	plan.ServerType = types.StringValue(updatedServer.ServerType)
	plan.ServerStatus = types.StringValue(updatedServer.ServerStatus)
	plan.Command = types.StringValue(updatedServer.Command)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServer(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting server", err.Error())
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
