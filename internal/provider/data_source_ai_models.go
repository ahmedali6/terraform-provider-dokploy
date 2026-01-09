package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AIModelsDataSource{}

func NewAIModelsDataSource() datasource.DataSource {
	return &AIModelsDataSource{}
}

type AIModelsDataSource struct {
	client *client.DokployClient
}

type AIModelsDataSourceModel struct {
	ApiURL types.String       `tfsdk:"api_url"`
	ApiKey types.String       `tfsdk:"api_key"`
	Models []AIModelDataModel `tfsdk:"models"`
}

type AIModelDataModel struct {
	ID      types.String `tfsdk:"id"`
	Object  types.String `tfsdk:"object"`
	Created types.Int64  `tfsdk:"created"`
	OwnedBy types.String `tfsdk:"owned_by"`
}

func (d *AIModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_models"
}

func (d *AIModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches available models from an AI provider. Useful for discovering which models are available before creating an AI configuration.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Required:    true,
				Description: "The API endpoint URL for the AI provider (e.g., https://api.openai.com/v1).",
			},
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "API key for authenticating with the AI provider.",
			},
			"models": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available models from the AI provider.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The model identifier (e.g., gpt-4, gpt-4o).",
						},
						"object": schema.StringAttribute{
							Computed:    true,
							Description: "The object type (usually 'model').",
						},
						"created": schema.Int64Attribute{
							Computed:    true,
							Description: "Unix timestamp when the model was created.",
						},
						"owned_by": schema.StringAttribute{
							Computed:    true,
							Description: "The organization that owns the model.",
						},
					},
				},
			},
		},
	}
}

func (d *AIModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AIModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AIModelsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	models, err := d.client.GetAIModels(config.ApiURL.ValueString(), config.ApiKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get AI Models", err.Error())
		return
	}

	var state AIModelsDataSourceModel
	state.ApiURL = config.ApiURL
	state.ApiKey = config.ApiKey

	for _, model := range models {
		modelData := AIModelDataModel{
			ID:      types.StringValue(model.ID),
			Object:  types.StringValue(model.Object),
			Created: types.Int64Value(model.Created),
			OwnedBy: types.StringValue(model.OwnedBy),
		}
		state.Models = append(state.Models, modelData)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
