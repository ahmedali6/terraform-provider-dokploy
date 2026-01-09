package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GiteaProviderResource{}
var _ resource.ResourceWithImportState = &GiteaProviderResource{}

func NewGiteaProviderResource() resource.Resource {
	return &GiteaProviderResource{}
}

type GiteaProviderResource struct {
	client *client.DokployClient
}

type GiteaProviderResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	GitProviderId       types.String `tfsdk:"git_provider_id"`
	Name                types.String `tfsdk:"name"`
	GiteaUrl            types.String `tfsdk:"gitea_url"`
	RedirectUri         types.String `tfsdk:"redirect_uri"`
	ClientId            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	AccessToken         types.String `tfsdk:"access_token"`
	RefreshToken        types.String `tfsdk:"refresh_token"`
	ExpiresAt           types.Int64  `tfsdk:"expires_at"`
	Scopes              types.String `tfsdk:"scopes"`
	LastAuthenticatedAt types.Int64  `tfsdk:"last_authenticated_at"`
	GiteaUsername       types.String `tfsdk:"gitea_username"`
	OrganizationName    types.String `tfsdk:"organization_name"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

func (r *GiteaProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitea_provider"
}

func (r *GiteaProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Gitea provider integration in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the Gitea provider (giteaId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_provider_id": schema.StringAttribute{
				Computed:    true,
				Description: "The git provider ID used for deletion.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Gitea provider.",
			},
			"gitea_url": schema.StringAttribute{
				Required:    true,
				Description: "The Gitea instance URL (e.g., https://gitea.example.com).",
			},
			"redirect_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth redirect URI.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Gitea OAuth client ID.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Gitea OAuth client secret.",
			},
			"access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The OAuth access token.",
			},
			"refresh_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The OAuth refresh token.",
			},
			"expires_at": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Token expiration timestamp.",
			},
			"scopes": schema.StringAttribute{
				Optional:    true,
				Description: "OAuth scopes.",
			},
			"last_authenticated_at": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Last authentication timestamp.",
			},
			"gitea_username": schema.StringAttribute{
				Optional:    true,
				Description: "The Gitea username.",
			},
			"organization_name": schema.StringAttribute{
				Optional:    true,
				Description: "The Gitea organization name.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Dokploy organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GiteaProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GiteaProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GiteaProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.GiteaProvider{
		Name:                plan.Name.ValueString(),
		GiteaUrl:            plan.GiteaUrl.ValueString(),
		RedirectUri:         plan.RedirectUri.ValueString(),
		ClientId:            plan.ClientId.ValueString(),
		ClientSecret:        plan.ClientSecret.ValueString(),
		AccessToken:         plan.AccessToken.ValueString(),
		RefreshToken:        plan.RefreshToken.ValueString(),
		ExpiresAt:           plan.ExpiresAt.ValueInt64(),
		Scopes:              plan.Scopes.ValueString(),
		LastAuthenticatedAt: plan.LastAuthenticatedAt.ValueInt64(),
		GiteaUsername:       plan.GiteaUsername.ValueString(),
		OrganizationName:    plan.OrganizationName.ValueString(),
	}

	created, err := r.client.CreateGiteaProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Gitea provider", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.GitProviderId = types.StringValue(created.GitProviderId)
	plan.OrganizationID = types.StringValue(created.OrganizationID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	if created.ExpiresAt != 0 {
		plan.ExpiresAt = types.Int64Value(created.ExpiresAt)
	}
	if created.LastAuthenticatedAt != 0 {
		plan.LastAuthenticatedAt = types.Int64Value(created.LastAuthenticatedAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GiteaProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.client.GetGiteaProvider(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Gitea provider", err.Error())
		return
	}

	state.ID = types.StringValue(provider.ID)
	state.GitProviderId = types.StringValue(provider.GitProviderId)
	state.Name = types.StringValue(provider.Name)
	state.GiteaUrl = types.StringValue(provider.GiteaUrl)
	state.OrganizationID = types.StringValue(provider.OrganizationID)
	state.CreatedAt = types.StringValue(provider.CreatedAt)

	if provider.RedirectUri != "" {
		state.RedirectUri = types.StringValue(provider.RedirectUri)
	}
	if provider.ClientId != "" {
		state.ClientId = types.StringValue(provider.ClientId)
	}
	if provider.Scopes != "" {
		state.Scopes = types.StringValue(provider.Scopes)
	}
	if provider.GiteaUsername != "" {
		state.GiteaUsername = types.StringValue(provider.GiteaUsername)
	}
	if provider.OrganizationName != "" {
		state.OrganizationName = types.StringValue(provider.OrganizationName)
	}
	if provider.ExpiresAt != 0 {
		state.ExpiresAt = types.Int64Value(provider.ExpiresAt)
	}
	if provider.LastAuthenticatedAt != 0 {
		state.LastAuthenticatedAt = types.Int64Value(provider.LastAuthenticatedAt)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GiteaProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state GiteaProviderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.GiteaProvider{
		ID:                  state.ID.ValueString(),
		GitProviderId:       state.GitProviderId.ValueString(),
		Name:                plan.Name.ValueString(),
		GiteaUrl:            plan.GiteaUrl.ValueString(),
		RedirectUri:         plan.RedirectUri.ValueString(),
		ClientId:            plan.ClientId.ValueString(),
		ClientSecret:        plan.ClientSecret.ValueString(),
		AccessToken:         plan.AccessToken.ValueString(),
		RefreshToken:        plan.RefreshToken.ValueString(),
		ExpiresAt:           plan.ExpiresAt.ValueInt64(),
		Scopes:              plan.Scopes.ValueString(),
		LastAuthenticatedAt: plan.LastAuthenticatedAt.ValueInt64(),
		GiteaUsername:       plan.GiteaUsername.ValueString(),
		OrganizationName:    plan.OrganizationName.ValueString(),
	}

	updated, err := r.client.UpdateGiteaProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Gitea provider", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.GitProviderId = types.StringValue(updated.GitProviderId)
	plan.OrganizationID = types.StringValue(updated.OrganizationID)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	if updated.ExpiresAt != 0 {
		plan.ExpiresAt = types.Int64Value(updated.ExpiresAt)
	}
	if updated.LastAuthenticatedAt != 0 {
		plan.LastAuthenticatedAt = types.Int64Value(updated.LastAuthenticatedAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GiteaProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use gitProviderId for deletion
	gitProviderId := state.GitProviderId.ValueString()
	if gitProviderId == "" {
		resp.Diagnostics.AddError("Error deleting Gitea provider", "gitProviderId is not set")
		return
	}

	err := r.client.DeleteGitProvider(gitProviderId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Gitea provider", err.Error())
		return
	}
}

func (r *GiteaProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
