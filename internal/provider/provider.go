package provider

import (
	"context"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DokployProvider{}
var _ provider.ProviderWithFunctions = &DokployProvider{}

type DokployProvider struct {
	version string
}

type DokployProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *DokployProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dokploy"
	resp.Version = p.version
}

func (p *DokployProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "The URL of your Dokploy instance (e.g., https://dokploy.example.com/api)",
			},
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Your Dokploy API Key",
			},
		},
	}
}

func (p *DokployProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DokployProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Missing Host Configuration",
			"While configuring the provider, the Host was unknown. This can happen when the value is not yet known.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Missing API Key Configuration",
			"While configuring the provider, the API Key was unknown. This can happen when the value is not yet known.",
		)
	}

	if config.Host.IsNull() || config.ApiKey.IsNull() {
		return
	}

	// Create client
	c := client.NewDokployClient(config.Host.ValueString(), config.ApiKey.ValueString())

	// Make client available to resources
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *DokployProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewEnvironmentResource,
		NewApplicationResource,
		NewComposeResource,
		NewDatabaseResource,
		NewDomainResource,
		NewEnvironmentVariablesResource,
		NewSSHKeyResource,
		NewMountResource,
		NewPortResource,
		NewRedirectResource,
		NewRegistryResource,
		NewDestinationResource,
		NewBackupResource,
	}
}

func (p *DokployProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServersDataSource,
		NewGithubProvidersDataSource,
	}
}

func (p *DokployProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DokployProvider{version: version}
	}
}
