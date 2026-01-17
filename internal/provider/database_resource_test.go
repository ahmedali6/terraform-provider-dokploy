package provider

import (
	"testing"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
)

func TestMapPostgresToState_InternalConnection(t *testing.T) {
	r := &PostgresResource{}
	state := &PostgresResourceModel{}

	postgres := &client.Postgres{
		PostgresID:        "test-id",
		Name:              "test",
		AppName:           "testapp",
		EnvironmentID:     "env-id",
		ApplicationStatus: "running",
		DatabaseName:      "testdb",
		DatabaseUser:      "user",
		DatabasePassword:  "password",
	}

	r.mapPostgresToState(state, postgres)

	if state.InternalConnection.ValueString() != "postgres://user:password@testapp:5432/testdb" {
		t.Errorf("Expected internal_connection to be %q, got %q", "postgres://user:password@testapp:5432/testdb", state.InternalConnection.ValueString())
	}

	if state.InternalConnection.IsNull() || state.InternalConnection.IsUnknown() {
		t.Error("internal_connection should never be null or unknown")
	}

	if state.InternalPort.ValueInt64() != 5432 {
		t.Errorf("Expected internal_port to be 5432, got %d", state.InternalPort.ValueInt64())
	}
}

func TestMapPostgresToState_ExternalConnection(t *testing.T) {
	r := &PostgresResource{}
	state := &PostgresResourceModel{}

	tests := []struct {
		name         string
		serverIP     string
		externalPort int
		expectValue  string
	}{
		{
			name:         "With ServerIP and ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 5432,
			expectValue:  "postgres://user:password@192.168.1.1:5432/testdb",
		},
		{
			name:         "Without ServerIP",
			serverIP:     "",
			externalPort: 5432,
			expectValue:  "",
		},
		{
			name:         "Without ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 0,
			expectValue:  "",
		},
		{
			name:         "Without ServerIP and ExternalPort",
			serverIP:     "",
			externalPort: 0,
			expectValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := &client.Postgres{
				PostgresID:        "test-id",
				Name:              "test",
				AppName:           "testapp",
				EnvironmentID:     "env-id",
				ApplicationStatus: "running",
				DatabaseName:      "testdb",
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				ServerIP:          tt.serverIP,
				ExternalPort:      tt.externalPort,
			}

			r.mapPostgresToState(state, postgres)

			if state.ExternalConnection.ValueString() != tt.expectValue {
				t.Errorf("Expected external_connection to be %q, got %q", tt.expectValue, state.ExternalConnection.ValueString())
			}

			if state.ExternalConnection.IsNull() || state.ExternalConnection.IsUnknown() {
				t.Error("external_connection should never be null or unknown")
			}
		})
	}
}

func TestMapMySQLToState_ExternalConnection(t *testing.T) {
	r := &MySQLResource{}
	state := &MySQLResourceModel{}

	tests := []struct {
		name         string
		serverIP     string
		externalPort int
		expectValue  string
	}{
		{
			name:         "With ServerIP and ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 3306,
			expectValue:  "mysql://user:password@192.168.1.1:3306/testdb",
		},
		{
			name:         "Without ServerIP",
			serverIP:     "",
			externalPort: 3306,
			expectValue:  "",
		},
		{
			name:         "Without ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 0,
			expectValue:  "",
		},
		{
			name:         "Without ServerIP and ExternalPort",
			serverIP:     "",
			externalPort: 0,
			expectValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysql := &client.MySQL{
				MySQLID:           "test-id",
				Name:              "test",
				AppName:           "testapp",
				EnvironmentID:     "env-id",
				ApplicationStatus: "running",
				DatabaseName:      "testdb",
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				ServerIP:          tt.serverIP,
				ExternalPort:      tt.externalPort,
			}

			r.mapMySQLToState(state, mysql)

			if state.ExternalConnection.ValueString() != tt.expectValue {
				t.Errorf("Expected external_connection to be %q, got %q", tt.expectValue, state.ExternalConnection.ValueString())
			}

			if state.ExternalConnection.IsNull() || state.ExternalConnection.IsUnknown() {
				t.Error("external_connection should never be null or unknown")
			}
		})
	}
}

