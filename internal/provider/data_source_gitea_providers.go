package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GiteaProvidersDataSource{}

func NewGiteaProvidersDataSource() datasource.DataSource {
	return &GiteaProvidersDataSource{}
}

type GiteaProvidersDataSource struct {
	client *client.DokployClient
}

type GiteaProvidersDataSourceModel struct {
	Providers []GiteaProviderDataModel `tfsdk:"providers"`
}

type GiteaProviderDataModel struct {
	ID             types.String `tfsdk:"id"`
	GitProviderId  types.String `tfsdk:"git_provider_id"`
	Name           types.String `tfsdk:"name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *GiteaProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitea_providers"
}

func (d *GiteaProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Gitea providers configured in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"providers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Gitea providers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier (giteaId) of the Gitea provider.",
						},
						"git_provider_id": schema.StringAttribute{
							Computed:    true,
							Description: "The git provider ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Gitea provider.",
						},
						"provider_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of provider (gitea).",
						},
						"organization_id": schema.StringAttribute{
							Computed:    true,
							Description: "The Dokploy organization ID this provider belongs to.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The creation timestamp of the provider.",
						},
					},
				},
			},
		},
	}
}

func (d *GiteaProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GiteaProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GiteaProvidersDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providers, err := d.client.ListGiteaProviders()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gitea Providers", err.Error())
		return
	}

	var state GiteaProvidersDataSourceModel

	for _, provider := range providers {
		providerModel := GiteaProviderDataModel{
			ID:             types.StringValue(provider.ID),
			GitProviderId:  types.StringValue(provider.GitProvider.GitProviderId),
			Name:           types.StringValue(provider.GitProvider.Name),
			ProviderType:   types.StringValue(provider.GitProvider.ProviderType),
			OrganizationID: types.StringValue(provider.GitProvider.OrganizationID),
			CreatedAt:      types.StringValue(provider.GitProvider.CreatedAt),
		}
		state.Providers = append(state.Providers, providerModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
