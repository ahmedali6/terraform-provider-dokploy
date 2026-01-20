package provider

import (
	"testing"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReadApplicationIntoState(t *testing.T) {
	state := &ApplicationResourceModel{}
	app := &client.Application{
		ID:            "test-id",
		Name:          "test-app",
		AppName:       "testapp",
		SourceType:    "docker",
		Description:   "Test application",
		EnvironmentID: "env-1",
		ServerID:      "server-1",
	}

	readApplicationIntoState(state, app)

	if state.Name.ValueString() != "test-app" {
		t.Errorf("Expected name 'test-app', got %s", state.Name.ValueString())
	}

	if state.SourceType.ValueString() != "docker" {
		t.Errorf("Expected source_type 'docker', got %s", state.SourceType.ValueString())
	}
}

func TestUpdatePlanFromApplication(t *testing.T) {
	plan := &ApplicationResourceModel{
		Name: types.StringValue("test-app"),
	}
	app := &client.Application{
		ID:         "test-id",
		Name:       "test-app",
		AppName:    "testapp",
		SourceType: "docker",
		AutoDeploy: true,
	}

	updatePlanFromApplication(plan, app)

	if plan.AppName.ValueString() != "testapp" {
		t.Errorf("Expected app_name 'testapp', got %s", plan.AppName.ValueString())
	}

	if plan.SourceType.ValueString() != "docker" {
		t.Errorf("Expected source_type 'docker', got %s", plan.SourceType.ValueString())
	}

	if plan.AutoDeploy.ValueBool() != true {
		t.Errorf("Expected auto_deploy true, got %v", plan.AutoDeploy.ValueBool())
	}
}

func TestApplicationResourceModelDefaults(t *testing.T) {
	model := &ApplicationResourceModel{
		ID:   types.StringValue("test-id"),
		Name: types.StringValue("test-app"),
	}

	if model.ID.IsNull() {
		t.Error("ID should not be null when set")
	}

	if model.Name.IsNull() {
		t.Error("Name should not be null when set")
	}

	if !model.AutoDeploy.IsNull() {
		t.Error("AutoDeploy should be null when not set")
	}
}

func TestComposeResourceModelDefaults(t *testing.T) {
	model := &ComposeResourceModel{
		ID:   types.StringValue("test-id"),
		Name: types.StringValue("test-compose"),
	}

	if model.ID.IsNull() {
		t.Error("ID should not be null when set")
	}

	if model.Name.IsNull() {
		t.Error("Name should not be null when set")
	}
}
