package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DomainResource{}
var _ resource.ResourceWithImportState = &DomainResource{}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	client *client.DokployClient
}

type DomainResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ApplicationID     types.String `tfsdk:"application_id"`
	ComposeID         types.String `tfsdk:"compose_id"`
	ServiceName       types.String `tfsdk:"service_name"`
	Host              types.String `tfsdk:"host"`
	Path              types.String `tfsdk:"path"`
	Port              types.Int64  `tfsdk:"port"`
	HTTPS             types.Bool   `tfsdk:"https"`
	CertificateType   types.String `tfsdk:"certificate_type"`
	GenerateTraefikMe types.Bool   `tfsdk:"generate_traefik_me"`
	RedeployOnUpdate  types.Bool   `tfsdk:"redeploy_on_update"`
}

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"host": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"https": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable HTTPS for the domain.",
			},
			"certificate_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Certificate type: 'none', 'letsencrypt'. Defaults to 'letsencrypt' when https is true.",
			},
			"generate_traefik_me": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, generates a traefik.me domain for the application.",
			},
			"redeploy_on_update": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, triggers a redeploy of the associated application or compose stack when the domain is created or updated.",
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ApplicationID.IsNull() && plan.ComposeID.IsNull() {
		resp.Diagnostics.AddError("Missing Association", "Either application_id or compose_id must be provided")
		return
	}

	// Logic for domain generation
	if !plan.GenerateTraefikMe.IsNull() && plan.GenerateTraefikMe.ValueBool() {
		var name string
		if !plan.ApplicationID.IsNull() {
			app, err := r.client.GetApplication(plan.ApplicationID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error fetching application for domain generation", err.Error())
				return
			}
			name = app.Name
		} else {
			comp, err := r.client.GetCompose(plan.ComposeID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error fetching compose for domain generation", err.Error())
				return
			}
			name = comp.Name
		}

		generatedDomain, err := r.client.GenerateDomain(name)
		if err != nil {
			resp.Diagnostics.AddError("Error generating traefik.me domain", err.Error())
			return
		}
		plan.Host = types.StringValue(generatedDomain)
	} else {
		if plan.Host.IsNull() || plan.Host.IsUnknown() {
			resp.Diagnostics.AddError("Missing Host", "Host is required when generate_traefik_me is false")
			return
		}
	}

	// Apply defaults
	if plan.Path.IsUnknown() || plan.Path.IsNull() {
		plan.Path = types.StringValue("/")
	}
	if plan.Port.IsUnknown() || plan.Port.IsNull() {
		plan.Port = types.Int64Value(3000)
	}
	if plan.HTTPS.IsUnknown() || plan.HTTPS.IsNull() {
		plan.HTTPS = types.BoolValue(true)
	}

	domain := client.Domain{
		ApplicationID:   plan.ApplicationID.ValueString(),
		ComposeID:       plan.ComposeID.ValueString(),
		ServiceName:     plan.ServiceName.ValueString(),
		Host:            plan.Host.ValueString(),
		Path:            plan.Path.ValueString(),
		Port:            plan.Port.ValueInt64(),
		HTTPS:           plan.HTTPS.ValueBool(),
		CertificateType: plan.CertificateType.ValueString(),
	}

	createdDomain, err := r.client.CreateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", err.Error())
		return
	}

	plan.ID = types.StringValue(createdDomain.ID)
	plan.ServiceName = types.StringValue(createdDomain.ServiceName)
	plan.CertificateType = types.StringValue(createdDomain.CertificateType)

	// Trigger Redeploy if requested
	if !plan.RedeployOnUpdate.IsNull() && plan.RedeployOnUpdate.ValueBool() {
		if !plan.ApplicationID.IsNull() {
			_ = r.client.DeployApplication(plan.ApplicationID.ValueString(), "")
		} else if !plan.ComposeID.IsNull() {
			_ = r.client.DeployCompose(plan.ComposeID.ValueString(), "")
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var domains []client.Domain
	var err error
	if !state.ApplicationID.IsNull() {
		domains, err = r.client.GetDomainsByApplication(state.ApplicationID.ValueString())
	} else {
		domains, err = r.client.GetDomainsByCompose(state.ComposeID.ValueString())
	}

	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading domains", err.Error())
		return
	}

	found := false
	for _, d := range domains {
		if d.ID == state.ID.ValueString() {
			state.Host = types.StringValue(d.Host)
			state.Path = types.StringValue(d.Path)
			state.Port = types.Int64Value(d.Port)
			state.HTTPS = types.BoolValue(d.HTTPS)
			state.ServiceName = types.StringValue(d.ServiceName)
			state.CertificateType = types.StringValue(d.CertificateType)
			if d.ApplicationID != "" {
				state.ApplicationID = types.StringValue(d.ApplicationID)
			}
			if d.ComposeID != "" {
				state.ComposeID = types.StringValue(d.ComposeID)
			}
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := client.Domain{
		ID:              plan.ID.ValueString(),
		ApplicationID:   plan.ApplicationID.ValueString(),
		ComposeID:       plan.ComposeID.ValueString(),
		ServiceName:     plan.ServiceName.ValueString(),
		Host:            plan.Host.ValueString(),
		Path:            plan.Path.ValueString(),
		Port:            plan.Port.ValueInt64(),
		HTTPS:           plan.HTTPS.ValueBool(),
		CertificateType: plan.CertificateType.ValueString(),
	}

	updatedDomain, err := r.client.UpdateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain", err.Error())
		return
	}

	plan.Host = types.StringValue(updatedDomain.Host)
	plan.Path = types.StringValue(updatedDomain.Path)
	plan.Port = types.Int64Value(updatedDomain.Port)
	plan.HTTPS = types.BoolValue(updatedDomain.HTTPS)
	plan.ServiceName = types.StringValue(updatedDomain.ServiceName)
	plan.CertificateType = types.StringValue(updatedDomain.CertificateType)

	// Trigger Redeploy if requested
	if !plan.RedeployOnUpdate.IsNull() && plan.RedeployOnUpdate.ValueBool() {
		if !plan.ApplicationID.IsNull() {
			_ = r.client.DeployApplication(plan.ApplicationID.ValueString(), "")
		} else if !plan.ComposeID.IsNull() {
			_ = r.client.DeployCompose(plan.ComposeID.ValueString(), "")
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting domain", err.Error())
		return
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <id> or <type>:<parent-id>:<id>
	// Where type is "application" or "compose"
	importID := req.ID

	// Just set the ID and let the Read function fail if we can't find it
	// The Read function will need to be updated to handle this case
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)

	// We can't determine application_id or compose_id from just the domain ID
	// The user should import using format: application:<app-id>:<domain-id> or compose:<compose-id>:<domain-id>
	// For now, set dummy values and let Read try to find it
	// Actually, this won't work. We need a different approach.

	// Try to parse the import ID
	var parentType, parentID, domainID string

	parts := strings.Split(importID, ":")
	if len(parts) == 3 {
		// Format: application:app-id:domain-id or compose:compose-id:domain-id
		parentType = parts[0]
		parentID = parts[1]
		domainID = parts[2]
	} else if len(parts) == 1 {
		// Just domain ID - we need to search for it
		// This is not ideal but we'll return an error asking for proper format
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Please use format 'application:<app-id>:<domain-id>' or 'compose:<compose-id>:<domain-id>'. Got: %s", importID),
		)
		return
	} else {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected format 'application:<app-id>:<domain-id>' or 'compose:<compose-id>:<domain-id>'. Got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), domainID)...)

	switch parentType {
	case "application":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parentID)...)
	case "compose":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("compose_id"), parentID)...)
	default:
		resp.Diagnostics.AddError(
			"Invalid parent type",
			fmt.Sprintf("Parent type must be 'application' or 'compose'. Got: %s", parentType),
		)
		return
	}
}