func TestMapMariaDBToState_ExternalConnection(t *testing.T) {
	r := &MariaDBResource{}
	state := &MariaDBResourceModel{}

	tests := []struct {
		name         string
		serverIP     string
		externalPort int
		expectValue  string
	}{
		{
			name:         "With ServerIP and ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 3306,
			expectValue:  "mariadb://user:password@192.168.1.1:3306/testdb",
		},
		{
			name:         "Without ServerIP",
			serverIP:     "",
			externalPort: 3306,
			expectValue:  "",
		},
		{
			name:         "Without ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 0,
			expectValue:  "",
		},
		{
			name:         "Without ServerIP and ExternalPort",
			serverIP:     "",
			externalPort: 0,
			expectValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mariadb := &client.MariaDB{
				MariaDBID:         "test-id",
				Name:              "test",
				AppName:           "testapp",
				EnvironmentID:     "env-id",
				ApplicationStatus: "running",
				DatabaseName:      "testdb",
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				ServerIP:          tt.serverIP,
				ExternalPort:      tt.externalPort,
			}

			r.mapMariaDBToState(state, mariadb)

			if state.ExternalConnection.ValueString() != tt.expectValue {
				t.Errorf("Expected external_connection to be %q, got %q", tt.expectValue, state.ExternalConnection.ValueString())
			}

			if state.ExternalConnection.IsNull() || state.ExternalConnection.IsUnknown() {
				t.Error("external_connection should never be null or unknown")
			}
		})
	}
}

func TestMapMongoDBToState_ExternalConnection(t *testing.T) {
	r := &MongoDBResource{}
	state := &MongoDBResourceModel{}

	tests := []struct {
		name         string
		serverIP     string
		externalPort int
		expectValue  string
	}{
		{
			name:         "With ServerIP and ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 27017,
			expectValue:  "mongodb://user:password@192.168.1.1:27017",
		},
		{
			name:         "Without ServerIP",
			serverIP:     "",
			externalPort: 27017,
			expectValue:  "",
		},
		{
			name:         "Without ExternalPort",
			serverIP:     "192.168.1.1",
			externalPort: 0,
			expectValue:  "",
		},
		{
			name:         "Without ServerIP and ExternalPort",
			serverIP:     "",
			externalPort: 0,
			expectValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo := &client.MongoDB{
				MongoID:           "test-id",
				Name:              "test",
				AppName:           "testapp",
				EnvironmentID:     "env-id",
				ApplicationStatus: "running",
				DatabaseUser:      "user",
				DatabasePassword:  "password",
				ServerIP:          tt.serverIP,
				ExternalPort:      tt.externalPort,
			}

			r.mapMongoDBToState(state, mongo)

			if state.ExternalConnection.ValueString() != tt.expectValue {
				t.Errorf("Expected external_connection to be %q, got %q", tt.expectValue, state.ExternalConnection.ValueString())
			}

			if state.ExternalConnection.IsNull() || state.ExternalConnection.IsUnknown() {
				t.Error("external_connection should never be null or unknown")
			}
		})
	}
}

func TestPostgresResourceModelMap(t *testing.T) {
	r := &PostgresResource{}
	state := &PostgresResourceModel{}

	postgres := &client.Postgres{
		PostgresID:        "test-id",
		Name:              "test",
		AppName:           "testapp",
		EnvironmentID:     "env-id",
		ApplicationStatus: "running",
		DatabaseName:      "testdb",
		DatabaseUser:      "user",
		DatabasePassword:  "password",
	}

	r.mapPostgresToState(state, postgres)

	if state.ExternalConnection.IsNull() || state.ExternalConnection.IsUnknown() {
		t.Error("external_connection should always be known and not null")
	}
}
