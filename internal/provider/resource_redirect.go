package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ resource.Resource = &RedirectResource{}
var _ resource.ResourceWithImportState = &RedirectResource{}

func NewRedirectResource() resource.Resource {
	return &RedirectResource{}
}

type RedirectResource struct {
	client *client.DokployClient
}

type RedirectResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Regex         types.String `tfsdk:"regex"`
	Replacement   types.String `tfsdk:"replacement"`
	Permanent     types.Bool   `tfsdk:"permanent"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *RedirectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redirect"
}

func (r *RedirectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages URL redirects for a Dokploy application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the redirect.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"regex": schema.StringAttribute{
				Required:    true,
				Description: "Regular expression to match the URL.",
			},
			"replacement": schema.StringAttribute{
				Required:    true,
				Description: "Replacement URL pattern.",
			},
			"permanent": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the redirect is permanent (301) or temporary (302).",
				Default:     booldefault.StaticBool(true),
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RedirectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DokployClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *RedirectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedirectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirect := client.Redirect{
		Regex:         plan.Regex.ValueString(),
		Replacement:   plan.Replacement.ValueString(),
		Permanent:     plan.Permanent.ValueBool(),
		ApplicationID: plan.ApplicationID.ValueString(),
	}

	createdRedirect, err := r.client.CreateRedirect(redirect)
	if err != nil {
		resp.Diagnostics.AddError("Error creating redirect", err.Error())
		return
	}

	plan.ID = types.StringValue(createdRedirect.ID)
	plan.Permanent = types.BoolValue(createdRedirect.Permanent)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedirectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirect, err := r.client.GetRedirect(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading redirect", err.Error())
		return
	}

	state.Regex = types.StringValue(redirect.Regex)
	state.Replacement = types.StringValue(redirect.Replacement)
	state.Permanent = types.BoolValue(redirect.Permanent)
	state.ApplicationID = types.StringValue(redirect.ApplicationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedirectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirect := client.Redirect{
		ID:          plan.ID.ValueString(),
		Regex:       plan.Regex.ValueString(),
		Replacement: plan.Replacement.ValueString(),
		Permanent:   plan.Permanent.ValueBool(),
	}

	updatedRedirect, err := r.client.UpdateRedirect(redirect)
	if err != nil {
		resp.Diagnostics.AddError("Error updating redirect", err.Error())
		return
	}

	plan.Regex = types.StringValue(updatedRedirect.Regex)
	plan.Replacement = types.StringValue(updatedRedirect.Replacement)
	plan.Permanent = types.BoolValue(updatedRedirect.Permanent)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedirectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRedirect(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting redirect", err.Error())
		return
	}
}

func (r *RedirectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
