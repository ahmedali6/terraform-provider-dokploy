package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CertificateDataSource{}

func NewCertificateDataSource() datasource.DataSource {
	return &CertificateDataSource{}
}

type CertificateDataSource struct {
	client *client.DokployClient
}

type CertificateDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CertificatePath types.String `tfsdk:"certificate_path"`
	AutoRenew       types.Bool   `tfsdk:"auto_renew"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	ServerID        types.String `tfsdk:"server_id"`
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single TLS certificate by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the certificate.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Display name for the certificate.",
			},
			"certificate_path": schema.StringAttribute{
				Computed:    true,
				Description: "The path where the certificate is stored.",
			},
			"auto_renew": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the certificate is set to auto-renew.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organization ID this certificate belongs to.",
			},
			"server_id": schema.StringAttribute{
				Computed:    true,
				Description: "The server ID this certificate is associated with.",
			},
		},
	}
}

func (d *CertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificateDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := d.client.GetCertificate(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Certificate", err.Error())
		return
	}

	data.Name = types.StringValue(cert.Name)
	data.CertificatePath = types.StringValue(cert.CertificatePath)
	data.OrganizationID = types.StringValue(cert.OrganizationID)

	if cert.AutoRenew != nil {
		data.AutoRenew = types.BoolValue(*cert.AutoRenew)
	}

	if cert.ServerID != nil && *cert.ServerID != "" {
		data.ServerID = types.StringValue(*cert.ServerID)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
