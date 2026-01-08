package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &ServersDataSource{}

func NewServersDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

type ServersDataSource struct {
	client *client.DokployClient
}

type ServersDataSourceModel struct {
	ServerType types.String  `tfsdk:"server_type"`
	Servers    []ServerModel `tfsdk:"servers"`
}

type ServerModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	IPAddress    types.String `tfsdk:"ip_address"`
	Port         types.Int64  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	SSHKeyID     types.String `tfsdk:"ssh_key_id"`
	ServerStatus types.String `tfsdk:"server_status"`
	ServerType   types.String `tfsdk:"server_type"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (d *ServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of servers configured in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"server_type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter servers by type. Valid values: 'deploy', 'build'. If not specified, returns all servers.",
			},
			"servers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of servers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the server.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the server.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "The description of the server.",
						},
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the server.",
						},
						"port": schema.Int64Attribute{
							Computed:    true,
							Description: "The SSH port of the server.",
						},
						"username": schema.StringAttribute{
							Computed:    true,
							Description: "The SSH username for the server.",
						},
						"ssh_key_id": schema.StringAttribute{
							Computed:    true,
							Description: "The SSH key ID used for the server.",
						},
						"server_status": schema.StringAttribute{
							Computed:    true,
							Description: "The current status of the server.",
						},
						"server_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of server: 'deploy' or 'build'.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The creation timestamp of the server.",
						},
					},
				},
			},
		},
	}
}

func (d *ServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ServersDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	servers, err := d.client.ListServers()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Servers", err.Error())
		return
	}

	var state ServersDataSourceModel
	state.ServerType = config.ServerType

	filterType := ""
	if !config.ServerType.IsNull() && !config.ServerType.IsUnknown() {
		filterType = config.ServerType.ValueString()
	}

	for _, server := range servers {
		// Filter by server_type if specified
		if filterType != "" && server.ServerType != filterType {
			continue
		}

		serverModel := ServerModel{
			ID:           types.StringValue(server.ID),
			Name:         types.StringValue(server.Name),
			Description:  types.StringValue(server.Description),
			IPAddress:    types.StringValue(server.IPAddress),
			Port:         types.Int64Value(int64(server.Port)),
			Username:     types.StringValue(server.Username),
			SSHKeyID:     types.StringValue(server.SSHKeyID),
			ServerStatus: types.StringValue(server.ServerStatus),
			ServerType:   types.StringValue(server.ServerType),
			CreatedAt:    types.StringValue(server.CreatedAt),
		}
		state.Servers = append(state.Servers, serverModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
