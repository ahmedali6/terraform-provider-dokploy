package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CertificatesDataSource{}

func NewCertificatesDataSource() datasource.DataSource {
	return &CertificatesDataSource{}
}

type CertificatesDataSource struct {
	client *client.DokployClient
}

type CertificatesDataSourceModel struct {
	Certificates []CertificateDataModel `tfsdk:"certificates"`
}

type CertificateDataModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CertificatePath types.String `tfsdk:"certificate_path"`
	AutoRenew       types.Bool   `tfsdk:"auto_renew"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	ServerID        types.String `tfsdk:"server_id"`
}

func (d *CertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *CertificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all TLS certificates in the current Dokploy organization.",
		Attributes: map[string]schema.Attribute{
			"certificates": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of certificates.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier for the certificate.",
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
				},
			},
		},
	}
}

func (d *CertificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	certs, err := d.client.ListCertificates()
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Certificates", err.Error())
		return
	}

	var state CertificatesDataSourceModel

	for _, cert := range certs {
		certModel := CertificateDataModel{
			ID:              types.StringValue(cert.ID),
			Name:            types.StringValue(cert.Name),
			CertificatePath: types.StringValue(cert.CertificatePath),
			OrganizationID:  types.StringValue(cert.OrganizationID),
		}

		if cert.AutoRenew != nil {
			certModel.AutoRenew = types.BoolValue(*cert.AutoRenew)
		}

		if cert.ServerID != nil && *cert.ServerID != "" {
			certModel.ServerID = types.StringValue(*cert.ServerID)
		}

		state.Certificates = append(state.Certificates, certModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
