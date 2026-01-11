package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DestinationsDataSource{}

func NewDestinationsDataSource() datasource.DataSource {
	return &DestinationsDataSource{}
}

type DestinationsDataSource struct {
	client *client.DokployClient
}

type DestinationsDataSourceModel struct {
	Destinations []DestinationModel `tfsdk:"destinations"`
}

type DestinationModel struct {
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

func (d *DestinationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destinations"
}

func (d *DestinationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Dokploy backup destinations.",
		Attributes: map[string]schema.Attribute{
			"destinations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of all backup destinations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *DestinationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DestinationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DestinationsDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinations, err := d.client.ListDestinations()
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Destinations", err.Error())
		return
	}

	// Map destinations
	data.Destinations = make([]DestinationModel, len(destinations))
	for i, dest := range destinations {
		data.Destinations[i] = DestinationModel{
			ID:              types.StringValue(dest.DestinationID),
			Name:            types.StringValue(dest.Name),
			StorageProvider: types.StringValue(dest.Provider),
			AccessKey:       types.StringValue(dest.AccessKey),
			Bucket:          types.StringValue(dest.Bucket),
			Region:          types.StringValue(dest.Region),
			Endpoint:        types.StringValue(dest.Endpoint),
			OrganizationID:  types.StringValue(dest.OrganizationID),
			CreatedAt:       types.StringValue(dest.CreatedAt),
		}

		// Handle nullable server_id
		if dest.ServerID != nil {
			data.Destinations[i].ServerID = types.StringValue(*dest.ServerID)
		} else {
			data.Destinations[i].ServerID = types.StringNull()
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
