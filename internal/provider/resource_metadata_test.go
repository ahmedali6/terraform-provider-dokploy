package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestApplicationResource_Metadata(t *testing.T) {
	r := NewApplicationResource()

	if r == nil {
		t.Fatal("Expected application resource, got nil")
	}

	req := resource.MetadataRequest{
		ProviderTypeName: "dokploy",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName == "" {
		t.Error("Expected type name to be set")
	}
}

func TestProjectResource_Metadata(t *testing.T) {
	r := NewProjectResource()

	if r == nil {
		t.Fatal("Expected project resource, got nil")
	}

	req := resource.MetadataRequest{
		ProviderTypeName: "dokploy",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName == "" {
		t.Error("Expected type name to be set")
	}
}

func TestEnvironmentResource_Metadata(t *testing.T) {
	r := NewEnvironmentResource()

	if r == nil {
		t.Fatal("Expected environment resource, got nil")
	}

	req := resource.MetadataRequest{
		ProviderTypeName: "dokploy",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName == "" {
		t.Error("Expected type name to be set")
	}
}

func TestApplicationResource_Schema(t *testing.T) {
	r := NewApplicationResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no schema errors, got: %v", resp.Diagnostics)
	}

	if len(resp.Schema.Attributes) == 0 {
		t.Error("Expected schema attributes to be defined")
	}
}

func TestProjectResource_Schema(t *testing.T) {
	r := NewProjectResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no schema errors, got: %v", resp.Diagnostics)
	}

	if len(resp.Schema.Attributes) == 0 {
		t.Error("Expected schema attributes to be defined")
	}
}

func TestEnvironmentResource_Schema(t *testing.T) {
	r := NewEnvironmentResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no schema errors, got: %v", resp.Diagnostics)
	}

	if len(resp.Schema.Attributes) == 0 {
		t.Error("Expected schema attributes to be defined")
	}
}

func TestResourceSchemaHelper(t *testing.T) {
	schemaDef := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	if schemaDef.Attributes == nil {
		t.Error("Expected attributes to be defined")
	}

	if _, ok := schemaDef.Attributes["test"]; !ok {
		t.Error("Expected 'test' attribute to exist")
	}
}
