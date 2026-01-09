package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BitbucketProvidersDataSource{}

func NewBitbucketProvidersDataSource() datasource.DataSource {
	return &BitbucketProvidersDataSource{}
}

type BitbucketProvidersDataSource struct {
	client *client.DokployClient
}

type BitbucketProvidersDataSourceModel struct {
	Providers []BitbucketProviderDataModel `tfsdk:"providers"`
}

type BitbucketProviderDataModel struct {
	ID             types.String `tfsdk:"id"`
	GitProviderId  types.String `tfsdk:"git_provider_id"`
	Name           types.String `tfsdk:"name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *BitbucketProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitbucket_providers"
}

func (d *BitbucketProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Bitbucket providers configured in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"providers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Bitbucket providers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier (bitbucketId) of the Bitbucket provider.",
						},
						"git_provider_id": schema.StringAttribute{
							Computed:    true,
							Description: "The git provider ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Bitbucket provider.",
						},
						"provider_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of provider (bitbucket).",
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

func (d *BitbucketProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BitbucketProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BitbucketProvidersDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	providers, err := d.client.ListBitbucketProviders()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Bitbucket Providers", err.Error())
		return
	}

	var state BitbucketProvidersDataSourceModel

	for _, provider := range providers {
		providerModel := BitbucketProviderDataModel{
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
