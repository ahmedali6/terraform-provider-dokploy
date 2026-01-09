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

var _ resource.Resource = &BitbucketProviderResource{}
var _ resource.ResourceWithImportState = &BitbucketProviderResource{}

func NewBitbucketProviderResource() resource.Resource {
	return &BitbucketProviderResource{}
}

type BitbucketProviderResource struct {
	client *client.DokployClient
}

type BitbucketProviderResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	GitProviderId          types.String `tfsdk:"git_provider_id"`
	Name                   types.String `tfsdk:"name"`
	BitbucketUsername      types.String `tfsdk:"bitbucket_username"`
	BitbucketEmail         types.String `tfsdk:"bitbucket_email"`
	AppPassword            types.String `tfsdk:"app_password"`
	ApiToken               types.String `tfsdk:"api_token"`
	BitbucketWorkspaceName types.String `tfsdk:"bitbucket_workspace_name"`
	AuthId                 types.String `tfsdk:"auth_id"`
	OrganizationID         types.String `tfsdk:"organization_id"`
	CreatedAt              types.String `tfsdk:"created_at"`
}

func (r *BitbucketProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitbucket_provider"
}

func (r *BitbucketProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket provider integration in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the Bitbucket provider (bitbucketId).",
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
				Description: "The name of the Bitbucket provider.",
			},
			"bitbucket_username": schema.StringAttribute{
				Optional:    true,
				Description: "The Bitbucket username.",
			},
			"bitbucket_email": schema.StringAttribute{
				Optional:    true,
				Description: "The Bitbucket email address.",
			},
			"app_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Bitbucket app password for authentication.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Bitbucket API token.",
			},
			"bitbucket_workspace_name": schema.StringAttribute{
				Optional:    true,
				Description: "The Bitbucket workspace name.",
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

func (r *BitbucketProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BitbucketProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BitbucketProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.BitbucketProvider{
		Name:                   plan.Name.ValueString(),
		AuthId:                 plan.AuthId.ValueString(),
		BitbucketUsername:      plan.BitbucketUsername.ValueString(),
		BitbucketEmail:         plan.BitbucketEmail.ValueString(),
		AppPassword:            plan.AppPassword.ValueString(),
		ApiToken:               plan.ApiToken.ValueString(),
		BitbucketWorkspaceName: plan.BitbucketWorkspaceName.ValueString(),
	}

	created, err := r.client.CreateBitbucketProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Bitbucket provider", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.GitProviderId = types.StringValue(created.GitProviderId)
	plan.OrganizationID = types.StringValue(created.OrganizationID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BitbucketProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := r.client.GetBitbucketProvider(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Bitbucket provider", err.Error())
		return
	}

	state.ID = types.StringValue(provider.ID)
	state.GitProviderId = types.StringValue(provider.GitProviderId)
	state.Name = types.StringValue(provider.Name)
	state.OrganizationID = types.StringValue(provider.OrganizationID)
	state.CreatedAt = types.StringValue(provider.CreatedAt)

	if provider.BitbucketUsername != "" {
		state.BitbucketUsername = types.StringValue(provider.BitbucketUsername)
	}
	if provider.BitbucketEmail != "" {
		state.BitbucketEmail = types.StringValue(provider.BitbucketEmail)
	}
	if provider.BitbucketWorkspaceName != "" {
		state.BitbucketWorkspaceName = types.StringValue(provider.BitbucketWorkspaceName)
	}
	if provider.AuthId != "" {
		state.AuthId = types.StringValue(provider.AuthId)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BitbucketProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state BitbucketProviderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := client.BitbucketProvider{
		ID:                     state.ID.ValueString(),
		GitProviderId:          state.GitProviderId.ValueString(),
		Name:                   plan.Name.ValueString(),
		BitbucketUsername:      plan.BitbucketUsername.ValueString(),
		BitbucketEmail:         plan.BitbucketEmail.ValueString(),
		AppPassword:            plan.AppPassword.ValueString(),
		ApiToken:               plan.ApiToken.ValueString(),
		BitbucketWorkspaceName: plan.BitbucketWorkspaceName.ValueString(),
		AuthId:                 plan.AuthId.ValueString(),
	}

	updated, err := r.client.UpdateBitbucketProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Bitbucket provider", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.GitProviderId = types.StringValue(updated.GitProviderId)
	plan.OrganizationID = types.StringValue(updated.OrganizationID)
	plan.CreatedAt = types.StringValue(updated.CreatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BitbucketProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use gitProviderId for deletion
	gitProviderId := state.GitProviderId.ValueString()
	if gitProviderId == "" {
		resp.Diagnostics.AddError("Error deleting Bitbucket provider", "gitProviderId is not set")
		return
	}

	err := r.client.DeleteGitProvider(gitProviderId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Bitbucket provider", err.Error())
		return
	}
}

func (r *BitbucketProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
