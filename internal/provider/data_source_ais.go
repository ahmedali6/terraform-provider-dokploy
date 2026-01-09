package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AIsDataSource{}

func NewAIsDataSource() datasource.DataSource {
	return &AIsDataSource{}
}

type AIsDataSource struct {
	client *client.DokployClient
}

type AIsDataSourceModel struct {
	AIs []AIDataModel `tfsdk:"ais"`
}

type AIDataModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ApiURL         types.String `tfsdk:"api_url"`
	Model          types.String `tfsdk:"model"`
	IsEnabled      types.Bool   `tfsdk:"is_enabled"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *AIsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ais"
}

func (d *AIsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all AI provider configurations in the current Dokploy organization.",
		Attributes: map[string]schema.Attribute{
			"ais": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of AI configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier for the AI configuration.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name for the AI provider configuration.",
						},
						"api_url": schema.StringAttribute{
							Computed:    true,
							Description: "The API endpoint URL for the AI provider.",
						},
						"model": schema.StringAttribute{
							Computed:    true,
							Description: "The model being used.",
						},
						"is_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the AI configuration is enabled.",
						},
						"organization_id": schema.StringAttribute{
							Computed:    true,
							Description: "The organization ID this AI configuration belongs to.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the AI configuration was created.",
						},
					},
				},
			},
		},
	}
}

func (d *AIsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AIsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ais, err := d.client.ListAIs()
	if err != nil {
		resp.Diagnostics.AddError("Unable to List AI Configurations", err.Error())
		return
	}

	var state AIsDataSourceModel

	for _, ai := range ais {
		aiModel := AIDataModel{
			ID:             types.StringValue(ai.ID),
			Name:           types.StringValue(ai.Name),
			ApiURL:         types.StringValue(ai.ApiURL),
			Model:          types.StringValue(ai.Model),
			IsEnabled:      types.BoolValue(ai.IsEnabled),
			OrganizationID: types.StringValue(ai.OrganizationID),
			CreatedAt:      types.StringValue(ai.CreatedAt),
		}
		state.AIs = append(state.AIs, aiModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
