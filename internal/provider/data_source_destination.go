package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DestinationDataSource{}

func NewDestinationDataSource() datasource.DataSource {
	return &DestinationDataSource{}
}

type DestinationDataSource struct {
	client *client.DokployClient
}

type DestinationDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	AccessKey       types.String `tfsdk:"access_key"`
	Bucket          types.String `tfsdk:"bucket"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	ServerID        types.String `tfsdk:"server_id"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (d *DestinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (d *DestinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Dokploy backup destination by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the destination.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the destination.",
			},
			"storage_provider": schema.StringAttribute{
				Computed:    true,
				Description: "Storage provider type (e.g., 'aws-s3', 'minio').",
			},
			"access_key": schema.StringAttribute{
				Computed:    true,
				Description: "Access key for the storage provider.",
			},
			"bucket": schema.StringAttribute{
				Computed:    true,
				Description: "Bucket name for storing backups.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Region where the bucket is located.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Endpoint URL for the storage provider.",
			},
			"server_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server ID for remote server destinations.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization ID the destination belongs to.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the destination was created.",
			},
		},
	}
}

func (d *DestinationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *DestinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DestinationDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := d.client.GetDestination(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	// Map destination data
	data.Name = types.StringValue(dest.Name)
	data.StorageProvider = types.StringValue(dest.Provider)
	data.AccessKey = types.StringValue(dest.AccessKey)
	data.Bucket = types.StringValue(dest.Bucket)
	data.Region = types.StringValue(dest.Region)
	data.Endpoint = types.StringValue(dest.Endpoint)
	data.OrganizationID = types.StringValue(dest.OrganizationID)
	data.CreatedAt = types.StringValue(dest.CreatedAt)

	// Handle nullable server_id
	if dest.ServerID != nil {
		data.ServerID = types.StringValue(*dest.ServerID)
	} else {
		data.ServerID = types.StringNull()
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
