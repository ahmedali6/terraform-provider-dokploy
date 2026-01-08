package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &PortResource{}
var _ resource.ResourceWithImportState = &PortResource{}

func NewPortResource() resource.Resource {
	return &PortResource{}
}

type PortResource struct {
	client *client.DokployClient
}

type PortResourceModel struct {
	ID            types.String `tfsdk:"id"`
	PublishedPort types.Int64  `tfsdk:"published_port"`
	TargetPort    types.Int64  `tfsdk:"target_port"`
	Protocol      types.String `tfsdk:"protocol"`
	PublishMode   types.String `tfsdk:"publish_mode"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *PortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port"
}

func (r *PortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a port mapping for a Dokploy application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the port mapping.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_port": schema.Int64Attribute{
				Required:    true,
				Description: "The port exposed on the host.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"target_port": schema.Int64Attribute{
				Required:    true,
				Description: "The port inside the container.",
			},
			"protocol": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Protocol: tcp or udp.",
				Default:     stringdefault.StaticString("tcp"),
			},
			"publish_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Publish mode: ingress or host.",
				Default:     stringdefault.StaticString("ingress"),
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *PortResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DokployClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *PortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PortResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	port := client.Port{
		PublishedPort: plan.PublishedPort.ValueInt64(),
		TargetPort:    plan.TargetPort.ValueInt64(),
		Protocol:      plan.Protocol.ValueString(),
		PublishMode:   plan.PublishMode.ValueString(),
		ApplicationID: plan.ApplicationID.ValueString(),
	}

	createdPort, err := r.client.CreatePort(port)
	if err != nil {
		resp.Diagnostics.AddError("Error creating port", err.Error())
		return
	}

	plan.ID = types.StringValue(createdPort.ID)
	plan.Protocol = types.StringValue(createdPort.Protocol)
	plan.PublishMode = types.StringValue(createdPort.PublishMode)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	port, err := r.client.GetPort(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading port", err.Error())
		return
	}

	state.PublishedPort = types.Int64Value(port.PublishedPort)
	state.TargetPort = types.Int64Value(port.TargetPort)
	state.Protocol = types.StringValue(port.Protocol)
	state.PublishMode = types.StringValue(port.PublishMode)
	state.ApplicationID = types.StringValue(port.ApplicationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PortResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	port := client.Port{
		ID:            plan.ID.ValueString(),
		PublishedPort: plan.PublishedPort.ValueInt64(),
		TargetPort:    plan.TargetPort.ValueInt64(),
		Protocol:      plan.Protocol.ValueString(),
		PublishMode:   plan.PublishMode.ValueString(),
	}

	updatedPort, err := r.client.UpdatePort(port)
	if err != nil {
		resp.Diagnostics.AddError("Error updating port", err.Error())
		return
	}

	plan.PublishedPort = types.Int64Value(updatedPort.PublishedPort)
	plan.TargetPort = types.Int64Value(updatedPort.TargetPort)
	plan.Protocol = types.StringValue(updatedPort.Protocol)
	plan.PublishMode = types.StringValue(updatedPort.PublishMode)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePort(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting port", err.Error())
		return
	}
}

func (r *PortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
