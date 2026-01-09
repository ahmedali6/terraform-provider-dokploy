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

var _ resource.Resource = &GitlabProviderResource{}
var _ resource.ResourceWithImportState = &GitlabProviderResource{}

func NewGitlabProviderResource() resource.Resource {
	return &GitlabProviderResource{}
}

type GitlabProviderResource struct {
	client *client.DokployClient
}

type GitlabProviderResourceModel struct {
	ID             types.String `tfsdk:"id"`
	GitProviderId  types.String `tfsdk:"git_provider_id"`
	Name           types.String `tfsdk:"name"`
	GitlabUrl      types.String `tfsdk:"gitlab_url"`
	ApplicationId  types.String `tfsdk:"application_id"`
	RedirectUri    types.String `tfsdk:"redirect_uri"`
	Secret         types.String `tfsdk:"secret"`
	AccessToken    types.String `tfsdk:"access_token"`
	RefreshToken   types.String `tfsdk:"refresh_token"`
	GroupName      types.String `tfsdk:"group_name"`
	ExpiresAt      types.Int64  `tfsdk:"expires_at"`
	AuthId         types.String `tfsdk:"auth_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *GitlabProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_provider"
}

func (r *GitlabProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GitLab provider integration in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the GitLab provider (gitlabId).",
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
				Description: "The name of the GitLab provider.",
			},
			"gitlab_url": schema.StringAttribute{
				Required:    true,
				Description: "The GitLab instance URL (e.g., https://gitlab.com).",
			},
			"application_id": schema.StringAttribute{
				Optional:    true,
				Description: "The GitLab OAuth application ID.",
			},
			"redirect_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The OAuth redirect URI.",
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The GitLab OAuth application secret.",
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
			"group_name": schema.StringAttribute{
				Optional:    true,
				Description: "The GitLab group name to limit access.",
			},
			"expires_at": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Token expiration timestamp.",
			},
			"auth_id": schema.StringAttribute{
				Required:    true,
				Description: "The authentication ID (usually user ID from Dokploy).",
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

func (r *GitlabProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GitlabProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.GitlabProvider{
		Name:          plan.Name.ValueString(),
		GitlabUrl:     plan.GitlabUrl.ValueString(),
		AuthId:        plan.AuthId.ValueString(),
		ApplicationId: plan.ApplicationId.ValueString(),
		RedirectUri:   plan.RedirectUri.ValueString(),
		Secret:        plan.Secret.ValueString(),
		AccessToken:   plan.AccessToken.ValueString(),
		RefreshToken:  plan.RefreshToken.ValueString(),
		GroupName:     plan.GroupName.ValueString(),
		ExpiresAt:     plan.ExpiresAt.ValueInt64(),
	}

	created, err := r.client.CreateGitlabProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error creating GitLab provider", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.GitProviderId = types.StringValue(created.GitProviderId)
	plan.OrganizationID = types.StringValue(created.OrganizationID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	if created.ExpiresAt != 0 {
		plan.ExpiresAt = types.Int64Value(created.ExpiresAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.client.GetGitlabProvider(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading GitLab provider", err.Error())
		return
	}

	state.ID = types.StringValue(provider.ID)
	state.GitProviderId = types.StringValue(provider.GitProviderId)
	state.Name = types.StringValue(provider.Name)
	state.GitlabUrl = types.StringValue(provider.GitlabUrl)
	state.OrganizationID = types.StringValue(provider.OrganizationID)
	state.CreatedAt = types.StringValue(provider.CreatedAt)

	if provider.ApplicationId != "" {
		state.ApplicationId = types.StringValue(provider.ApplicationId)
	}
	if provider.RedirectUri != "" {
		state.RedirectUri = types.StringValue(provider.RedirectUri)
	}
	if provider.GroupName != "" {
		state.GroupName = types.StringValue(provider.GroupName)
	}
	if provider.ExpiresAt != 0 {
		state.ExpiresAt = types.Int64Value(provider.ExpiresAt)
	}
	if provider.AuthId != "" {
		state.AuthId = types.StringValue(provider.AuthId)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GitlabProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state GitlabProviderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.GitlabProvider{
		ID:            state.ID.ValueString(),
		GitProviderId: state.GitProviderId.ValueString(),
		Name:          plan.Name.ValueString(),
		GitlabUrl:     plan.GitlabUrl.ValueString(),
		ApplicationId: plan.ApplicationId.ValueString(),
		RedirectUri:   plan.RedirectUri.ValueString(),
		Secret:        plan.Secret.ValueString(),
		AccessToken:   plan.AccessToken.ValueString(),
		RefreshToken:  plan.RefreshToken.ValueString(),
		GroupName:     plan.GroupName.ValueString(),
		ExpiresAt:     plan.ExpiresAt.ValueInt64(),
		AuthId:        plan.AuthId.ValueString(),
	}

	updated, err := r.client.UpdateGitlabProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error updating GitLab provider", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.GitProviderId = types.StringValue(updated.GitProviderId)
	plan.OrganizationID = types.StringValue(updated.OrganizationID)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	if updated.ExpiresAt != 0 {
		plan.ExpiresAt = types.Int64Value(updated.ExpiresAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use gitProviderId for deletion
	gitProviderId := state.GitProviderId.ValueString()
	if gitProviderId == "" {
		resp.Diagnostics.AddError("Error deleting GitLab provider", "gitProviderId is not set")
		return
	}

	err := r.client.DeleteGitProvider(gitProviderId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting GitLab provider", err.Error())
		return
	}
}

func (r *GitlabProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
