package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	client *client.DokployClient
}

type CertificateResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CertificateData types.String `tfsdk:"certificate_data"`
	PrivateKey      types.String `tfsdk:"private_key"`
	CertificatePath types.String `tfsdk:"certificate_path"`
	AutoRenew       types.Bool   `tfsdk:"auto_renew"`
	ServerID        types.String `tfsdk:"server_id"`
}

func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TLS certificate in Dokploy. Certificates are used to provide HTTPS for your applications via Traefik.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_data": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The PEM-encoded certificate data.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The PEM-encoded private key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The path where the certificate is stored. Auto-generated if not provided.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_renew": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the certificate should be auto-renewed.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "The server ID to associate this certificate with. If not provided, uses the default server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Auto-fetch organization ID from current user
	orgID, err := r.client.GetCurrentOrganizationID()
	if err != nil {
		resp.Diagnostics.AddError("Error fetching organization ID", err.Error())
		return
	}

	cert := client.Certificate{
		Name:            plan.Name.ValueString(),
		CertificateData: plan.CertificateData.ValueString(),
		PrivateKey:      plan.PrivateKey.ValueString(),
		OrganizationID:  orgID,
	}

	if !plan.CertificatePath.IsNull() && !plan.CertificatePath.IsUnknown() {
		cert.CertificatePath = plan.CertificatePath.ValueString()
	}

	if !plan.AutoRenew.IsNull() && !plan.AutoRenew.IsUnknown() {
		autoRenew := plan.AutoRenew.ValueBool()
		cert.AutoRenew = &autoRenew
	}

	if !plan.ServerID.IsNull() && !plan.ServerID.IsUnknown() {
		serverID := plan.ServerID.ValueString()
		cert.ServerID = &serverID
	}

	created, err := r.client.CreateCertificate(cert)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.CertificatePath = types.StringValue(created.CertificatePath)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.GetCertificate(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	state.Name = types.StringValue(cert.Name)
	state.CertificatePath = types.StringValue(cert.CertificatePath)
	// Note: certificate_data and private_key are preserved from state since API returns them
	// but we want to keep the original values to avoid unnecessary diffs

	if cert.AutoRenew != nil {
		state.AutoRenew = types.BoolValue(*cert.AutoRenew)
	}

	if cert.ServerID != nil && *cert.ServerID != "" {
		state.ServerID = types.StringValue(*cert.ServerID)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Certificates are immutable - all changes require replacement
	// This method should never be called due to RequiresReplace on all attributes
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Certificates cannot be updated. Any changes require creating a new certificate.",
	)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCertificate(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
		return
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	resp.Diagnostics.AddWarning(
		"Sensitive Data Required After Import",
		"After importing, you must set the 'certificate_data' and 'private_key' attributes in your configuration. "+
			"These values cannot be securely retrieved from the server.",
	)
}
