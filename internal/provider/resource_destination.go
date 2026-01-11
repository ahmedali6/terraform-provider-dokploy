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

var _ resource.Resource = &DestinationResource{}
var _ resource.ResourceWithImportState = &DestinationResource{}

func NewDestinationResource() resource.Resource {
	return &DestinationResource{}
}

type DestinationResource struct {
	client *client.DokployClient
}

type DestinationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Bucket          types.String `tfsdk:"bucket"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	ServerID        types.String `tfsdk:"server_id"`
}

func (r *DestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backup destination (S3, MinIO, etc.) for Dokploy backups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the destination",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the destination",
			},
			"storage_provider": schema.StringAttribute{
				Required:    true,
				Description: "Storage provider type (e.g., 's3', 'minio')",
			},
			"access_key": schema.StringAttribute{
				Required:    true,
				Description: "Access key for the storage provider",
			},
			"secret_access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret access key for the storage provider",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Bucket name for storing backups",
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "Region where the bucket is located",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint URL for the storage provider",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID for remote server destinations",
			},
		},
	}
}

func (r *DestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest := client.Destination{
		Name:            plan.Name.ValueString(),
		Provider:        plan.StorageProvider.ValueString(),
		AccessKey:       plan.AccessKey.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Bucket:          plan.Bucket.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
	}
	if !plan.ServerID.IsNull() && !plan.ServerID.IsUnknown() {
		serverID := plan.ServerID.ValueString()
		dest.ServerID = &serverID
	}

	createdDest, err := r.client.CreateDestination(dest)
	if err != nil {
		resp.Diagnostics.AddError("Error creating destination", err.Error())
		return
	}

	plan.ID = types.StringValue(createdDest.DestinationID)
	plan.Name = types.StringValue(createdDest.Name)
	plan.StorageProvider = types.StringValue(createdDest.Provider)
	plan.AccessKey = types.StringValue(createdDest.AccessKey)
	plan.Bucket = types.StringValue(createdDest.Bucket)
	plan.Region = types.StringValue(createdDest.Region)
	plan.Endpoint = types.StringValue(createdDest.Endpoint)
	if createdDest.ServerID != nil {
		plan.ServerID = types.StringValue(*createdDest.ServerID)
	} else {
		plan.ServerID = types.StringNull()
	}
	// Don't update secret_access_key from response as it's sensitive

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := r.client.GetDestination(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading destination", err.Error())
		return
	}

	state.Name = types.StringValue(dest.Name)
	state.StorageProvider = types.StringValue(dest.Provider)
	state.AccessKey = types.StringValue(dest.AccessKey)
	state.Bucket = types.StringValue(dest.Bucket)
	state.Region = types.StringValue(dest.Region)
	state.Endpoint = types.StringValue(dest.Endpoint)
	if dest.ServerID != nil {
		state.ServerID = types.StringValue(*dest.ServerID)
	} else {
		state.ServerID = types.StringNull()
	}
	// Don't update secret_access_key from API response

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest := client.Destination{
		DestinationID:   plan.ID.ValueString(),
		Name:            plan.Name.ValueString(),
		Provider:        plan.StorageProvider.ValueString(),
		AccessKey:       plan.AccessKey.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Bucket:          plan.Bucket.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
	}
	// Handle ServerID updates explicitly:
	// - Unknown: do not modify dest.ServerID (no change requested)
	// - Non-null: set dest.ServerID to the provided value
	// - Null (known): explicitly set dest.ServerID to nil so the API can unset it
	if !plan.ServerID.IsUnknown() {
		if plan.ServerID.IsNull() {
			dest.ServerID = nil
		} else {
			serverID := plan.ServerID.ValueString()
			dest.ServerID = &serverID
		}
	}

	updatedDest, err := r.client.UpdateDestination(dest)
	if err != nil {
		resp.Diagnostics.AddError("Error updating destination", err.Error())
		return
	}

	plan.Name = types.StringValue(updatedDest.Name)
	plan.StorageProvider = types.StringValue(updatedDest.Provider)
	plan.AccessKey = types.StringValue(updatedDest.AccessKey)
	plan.Bucket = types.StringValue(updatedDest.Bucket)
	plan.Region = types.StringValue(updatedDest.Region)
	plan.Endpoint = types.StringValue(updatedDest.Endpoint)
	if updatedDest.ServerID != nil {
		plan.ServerID = types.StringValue(*updatedDest.ServerID)
	} else {
		plan.ServerID = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDestination(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting destination", err.Error())
		return
	}
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
