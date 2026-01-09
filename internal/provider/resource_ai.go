package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AIResource{}
var _ resource.ResourceWithImportState = &AIResource{}

func NewAIResource() resource.Resource {
	return &AIResource{}
}

type AIResource struct {
	client *client.DokployClient
}

type AIResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ApiURL         types.String `tfsdk:"api_url"`
	ApiKey         types.String `tfsdk:"api_key"`
	Model          types.String `tfsdk:"model"`
	IsEnabled      types.Bool   `tfsdk:"is_enabled"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *AIResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai"
}

func (r *AIResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AI provider configuration in Dokploy. This allows integration with AI services like OpenAI for suggestions and deployments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the AI configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the AI provider configuration.",
			},
			"api_url": schema.StringAttribute{
				Required:    true,
				Description: "The API endpoint URL for the AI provider (e.g., https://api.openai.com/v1).",
			},
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "API key for authenticating with the AI provider.",
			},
			"model": schema.StringAttribute{
				Required:    true,
				Description: "The model to use (e.g., gpt-4, gpt-4o, gpt-3.5-turbo).",
			},
			"is_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the AI configuration is enabled. Defaults to true.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organization ID this AI configuration belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the AI configuration was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *AIResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AIResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AIResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ai, err := r.client.CreateAI(
		plan.Name.ValueString(),
		plan.ApiURL.ValueString(),
		plan.ApiKey.ValueString(),
		plan.Model.ValueString(),
		plan.IsEnabled.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating AI configuration", err.Error())
		return
	}

	plan.ID = types.StringValue(ai.ID)
	plan.Name = types.StringValue(ai.Name)
	plan.ApiURL = types.StringValue(ai.ApiURL)
	plan.Model = types.StringValue(ai.Model)
	plan.IsEnabled = types.BoolValue(ai.IsEnabled)
	plan.OrganizationID = types.StringValue(ai.OrganizationID)
	plan.CreatedAt = types.StringValue(ai.CreatedAt)
	// Keep the api_key from plan since API returns it

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AIResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AIResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ai, err := r.client.GetAI(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading AI configuration", err.Error())
		return
	}

	state.Name = types.StringValue(ai.Name)
	state.ApiURL = types.StringValue(ai.ApiURL)
	state.Model = types.StringValue(ai.Model)
	state.IsEnabled = types.BoolValue(ai.IsEnabled)
	state.OrganizationID = types.StringValue(ai.OrganizationID)
	state.CreatedAt = types.StringValue(ai.CreatedAt)
	// Note: api_key is returned by API but we preserve the state value

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *AIResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AIResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AIResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// API requires all fields for update
	ai := client.AI{
		ID:        state.ID.ValueString(),
		Name:      plan.Name.ValueString(),
		ApiURL:    plan.ApiURL.ValueString(),
		ApiKey:    plan.ApiKey.ValueString(),
		Model:     plan.Model.ValueString(),
		IsEnabled: plan.IsEnabled.ValueBool(),
	}

	err := r.client.UpdateAI(ai)
	if err != nil {
		resp.Diagnostics.AddError("Error updating AI configuration", err.Error())
		return
	}

	// Read back the updated AI
	updatedAI, err := r.client.GetAI(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading AI configuration after update", err.Error())
		return
	}

	plan.ID = types.StringValue(updatedAI.ID)
	plan.Name = types.StringValue(updatedAI.Name)
	plan.ApiURL = types.StringValue(updatedAI.ApiURL)
	plan.Model = types.StringValue(updatedAI.Model)
	plan.IsEnabled = types.BoolValue(updatedAI.IsEnabled)
	plan.OrganizationID = types.StringValue(updatedAI.OrganizationID)
	plan.CreatedAt = types.StringValue(updatedAI.CreatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AIResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AIResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAI(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting AI configuration", err.Error())
		return
	}
}

func (r *AIResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Warn that api_key won't be imported properly
	resp.Diagnostics.AddWarning(
		"API Key Required After Import",
		"After importing, you must set the 'api_key' attribute in your configuration. "+
			"The API key value cannot be retrieved from the server for security reasons.",
	)
}
