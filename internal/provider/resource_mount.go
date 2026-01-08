package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &MountResource{}
var _ resource.ResourceWithImportState = &MountResource{}

func NewMountResource() resource.Resource {
	return &MountResource{}
}

type MountResource struct {
	client *client.DokployClient
}

type MountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	HostPath    types.String `tfsdk:"host_path"`
	VolumeName  types.String `tfsdk:"volume_name"`
	Content     types.String `tfsdk:"content"`
	MountPath   types.String `tfsdk:"mount_path"`
	ServiceType types.String `tfsdk:"service_type"`
	FilePath    types.String `tfsdk:"file_path"`
	ServiceID   types.String `tfsdk:"service_id"`
}

func (r *MountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mount"
}

func (r *MountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mount (volume, bind, or file) for Dokploy services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the mount.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of mount: bind, volume, or file.",
			},
			"host_path": schema.StringAttribute{
				Optional:    true,
				Description: "Host path for bind mounts.",
			},
			"volume_name": schema.StringAttribute{
				Optional:    true,
				Description: "Volume name for volume mounts.",
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Content for file mounts.",
			},
			"mount_path": schema.StringAttribute{
				Required:    true,
				Description: "Path where the mount will be mounted inside the container.",
			},
			"service_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type of service: application, postgres, mysql, mariadb, mongo, redis, compose.",
				Default:     stringdefault.StaticString("application"),
			},
			"file_path": schema.StringAttribute{
				Optional:    true,
				Description: "File path for file mounts.",
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the service (application, database, or compose) to mount to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *MountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mount := client.Mount{
		Type:        plan.Type.ValueString(),
		HostPath:    plan.HostPath.ValueString(),
		VolumeName:  plan.VolumeName.ValueString(),
		Content:     plan.Content.ValueString(),
		MountPath:   plan.MountPath.ValueString(),
		ServiceType: plan.ServiceType.ValueString(),
		FilePath:    plan.FilePath.ValueString(),
		ServiceID:   plan.ServiceID.ValueString(),
	}

	createdMount, err := r.client.CreateMount(mount)
	if err != nil {
		resp.Diagnostics.AddError("Error creating mount", err.Error())
		return
	}

	plan.ID = types.StringValue(createdMount.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mount, err := r.client.GetMount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading mount", err.Error())
		return
	}

	state.Type = types.StringValue(mount.Type)
	state.HostPath = types.StringValue(mount.HostPath)
	state.VolumeName = types.StringValue(mount.VolumeName)
	state.MountPath = types.StringValue(mount.MountPath)
	state.ServiceType = types.StringValue(mount.ServiceType)
	state.FilePath = types.StringValue(mount.FilePath)
	state.ServiceID = types.StringValue(mount.ServiceID)
	// Don't update Content from API as it might not be returned

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mount := client.Mount{
		ID:          plan.ID.ValueString(),
		Type:        plan.Type.ValueString(),
		HostPath:    plan.HostPath.ValueString(),
		VolumeName:  plan.VolumeName.ValueString(),
		Content:     plan.Content.ValueString(),
		MountPath:   plan.MountPath.ValueString(),
		ServiceType: plan.ServiceType.ValueString(),
		FilePath:    plan.FilePath.ValueString(),
	}

	updatedMount, err := r.client.UpdateMount(mount)
	if err != nil {
		resp.Diagnostics.AddError("Error updating mount", err.Error())
		return
	}

	plan.Type = types.StringValue(updatedMount.Type)
	plan.HostPath = types.StringValue(updatedMount.HostPath)
	plan.VolumeName = types.StringValue(updatedMount.VolumeName)
	plan.MountPath = types.StringValue(updatedMount.MountPath)
	plan.ServiceType = types.StringValue(updatedMount.ServiceType)
	plan.FilePath = types.StringValue(updatedMount.FilePath)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMount(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting mount", err.Error())
		return
	}
}

func (r *MountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
