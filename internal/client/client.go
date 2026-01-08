package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DokployClient holds connection details.
type DokployClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewDokployClient(baseURL, apiKey string) *DokployClient {
	return &DokployClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *DokployClient) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Fprintf(os.Stderr, "DEBUG RESPONSE [%s]: %s\n", endpoint, string(respBytes))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBytes))
	}

	return respBytes, nil
}

// --- User ---

type User struct {
	ID             string `json:"userId"`
	Email          string `json:"email"`
	OrganizationID string `json:"organizationId"`
}

func (c *DokployClient) GetUser() (*User, error) {
	resp, err := c.doRequest("GET", "user.get", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		User User `json:"user"` // Assuming wrapper based on other endpoints
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.User.ID != "" {
		// If org ID is missing on user, maybe we check roles/orgs?
		// For now assuming simple case.
		return &wrapper.User, nil
	}

	// Try direct
	var user User
	if err := json.Unmarshal(resp, &user); err == nil && user.ID != "" {
		return &user, nil
	}

	return nil, fmt.Errorf("failed to parse user response")
}

// --- Project ---

type Project struct {
	ID           string        `json:"projectId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Environments []Environment `json:"environments"`
}

type projectResponse struct {
	Project Project `json:"project"`
}

func (c *DokployClient) CreateProject(name, description string) (*Project, error) {
	payload := map[string]string{
		"name":        name,
		"description": description,
	}
	resp, err := c.doRequest("POST", "project.create", payload)
	if err != nil {
		return nil, err
	}

	var result projectResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result.Project, nil
}

func (c *DokployClient) GetProject(id string) (*Project, error) {
	endpoint := fmt.Sprintf("project.one?projectId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Project
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteProject(id string) error {
	payload := map[string]string{
		"projectId": id,
	}
	_, err := c.doRequest("POST", "project.remove", payload)
	return err
}

func (c *DokployClient) UpdateProject(id, name, description string) (*Project, error) {
	payload := map[string]string{
		"projectId":   id,
		"name":        name,
		"description": description,
	}
	resp, err := c.doRequest("POST", "project.update", payload)
	if err != nil {
		return nil, err
	}

	var result Project
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Environment ---

type Environment struct {
	ID          string     `json:"environmentId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ProjectID   string     `json:"projectId"`
	Postgres    []Database `json:"postgres"`
	Mysql       []Database `json:"mysql"`
	Mariadb     []Database `json:"mariadb"`
	Mongo       []Database `json:"mongo"`
	Redis       []Database `json:"redis"`
}

func (c *DokployClient) CreateEnvironment(projectID, name, description string) (*Environment, error) {
	payload := map[string]string{
		"projectId":   projectID,
		"name":        name,
		"description": description,
	}
	resp, err := c.doRequest("POST", "environment.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Environment Environment `json:"environment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Environment.ID != "" {
		return &wrapper.Environment, nil
	}

	var result Environment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateEnvironment(env Environment) (*Environment, error) {
	payload := map[string]interface{}{
		"environmentId": env.ID,
		"name":          env.Name,
		"description":   env.Description,
		"projectId":     env.ProjectID,
	}
	resp, err := c.doRequest("POST", "environment.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Environment Environment `json:"environment"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Environment.ID != "" {
		return &wrapper.Environment, nil
	}

	var result Environment
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteEnvironment(id string) error {
	payload := map[string]string{
		"environmentId": id,
	}
	_, err := c.doRequest("POST", "environment.remove", payload)
	return err
}

// --- Application ---

type Application struct {
	// Core identifiers
	ID            string `json:"applicationId"`
	Name          string `json:"name"`
	AppName       string `json:"appName"`
	Description   string `json:"description"`
	ProjectID     string `json:"projectId"`
	EnvironmentID string `json:"environmentId"`
	ServerID      string `json:"serverId"`

	// Source configuration
	SourceType string `json:"sourceType"` // github, gitlab, bitbucket, git, docker, drop

	// Git provider settings (application.saveGitProvider)
	CustomGitUrl       string `json:"customGitUrl"`
	CustomGitBranch    string `json:"customGitBranch"`
	CustomGitSSHKeyId  string `json:"customGitSSHKeyId"`
	CustomGitBuildPath string `json:"customGitBuildPath"`
	EnableSubmodules   bool   `json:"enableSubmodules"`
	WatchPaths         string `json:"watchPaths"` // Stored as JSON array string

	// GitHub provider settings (application.saveGithubProvider)
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Owner       string `json:"owner"`
	BuildPath   string `json:"buildPath"`
	GithubId    string `json:"githubId"`
	TriggerType string `json:"triggerType"` // push, tag

	// Docker provider settings (application.saveDockerProvider)
	DockerImage string `json:"dockerImage"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RegistryUrl string `json:"registryUrl"`

	// Build type settings (application.saveBuildType)
	BuildType         string `json:"buildType"` // dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, railpack
	DockerfilePath    string `json:"dockerfile"`
	DockerContextPath string `json:"dockerContextPath"`
	DockerBuildStage  string `json:"dockerBuildStage"`
	PublishDirectory  string `json:"publishDirectory"`
	Dockerfile        string `json:"dockerfileContent"` // Raw Dockerfile content for drop source

	// Environment settings (application.saveEnvironment)
	Env           string `json:"env"`
	BuildArgs     string `json:"buildArgs"`
	BuildSecrets  string `json:"buildSecrets"`
	CreateEnvFile bool   `json:"createEnvFile"`

	// Runtime configuration (application.update)
	AutoDeploy        bool                   `json:"autoDeploy"`
	Replicas          int                    `json:"replicas"`
	MemoryLimit       *int64                 `json:"memoryLimit"`
	MemoryReservation *int64                 `json:"memoryReservation"`
	CpuLimit          *int64                 `json:"cpuLimit"`
	CpuReservation    *int64                 `json:"cpuReservation"`
	Command           string                 `json:"command"`
	EntryPoint        string                 `json:"entrypoint"`
	HealthCheckSwarm  map[string]interface{} `json:"healthCheckSwarm"`

	// Application status
	ApplicationStatus string `json:"applicationStatus"` // idle, running, done, error

	// Domains
	Domains []Domain `json:"domains"`

	// Timestamps
	CreatedAt string `json:"createdAt"`
}

func (c *DokployClient) CreateApplication(app Application) (*Application, error) {
	// 1. Create application with minimal required fields
	createPayload := map[string]interface{}{
		"name":          app.Name,
		"environmentId": app.EnvironmentID,
	}

	// Include optional create-time fields
	if app.AppName != "" {
		createPayload["appName"] = app.AppName
	}
	if app.Description != "" {
		createPayload["description"] = app.Description
	}
	if app.ServerID != "" {
		createPayload["serverId"] = app.ServerID
	}

	resp, err := c.doRequest("POST", "application.create", createPayload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Application Application `json:"application"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, err
	}

	createdApp := wrapper.Application
	if createdApp.ID == "" {
		if err := json.Unmarshal(resp, &createdApp); err != nil {
			return nil, err
		}
	}

	// Preserve serverId since API may not return it
	if app.ServerID != "" {
		createdApp.ServerID = app.ServerID
	}

	return &createdApp, nil
}

func (c *DokployClient) GetApplication(id string) (*Application, error) {
	endpoint := fmt.Sprintf("application.one?applicationId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateApplicationGeneral updates the general application settings.
// Corresponds to application.update endpoint for general fields.
func (c *DokployClient) UpdateApplicationGeneral(app Application) (*Application, error) {
	payload := map[string]interface{}{
		"applicationId": app.ID,
	}

	// Only include fields that should be updated via application.update
	if app.Name != "" {
		payload["name"] = app.Name
	}
	if app.AppName != "" {
		payload["appName"] = app.AppName
	}
	if app.Description != "" {
		payload["description"] = app.Description
	}
	if app.SourceType != "" {
		payload["sourceType"] = app.SourceType
	}

	// Boolean fields - always include
	payload["autoDeploy"] = app.AutoDeploy

	// Numeric fields
	if app.Replicas > 0 {
		payload["replicas"] = app.Replicas
	}
	if app.MemoryLimit != nil {
		payload["memoryLimit"] = *app.MemoryLimit
	}
	if app.MemoryReservation != nil {
		payload["memoryReservation"] = *app.MemoryReservation
	}
	if app.CpuLimit != nil {
		payload["cpuLimit"] = *app.CpuLimit
	}
	if app.CpuReservation != nil {
		payload["cpuReservation"] = *app.CpuReservation
	}

	// String fields
	if app.Command != "" {
		payload["command"] = app.Command
	}
	if app.EntryPoint != "" {
		payload["entrypoint"] = app.EntryPoint
	}

	resp, err := c.doRequest("POST", "application.update", payload)
	if err != nil {
		return nil, err
	}

	// API might return true or the updated application
	if string(resp) == "true" {
		return c.GetApplication(app.ID)
	}

	var result Application
	if err := json.Unmarshal(resp, &result); err != nil {
		// If unmarshal fails, fetch the application
		return c.GetApplication(app.ID)
	}
	return &result, nil
}

// UpdateApplication is kept for backward compatibility.
// It calls UpdateApplicationGeneral.
func (c *DokployClient) UpdateApplication(app Application) (*Application, error) {
	return c.UpdateApplicationGeneral(app)
}

func (c *DokployClient) DeleteApplication(id string) error {
	payload := map[string]string{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.remove", payload)
	return err
}

func (c *DokployClient) DeployApplication(id string, serverId string) error {
	payload := map[string]interface{}{
		"applicationId": id,
	}
	if serverId != "" {
		payload["serverId"] = serverId
	}
	_, err := c.doRequest("POST", "application.deploy", payload)
	return err
}

func (c *DokployClient) RedeployApplication(id string) error {
	payload := map[string]interface{}{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.redeploy", payload)
	return err
}

func (c *DokployClient) StopApplication(id string) error {
	payload := map[string]interface{}{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.stop", payload)
	return err
}

func (c *DokployClient) StartApplication(id string) error {
	payload := map[string]interface{}{
		"applicationId": id,
	}
	_, err := c.doRequest("POST", "application.start", payload)
	return err
}

// SaveBuildType configures the build type settings for an application.
// Corresponds to application.saveBuildType endpoint.
func (c *DokployClient) SaveBuildType(appID string, buildType string, dockerfile string, dockerContextPath string, dockerBuildStage string, publishDirectory string) error {
	// The API requires all these fields to be present as strings (even if empty)
	payload := map[string]interface{}{
		"applicationId":     appID,
		"buildType":         buildType,
		"dockerfile":        dockerfile,
		"dockerContextPath": dockerContextPath,
		"dockerBuildStage":  dockerBuildStage,
		"publishDirectory":  publishDirectory,
	}

	_, err := c.doRequest("POST", "application.saveBuildType", payload)
	return err
}

// SaveGitProviderInput contains all the fields for the saveGitProvider endpoint.
type SaveGitProviderInput struct {
	ApplicationID      string
	CustomGitBranch    string
	CustomGitBuildPath string
	CustomGitUrl       string
	CustomGitSSHKeyId  string
	EnableSubmodules   bool
	WatchPaths         []string
}

// SaveGitProvider configures the git provider settings for an application.
// Corresponds to application.saveGitProvider endpoint.
func (c *DokployClient) SaveGitProvider(input SaveGitProviderInput) error {
	payload := map[string]interface{}{
		"applicationId": input.ApplicationID,
	}

	if input.CustomGitBranch != "" {
		payload["customGitBranch"] = input.CustomGitBranch
	}
	if input.CustomGitBuildPath != "" {
		payload["customGitBuildPath"] = input.CustomGitBuildPath
	}
	if input.CustomGitUrl != "" {
		payload["customGitUrl"] = input.CustomGitUrl
	}
	if input.CustomGitSSHKeyId != "" {
		payload["customGitSSHKeyId"] = input.CustomGitSSHKeyId
	}
	if input.EnableSubmodules {
		payload["enableSubmodules"] = input.EnableSubmodules
	}
	if len(input.WatchPaths) > 0 {
		payload["watchPaths"] = input.WatchPaths
	}

	_, err := c.doRequest("POST", "application.saveGitProvider", payload)
	return err
}

// SaveGithubProviderInput contains all the fields for the saveGithubProvider endpoint.
type SaveGithubProviderInput struct {
	ApplicationID    string
	Repository       string
	Branch           string
	Owner            string
	BuildPath        string
	GithubId         string
	WatchPaths       []string
	EnableSubmodules bool
	TriggerType      string // push, tag
}

// SaveGithubProvider configures the GitHub provider settings for an application.
// Corresponds to application.saveGithubProvider endpoint.
func (c *DokployClient) SaveGithubProvider(input SaveGithubProviderInput) error {
	payload := map[string]interface{}{
		"applicationId":    input.ApplicationID,
		"enableSubmodules": input.EnableSubmodules,
	}

	// Required fields that can be null
	if input.Owner != "" {
		payload["owner"] = input.Owner
	} else {
		payload["owner"] = nil
	}

	if input.GithubId != "" {
		payload["githubId"] = input.GithubId
	} else {
		payload["githubId"] = nil
	}

	// Optional fields
	if input.Repository != "" {
		payload["repository"] = input.Repository
	}
	if input.Branch != "" {
		payload["branch"] = input.Branch
	}
	if input.BuildPath != "" {
		payload["buildPath"] = input.BuildPath
	}
	if len(input.WatchPaths) > 0 {
		payload["watchPaths"] = input.WatchPaths
	}
	if input.TriggerType != "" {
		payload["triggerType"] = input.TriggerType
	}

	_, err := c.doRequest("POST", "application.saveGithubProvider", payload)
	return err
}

// SaveDockerProviderInput contains all the fields for the saveDockerProvider endpoint.
type SaveDockerProviderInput struct {
	ApplicationID string
	DockerImage   string
	Username      string
	Password      string
	RegistryUrl   string
}

// SaveDockerProvider configures the docker provider settings for an application.
// Corresponds to application.saveDockerProvider endpoint.
func (c *DokployClient) SaveDockerProvider(input SaveDockerProviderInput) error {
	payload := map[string]interface{}{
		"applicationId": input.ApplicationID,
	}

	if input.DockerImage != "" {
		payload["dockerImage"] = input.DockerImage
	}
	if input.Username != "" {
		payload["username"] = input.Username
	}
	if input.Password != "" {
		payload["password"] = input.Password
	}
	if input.RegistryUrl != "" {
		payload["registryUrl"] = input.RegistryUrl
	}

	_, err := c.doRequest("POST", "application.saveDockerProvider", payload)
	return err
}

// SaveEnvironmentInput contains all the fields for the saveEnvironment endpoint.
type SaveEnvironmentInput struct {
	ApplicationID string
	Env           string
	BuildArgs     string
	BuildSecrets  string
	CreateEnvFile *bool
}

// SaveEnvironment configures the environment variables for an application.
// Corresponds to application.saveEnvironment endpoint.
func (c *DokployClient) SaveEnvironment(input SaveEnvironmentInput) error {
	payload := map[string]interface{}{
		"applicationId": input.ApplicationID,
	}

	// env can be empty string, so we always include it
	payload["env"] = input.Env

	if input.BuildArgs != "" {
		payload["buildArgs"] = input.BuildArgs
	}
	if input.BuildSecrets != "" {
		payload["buildSecrets"] = input.BuildSecrets
	}
	if input.CreateEnvFile != nil {
		payload["createEnvFile"] = *input.CreateEnvFile
	}

	_, err := c.doRequest("POST", "application.saveEnvironment", payload)
	return err
}

// --- Compose ---

type Compose struct {
	ID                string   `json:"composeId"`
	Name              string   `json:"name"`
	ProjectID         string   `json:"projectId"`
	EnvironmentID     string   `json:"environmentId"`
	ComposeFile       string   `json:"composeFile"`
	SourceType        string   `json:"sourceType"`
	CustomGitUrl      string   `json:"customGitUrl"`
	CustomGitBranch   string   `json:"customGitBranch"`
	CustomGitSSHKeyId string   `json:"customGitSSHKeyId"`
	ComposePath       string   `json:"composePath"`
	AutoDeploy        bool     `json:"autoDeploy"`
	Domains           []Domain `json:"domains"`
	ServerID          string   `json:"serverId"`
}

func (c *DokployClient) CreateCompose(comp Compose) (*Compose, error) {
	// 1. Create compose with serverId
	payload := map[string]interface{}{
		"environmentId": comp.EnvironmentID,
		"name":          comp.Name,
		"composeType":   "docker-compose",
		"appName":       comp.Name,
	}

	// Include serverId if provided
	if comp.ServerID != "" {
		payload["serverId"] = comp.ServerID
	}

	// If raw content provided, include it
	if comp.ComposeFile != "" {
		payload["composeFile"] = comp.ComposeFile
	}

	resp, err := c.doRequest("POST", "compose.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Compose Compose `json:"compose"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Compose.ID != "" {
		// If serverId was passed, set it on the returned object
		if comp.ServerID != "" {
			wrapper.Compose.ServerID = comp.ServerID
		}
		return &wrapper.Compose, nil
	}

	createdComp := wrapper.Compose
	if createdComp.ID == "" {
		if err := json.Unmarshal(resp, &createdComp); err != nil {
			return nil, err
		}
	}

	// Preserve serverId
	if comp.ServerID != "" {
		createdComp.ServerID = comp.ServerID
	}

	// 2. Update with Git configuration if necessary
	updatePayload := map[string]interface{}{
		"composeId":  createdComp.ID,
		"name":       comp.Name,
		"sourceType": comp.SourceType,
		"autoDeploy": comp.AutoDeploy,
	}

	if comp.CustomGitUrl != "" {
		updatePayload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		updatePayload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		updatePayload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.ComposePath != "" {
		updatePayload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		updatePayload["composeFile"] = comp.ComposeFile
	}

	if comp.SourceType == "" {
		if comp.CustomGitUrl != "" {
			updatePayload["sourceType"] = "git"
		} else if comp.ComposeFile != "" {
			updatePayload["sourceType"] = "raw"
		} else {
			updatePayload["sourceType"] = "github"
		}
	}

	respUpdate, err := c.doRequest("POST", "compose.update", updatePayload)
	if err != nil {
		return nil, fmt.Errorf("created compose %s but failed to update config: %w", createdComp.ID, err)
	}

	if string(respUpdate) == "true" {
		result, err := c.GetCompose(createdComp.ID)
		if err != nil {
			return nil, err
		}
		// Preserve serverId
		if comp.ServerID != "" {
			result.ServerID = comp.ServerID
		}
		return result, nil
	}

	var updateResult Compose
	if err := json.Unmarshal(respUpdate, &wrapper); err == nil && wrapper.Compose.ID != "" {
		if comp.ServerID != "" {
			wrapper.Compose.ServerID = comp.ServerID
		}
		return &wrapper.Compose, nil
	}
	if err := json.Unmarshal(respUpdate, &updateResult); err == nil {
		if comp.ServerID != "" {
			updateResult.ServerID = comp.ServerID
		}
		return &updateResult, nil
	}

	return &createdComp, nil
}

func (c *DokployClient) GetCompose(id string) (*Compose, error) {
	endpoint := fmt.Sprintf("compose.one?composeId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result Compose
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateCompose(comp Compose) (*Compose, error) {
	payload := map[string]interface{}{
		"composeId":  comp.ID,
		"name":       comp.Name,
		"sourceType": comp.SourceType,
		"autoDeploy": comp.AutoDeploy,
	}

	if comp.CustomGitUrl != "" {
		payload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		payload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		payload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.ComposePath != "" {
		payload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		payload["composeFile"] = comp.ComposeFile
	}

	if comp.EnvironmentID != "" {
		payload["environmentId"] = comp.EnvironmentID
	}

	resp, err := c.doRequest("POST", "compose.update", payload)
	if err != nil {
		return nil, err
	}

	var result Compose
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteCompose(id string) error {
	payload := map[string]string{
		"composeId": id,
	}
	_, err := c.doRequest("POST", "compose.remove", payload)
	return err
}

func (c *DokployClient) DeployCompose(id string, serverId string) error {
	payload := map[string]interface{}{
		"composeId": id,
	}
	if serverId != "" {
		payload["serverId"] = serverId
	}
	_, err := c.doRequest("POST", "compose.deploy", payload)
	return err
}

// --- Database ---

type Database struct {
	ID            string `json:"databaseId"`
	Name          string `json:"name"`
	AppName       string `json:"appName"`
	Type          string `json:"type"`
	ProjectID     string `json:"projectId"`
	EnvironmentID string `json:"environmentId"`
	Version       string `json:"version"`
	DockerImage   string `json:"dockerImage"`
	ExternalPort  int64  `json:"externalPort"`
	InternalPort  int64  `json:"internalPort"`
	Password      string `json:"password"`
	PostgresID    string `json:"postgresId"`
	MysqlID       string `json:"mysqlId"`
	MariadbID     string `json:"mariadbId"`
	MongoID       string `json:"mongoId"`
	RedisID       string `json:"redisId"`
}

func (c *DokployClient) CreateDatabase(projectID, environmentID, name, dbType, password, dockerImage string) (*Database, error) {
	var endpoint string
	payload := map[string]string{
		"environmentId":    environmentID,
		"name":             name,
		"appName":          name,
		"databaseName":     name,
		"databasePassword": password,
		"dockerImage":      dockerImage,
	}

	switch dbType {
	case "postgres":
		endpoint = "postgres.create"
		payload["databaseUser"] = "postgres"
	case "mysql":
		endpoint = "mysql.create"
		payload["databaseUser"] = "root"
		payload["databaseRootPassword"] = password // MySQL requires separate root password
	case "mariadb":
		endpoint = "mariadb.create"
		payload["databaseUser"] = "root"
		payload["databaseRootPassword"] = password // MariaDB requires separate root password
	case "mongo":
		endpoint = "mongo.create"
		payload["databaseUser"] = "mongo"
	case "redis":
		endpoint = "redis.create"
		payload["databaseUser"] = "default"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	resp, err := c.doRequest("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	if string(resp) == "true" {
		project, err := c.GetProject(projectID)
		if err != nil {
			return nil, fmt.Errorf("database created but failed to fetch project: %w", err)
		}

		for _, env := range project.Environments {
			if env.ID == environmentID {
				var dbs []Database
				switch dbType {
				case "postgres":
					dbs = env.Postgres
				case "mysql":
					dbs = env.Mysql
				case "mariadb":
					dbs = env.Mariadb
				case "mongo":
					dbs = env.Mongo
				case "redis":
					dbs = env.Redis
				}

				for _, db := range dbs {
					if db.Name == name || db.AppName == name {
						id := db.PostgresID
						if db.MysqlID != "" {
							id = db.MysqlID
						}
						if db.MariadbID != "" {
							id = db.MariadbID
						}
						if db.MongoID != "" {
							id = db.MongoID
						}
						if db.RedisID != "" {
							id = db.RedisID
						}
						if id != "" {
							db.ID = id
						}

						if db.Type == "" {
							db.Type = dbType
						}
						return &db, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("database created but not found in project environments")
	}

	var wrapper struct {
		Database Database `json:"database"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		db := wrapper.Database

		// Extract ID from type-specific fields if generic ID is not set
		if db.ID == "" {
			switch dbType {
			case "postgres":
				db.ID = db.PostgresID
			case "mysql":
				db.ID = db.MysqlID
			case "mariadb":
				db.ID = db.MariadbID
			case "mongo":
				db.ID = db.MongoID
			case "redis":
				db.ID = db.RedisID
			}
		}

		if db.ID != "" {
			if db.Type == "" {
				db.Type = dbType
			}
			return &db, nil
		}
	}

	var result Database
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if result.Type == "" {
		result.Type = dbType
	}

	// Extract ID from type-specific fields if generic ID is not set
	if result.ID == "" {
		switch dbType {
		case "postgres":
			result.ID = result.PostgresID
		case "mysql":
			result.ID = result.MysqlID
		case "mariadb":
			result.ID = result.MariadbID
		case "mongo":
			result.ID = result.MongoID
		case "redis":
			result.ID = result.RedisID
		}
	}

	return &result, nil
}

func (c *DokployClient) GetDatabase(dbID string, databaseType string) (*Database, error) {
	var endpoint string
	switch databaseType {
	case "postgres":
		endpoint = fmt.Sprintf("postgres.one?postgresId=%s", dbID)
	case "mysql":
		endpoint = fmt.Sprintf("mysql.one?mysqlId=%s", dbID)
	case "mariadb":
		endpoint = fmt.Sprintf("mariadb.one?mariadbId=%s", dbID)
	case "mongo":
		endpoint = fmt.Sprintf("mongo.one?mongoId=%s", dbID)
	case "redis":
		endpoint = fmt.Sprintf("redis.one?redisId=%s", dbID)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", databaseType)
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var db Database
	if err := json.Unmarshal(resp, &db); err == nil {
		valid := false
		if db.ID != "" {
			valid = true
		}
		if db.PostgresID != "" {
			valid = true
		}
		if db.MysqlID != "" {
			valid = true
		}
		if db.MariadbID != "" {
			valid = true
		}
		if db.MongoID != "" {
			valid = true
		}
		if db.RedisID != "" {
			valid = true
		}

		if valid {
			if db.ID == "" {
				if db.PostgresID != "" {
					db.ID = db.PostgresID
				}
				if db.MysqlID != "" {
					db.ID = db.MysqlID
				}
				if db.MariadbID != "" {
					db.ID = db.MariadbID
				}
				if db.MongoID != "" {
					db.ID = db.MongoID
				}
				if db.RedisID != "" {
					db.ID = db.RedisID
				}
			}
			db.Type = databaseType
			return &db, nil
		}
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	var dbBytes json.RawMessage
	var ok bool

	switch databaseType {
	case "postgres":
		dbBytes, ok = result["postgres"]
	case "mysql":
		dbBytes, ok = result["mysql"]
	case "mariadb":
		dbBytes, ok = result["mariadb"]
	case "mongo":
		dbBytes, ok = result["mongo"]
	case "redis":
		dbBytes, ok = result["redis"]
	}

	if !ok {
		if val, found := result["database"]; found {
			dbBytes = val
		} else {
			return nil, fmt.Errorf("database key not found in response for type %s", databaseType)
		}
	}

	if err := json.Unmarshal(dbBytes, &db); err != nil {
		return nil, err
	}

	if db.ID == "" {
		if db.PostgresID != "" {
			db.ID = db.PostgresID
		}
		if db.MysqlID != "" {
			db.ID = db.MysqlID
		}
		if db.MariadbID != "" {
			db.ID = db.MariadbID
		}
		if db.MongoID != "" {
			db.ID = db.MongoID
		}
		if db.RedisID != "" {
			db.ID = db.RedisID
		}
	}
	db.Type = databaseType

	return &db, nil
}

func (c *DokployClient) DeleteDatabase(id string) error {
	return fmt.Errorf("delete database requires type update")
}

func (c *DokployClient) DeleteDatabaseWithType(id, dbType string) error {
	var endpoint string
	var idKey string
	switch dbType {
	case "postgres":
		endpoint = "postgres.remove"
		idKey = "postgresId"
	case "mysql":
		endpoint = "mysql.remove"
		idKey = "mysqlId"
	case "mariadb":
		endpoint = "mariadb.remove"
		idKey = "mariadbId"
	case "mongo":
		endpoint = "mongo.remove"
		idKey = "mongoId"
	case "redis":
		endpoint = "redis.remove"
		idKey = "redisId"
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	payload := map[string]string{
		idKey: id,
	}
	_, err := c.doRequest("POST", endpoint, payload)
	return err
}

// --- Domain ---

type Domain struct {
	ID              string `json:"domainId"`
	ApplicationID   string `json:"applicationId"`
	ComposeID       string `json:"composeId"`
	ServiceName     string `json:"serviceName"`
	Host            string `json:"host"`
	Path            string `json:"path"`
	Port            int64  `json:"port"`
	HTTPS           bool   `json:"https"`
	CertificateType string `json:"certificateType"`
}

func (c *DokployClient) CreateDomain(domain Domain) (*Domain, error) {
	payload := map[string]interface{}{
		"host":  domain.Host,
		"path":  domain.Path,
		"port":  domain.Port,
		"https": domain.HTTPS,
	}
	// Set certificate type based on HTTPS setting
	if domain.HTTPS {
		if domain.CertificateType != "" {
			payload["certificateType"] = domain.CertificateType
		} else {
			payload["certificateType"] = "letsencrypt"
		}
	} else {
		payload["certificateType"] = "none"
	}
	if domain.ApplicationID != "" {
		payload["applicationId"] = domain.ApplicationID
	}
	if domain.ComposeID != "" {
		payload["composeId"] = domain.ComposeID
	}
	if domain.ServiceName != "" {
		payload["serviceName"] = domain.ServiceName
	}

	resp, err := c.doRequest("POST", "domain.create", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Domain Domain `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain.ID != "" {
		return &wrapper.Domain, nil
	}

	var result Domain
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetDomainsByApplication(appID string) ([]Domain, error) {
	app, err := c.GetApplication(appID)
	if err != nil {
		return nil, err
	}
	return app.Domains, nil
}

func (c *DokployClient) GetDomainsByCompose(composeID string) ([]Domain, error) {
	comp, err := c.GetCompose(composeID)
	if err != nil {
		return nil, err
	}
	return comp.Domains, nil
}

func (c *DokployClient) DeleteDomain(id string) error {
	payload := map[string]string{
		"domainId": id,
	}
	_, err := c.doRequest("POST", "domain.remove", payload)
	return err
}

func (c *DokployClient) GenerateDomain(appName string) (string, error) {
	payload := map[string]string{
		"appName": appName,
	}
	resp, err := c.doRequest("POST", "domain.generateDomain", payload)
	if err != nil {
		return "", err
	}

	// Try to parse as JSON wrapper
	var wrapper struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain != "" {
		return wrapper.Domain, nil
	}

	// Fallback: maybe it returns just the string in quotes or raw?
	// If it is a simple string "foo.bar", Unmarshal might fail or we just return string(resp) trimmed.
	return strings.Trim(string(resp), "\""), nil
}

func (c *DokployClient) UpdateDomain(domain Domain) (*Domain, error) {
	payload := map[string]interface{}{
		"domainId":    domain.ID,
		"host":        domain.Host,
		"path":        domain.Path,
		"port":        domain.Port,
		"https":       domain.HTTPS,
		"serviceName": domain.ServiceName,
	}
	// Set certificate type based on HTTPS setting
	if domain.HTTPS {
		if domain.CertificateType != "" {
			payload["certificateType"] = domain.CertificateType
		} else {
			payload["certificateType"] = "letsencrypt"
		}
	} else {
		payload["certificateType"] = "none"
	}
	resp, err := c.doRequest("POST", "domain.update", payload)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Domain Domain `json:"domain"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.Domain.ID != "" {
		return &wrapper.Domain, nil
	}

	var result Domain
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Environment Variable ---

type EnvironmentVariable struct {
	ID            string `json:"id"`
	ApplicationID string `json:"applicationId"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	Scope         string `json:"scope"`
}

func (c *DokployClient) UpdateApplicationEnv(appID string, updateFn func(envMap map[string]string), createEnvFile *bool) error {
	var lastErr error
	for i := 0; i < 5; i++ { // Retry up to 5 times
		app, err := c.GetApplication(appID)
		if err != nil {
			return err
		}

		envMap := ParseEnv(app.Env)
		originalEnvStr := app.Env

		updateFn(envMap) // Modify the map

		newEnvStr := formatEnv(envMap)

		if newEnvStr == originalEnvStr {
			return nil // No changes to be made
		}

		payload := map[string]interface{}{
			"applicationId": appID,
			"env":           newEnvStr,
		}
		if createEnvFile != nil {
			payload["createEnvFile"] = *createEnvFile
		}

		_, err = c.doRequest("POST", "application.saveEnvironment", payload)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond) // Backoff
			continue
		}

		// Verify write
		verifyApp, err := c.GetApplication(appID)
		if err != nil {
			// If we can't verify, we have to assume it worked or retry
			lastErr = fmt.Errorf("failed to verify environment update: %w", err)
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		if verifyApp.Env == newEnvStr {
			return nil // Success
		}
		lastErr = fmt.Errorf("environment update conflict, retrying")
	}
	return lastErr
}

func (c *DokployClient) CreateVariable(appID, key, value, scope string, createEnvFile *bool) (*EnvironmentVariable, error) {
	err := c.UpdateApplicationEnv(appID, func(envMap map[string]string) {
		envMap[key] = value
	}, createEnvFile)

	if err != nil {
		return nil, err
	}

	return &EnvironmentVariable{
		ID:            appID + "_" + key,
		ApplicationID: appID,
		Key:           key,
		Value:         value,
		Scope:         scope,
	}, nil
}

func (c *DokployClient) GetVariablesByApplication(appID string) ([]EnvironmentVariable, error) {
	app, err := c.GetApplication(appID)
	if err != nil {
		return nil, err
	}
	envMap := ParseEnv(app.Env)
	var vars []EnvironmentVariable
	for k, v := range envMap {
		vars = append(vars, EnvironmentVariable{
			ID:            appID + "_" + k,
			ApplicationID: appID,
			Key:           k,
			Value:         v,
			Scope:         "runtime",
		})
	}
	return vars, nil
}

func (c *DokployClient) DeleteVariable(id string, createEnvFile *bool) error {
	parts := strings.SplitN(id, "_", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid variable ID format")
	}
	appID, key := parts[0], parts[1]

	return c.UpdateApplicationEnv(appID, func(envMap map[string]string) {
		delete(envMap, key)
	}, createEnvFile)
}

func ParseEnv(env string) map[string]string {
	m := make(map[string]string)
	if env == "" {
		return m
	}
	lines := strings.Split(env, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func formatEnv(m map[string]string) string {
	var lines []string
	for k, v := range m {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(lines, "\n")
}

// --- SSH Key ---

type SSHKey struct {
	ID          string `json:"sshKeyId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
}

func (c *DokployClient) CreateSSHKey(name, description, privateKey, publicKey string) (*SSHKey, error) {
	// Fetch user to get Organization ID
	user, err := c.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user for organization ID: %w", err)
	}

	payload := map[string]string{
		"name":           name,
		"description":    description,
		"privateKey":     privateKey,
		"publicKey":      publicKey,
		"organizationId": user.OrganizationID,
	}

	resp, err := c.doRequest("POST", "sshKey.create", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response or boolean by fetching list
	if len(resp) == 0 || string(resp) == "true" {
		return c.findSSHKeyByName(name)
	}

	var wrapper struct {
		SSHKey SSHKey `json:"sshKey"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.SSHKey.ID != "" {
		return &wrapper.SSHKey, nil
	}

	var result SSHKey
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return c.findSSHKeyByName(name)
	}

	// Fallback to list lookup if unmarshal failed to produce ID
	return &result, nil
}

func (c *DokployClient) ListSSHKeys() ([]SSHKey, error) {
	resp, err := c.doRequest("GET", "sshKey.all", nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		SSHKeys []SSHKey `json:"sshKeys"` // Guessing wrapper
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.SSHKeys != nil {
		return wrapper.SSHKeys, nil
	}

	var list []SSHKey
	if err := json.Unmarshal(resp, &list); err == nil {
		return list, nil
	}

	return nil, fmt.Errorf("failed to parse sshKey.all response")
}

func (c *DokployClient) findSSHKeyByName(name string) (*SSHKey, error) {
	keys, err := c.ListSSHKeys()
	if err != nil {
		return nil, fmt.Errorf("ssh key created but failed to list keys: %w", err)
	}
	for _, key := range keys {
		if key.Name == name {
			return &key, nil
		}
	}
	return nil, fmt.Errorf("ssh key created but not found in list by name: %s", name)
}

func (c *DokployClient) GetSSHKey(id string) (*SSHKey, error) {
	endpoint := fmt.Sprintf("sshKey.one?sshKeyId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var result SSHKey
	if err := json.Unmarshal(resp, &result); err != nil {
		// Try wrapper?
		var wrapper struct {
			SSHKey SSHKey `json:"sshKey"`
		}
		if err2 := json.Unmarshal(resp, &wrapper); err2 == nil {
			return &wrapper.SSHKey, nil
		}
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteSSHKey(id string) error {
	payload := map[string]string{
		"sshKeyId": id,
	}
	_, err := c.doRequest("POST", "sshKey.remove", payload)
	return err
}

// --- Server ---

type Server struct {
	ID                  string `json:"serverId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	IPAddress           string `json:"ipAddress"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	SSHKeyID            string `json:"sshKeyId"`
	ServerStatus        string `json:"serverStatus"`
	ServerType          string `json:"serverType"`
	CreatedAt           string `json:"createdAt"`
	OrganizationID      string `json:"organizationId"`
	AppName             string `json:"appName"`
	EnableDockerCleanup bool   `json:"enableDockerCleanup"`
}

func (c *DokployClient) ListServers() ([]Server, error) {
	resp, err := c.doRequest("GET", "server.all", nil)
	if err != nil {
		return nil, err
	}

	var servers []Server
	if err := json.Unmarshal(resp, &servers); err != nil {
		// Try wrapper format
		var wrapper struct {
			Servers []Server `json:"servers"`
		}
		if err2 := json.Unmarshal(resp, &wrapper); err2 == nil {
			return wrapper.Servers, nil
		}
		return nil, err
	}
	return servers, nil
}

func (c *DokployClient) GetServer(id string) (*Server, error) {
	endpoint := fmt.Sprintf("server.one?serverId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var server Server
	if err := json.Unmarshal(resp, &server); err != nil {
		var wrapper struct {
			Server Server `json:"server"`
		}
		if err2 := json.Unmarshal(resp, &wrapper); err2 == nil {
			return &wrapper.Server, nil
		}
		return nil, err
	}
	return &server, nil
}

// --- GitHub Provider ---

type GithubProvider struct {
	ID              string `json:"githubId"`
	GitHubAppName   string `json:"gitHubAppName"`
	GitHubInstallID int64  `json:"gitHubInstallationId"`
	GitHubOwner     string `json:"gitHubOwner"`
	GitHubOwnerType string `json:"gitHubOwnerType"`
	OrganizationID  string `json:"organizationId"`
	CreatedAt       string `json:"createdAt"`
}

func (c *DokployClient) ListGithubProviders() ([]GithubProvider, error) {
	resp, err := c.doRequest("GET", "github.githubProviders", nil)
	if err != nil {
		return nil, err
	}

	// Try direct array response
	var providers []GithubProvider
	if err := json.Unmarshal(resp, &providers); err == nil {
		return providers, nil
	}

	// Try wrapper format
	var wrapper struct {
		Providers []GithubProvider `json:"providers"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Providers, nil
	}

	// Try githubProviders key
	var wrapper2 struct {
		Providers []GithubProvider `json:"githubProviders"`
	}
	if err := json.Unmarshal(resp, &wrapper2); err == nil {
		return wrapper2.Providers, nil
	}

	return nil, fmt.Errorf("failed to parse github providers response")
}

// --- Mount ---

type Mount struct {
	ID          string `json:"mountId"`
	Type        string `json:"type"` // bind, volume, file
	HostPath    string `json:"hostPath"`
	VolumeName  string `json:"volumeName"`
	Content     string `json:"content"`
	MountPath   string `json:"mountPath"`
	ServiceType string `json:"serviceType"` // application, postgres, mysql, mariadb, mongo, redis, compose
	FilePath    string `json:"filePath"`
	ServiceID   string `json:"serviceId"`
	// Foreign keys
	ApplicationID string `json:"applicationId"`
	PostgresID    string `json:"postgresId"`
	MariadbID     string `json:"mariadbId"`
	MongoID       string `json:"mongoId"`
	MysqlID       string `json:"mysqlId"`
	RedisID       string `json:"redisId"`
	ComposeID     string `json:"composeId"`
}

func (c *DokployClient) CreateMount(mount Mount) (*Mount, error) {
	payload := map[string]interface{}{
		"type":        mount.Type,
		"mountPath":   mount.MountPath,
		"serviceId":   mount.ServiceID,
		"serviceType": mount.ServiceType,
	}

	if mount.HostPath != "" {
		payload["hostPath"] = mount.HostPath
	}
	if mount.VolumeName != "" {
		payload["volumeName"] = mount.VolumeName
	}
	if mount.Content != "" {
		payload["content"] = mount.Content
	}
	if mount.FilePath != "" {
		payload["filePath"] = mount.FilePath
	}

	resp, err := c.doRequest("POST", "mounts.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as Mount object
	var result Mount
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	return nil, fmt.Errorf("failed to parse mount response or mount ID not set: %s", string(resp))
}

func (c *DokployClient) GetMount(id string) (*Mount, error) {
	endpoint := fmt.Sprintf("mounts.one?mountId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Mount
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateMount(mount Mount) (*Mount, error) {
	payload := map[string]interface{}{
		"mountId": mount.ID,
	}

	if mount.Type != "" {
		payload["type"] = mount.Type
	}
	if mount.HostPath != "" {
		payload["hostPath"] = mount.HostPath
	}
	if mount.VolumeName != "" {
		payload["volumeName"] = mount.VolumeName
	}
	if mount.Content != "" {
		payload["content"] = mount.Content
	}
	if mount.FilePath != "" {
		payload["filePath"] = mount.FilePath
	}
	if mount.MountPath != "" {
		payload["mountPath"] = mount.MountPath
	}
	if mount.ServiceType != "" {
		payload["serviceType"] = mount.ServiceType
	}

	resp, err := c.doRequest("POST", "mounts.update", payload)
	if err != nil {
		return nil, err
	}

	var result Mount
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteMount(id string) error {
	payload := map[string]string{
		"mountId": id,
	}
	_, err := c.doRequest("POST", "mounts.remove", payload)
	return err
}

// --- Port ---

type Port struct {
	ID            string `json:"portId"`
	PublishedPort int64  `json:"publishedPort"`
	TargetPort    int64  `json:"targetPort"`
	Protocol      string `json:"protocol"`    // tcp, udp
	PublishMode   string `json:"publishMode"` // ingress, host
	ApplicationID string `json:"applicationId"`
}

func (c *DokployClient) CreatePort(port Port) (*Port, error) {
	payload := map[string]interface{}{
		"publishedPort": port.PublishedPort,
		"targetPort":    port.TargetPort,
		"applicationId": port.ApplicationID,
	}

	if port.Protocol != "" {
		payload["protocol"] = port.Protocol
	}
	if port.PublishMode != "" {
		payload["publishMode"] = port.PublishMode
	}

	resp, err := c.doRequest("POST", "port.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as Port object
	var result Port
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	return nil, fmt.Errorf("failed to parse port response or port ID not set: %s", string(resp))
}

func (c *DokployClient) GetPort(id string) (*Port, error) {
	endpoint := fmt.Sprintf("port.one?portId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Port
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdatePort(port Port) (*Port, error) {
	payload := map[string]interface{}{
		"portId":        port.ID,
		"publishedPort": port.PublishedPort,
		"targetPort":    port.TargetPort,
	}

	if port.Protocol != "" {
		payload["protocol"] = port.Protocol
	}
	if port.PublishMode != "" {
		payload["publishMode"] = port.PublishMode
	}

	resp, err := c.doRequest("POST", "port.update", payload)
	if err != nil {
		return nil, err
	}

	var result Port
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeletePort(id string) error {
	payload := map[string]string{
		"portId": id,
	}
	_, err := c.doRequest("POST", "port.delete", payload)
	return err
}

// --- Redirect ---

type Redirect struct {
	ID            string `json:"redirectId"`
	Regex         string `json:"regex"`
	Replacement   string `json:"replacement"`
	Permanent     bool   `json:"permanent"`
	ApplicationID string `json:"applicationId"`
	CreatedAt     string `json:"createdAt"`
}

func (c *DokployClient) CreateRedirect(redirect Redirect) (*Redirect, error) {
	payload := map[string]interface{}{
		"regex":         redirect.Regex,
		"replacement":   redirect.Replacement,
		"permanent":     redirect.Permanent,
		"applicationId": redirect.ApplicationID,
	}

	resp, err := c.doRequest("POST", "redirects.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as Redirect object first
	var result Redirect
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	// API returns boolean true on success - we don't have the ID
	if string(resp) == "true" {
		return nil, fmt.Errorf("redirect created but API did not return redirect details (no ID available)")
	}

	return nil, fmt.Errorf("unexpected API response format: %s", string(resp))
}

func (c *DokployClient) GetRedirect(id string) (*Redirect, error) {
	endpoint := fmt.Sprintf("redirects.one?redirectId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Redirect
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateRedirect(redirect Redirect) (*Redirect, error) {
	payload := map[string]interface{}{
		"redirectId":  redirect.ID,
		"regex":       redirect.Regex,
		"replacement": redirect.Replacement,
		"permanent":   redirect.Permanent,
	}

	resp, err := c.doRequest("POST", "redirects.update", payload)
	if err != nil {
		return nil, err
	}

	var result Redirect
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteRedirect(id string) error {
	payload := map[string]string{
		"redirectId": id,
	}
	_, err := c.doRequest("POST", "redirects.delete", payload)
	return err
}

// --- Registry ---

type Registry struct {
	ID             string `json:"registryId"`
	RegistryName   string `json:"registryName"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	RegistryUrl    string `json:"registryUrl"`
	RegistryType   string `json:"registryType"` // cloud
	ImagePrefix    string `json:"imagePrefix"`
	ServerID       string `json:"serverId"`
	OrganizationID string `json:"organizationId"`
	CreatedAt      string `json:"createdAt"`
}

func (c *DokployClient) CreateRegistry(registry Registry) (*Registry, error) {
	payload := map[string]interface{}{
		"registryName": registry.RegistryName,
		"username":     registry.Username,
		"password":     registry.Password,
		"registryUrl":  registry.RegistryUrl,
		"registryType": registry.RegistryType,
		"imagePrefix":  registry.ImagePrefix,
	}

	if registry.ServerID != "" {
		payload["serverId"] = registry.ServerID
	}

	resp, err := c.doRequest("POST", "registry.create", payload)
	if err != nil {
		return nil, err
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetRegistry(id string) (*Registry, error) {
	endpoint := fmt.Sprintf("registry.one?registryId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateRegistry(registry Registry) (*Registry, error) {
	payload := map[string]interface{}{
		"registryId": registry.ID,
	}

	if registry.RegistryName != "" {
		payload["registryName"] = registry.RegistryName
	}
	if registry.Username != "" {
		payload["username"] = registry.Username
	}
	if registry.Password != "" {
		payload["password"] = registry.Password
	}
	if registry.RegistryUrl != "" {
		payload["registryUrl"] = registry.RegistryUrl
	}
	if registry.RegistryType != "" {
		payload["registryType"] = registry.RegistryType
	}
	if registry.ImagePrefix != "" {
		payload["imagePrefix"] = registry.ImagePrefix
	}
	if registry.ServerID != "" {
		payload["serverId"] = registry.ServerID
	}

	resp, err := c.doRequest("POST", "registry.update", payload)
	if err != nil {
		return nil, err
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteRegistry(id string) error {
	payload := map[string]string{
		"registryId": id,
	}
	_, err := c.doRequest("POST", "registry.remove", payload)
	return err
}

func (c *DokployClient) ListRegistries() ([]Registry, error) {
	resp, err := c.doRequest("GET", "registry.all", nil)
	if err != nil {
		return nil, err
	}

	var registries []Registry
	if err := json.Unmarshal(resp, &registries); err != nil {
		return nil, err
	}
	return registries, nil
}

// Destination represents a backup destination (S3, MinIO, etc.)
type Destination struct {
	DestinationID   string `json:"destinationId"`
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	OrganizationID  string `json:"organizationId"`
	CreatedAt       string `json:"createdAt"`
}

func (c *DokployClient) CreateDestination(dest Destination) (*Destination, error) {
	payload := map[string]interface{}{
		"name":            dest.Name,
		"provider":        dest.Provider,
		"accessKey":       dest.AccessKey,
		"secretAccessKey": dest.SecretAccessKey,
		"bucket":          dest.Bucket,
		"region":          dest.Region,
		"endpoint":        dest.Endpoint,
	}

	resp, err := c.doRequest("POST", "destination.create", payload)
	if err != nil {
		return nil, err
	}

	var result Destination
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetDestination(id string) (*Destination, error) {
	endpoint := fmt.Sprintf("destination.one?destinationId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Destination
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateDestination(dest Destination) (*Destination, error) {
	payload := map[string]interface{}{
		"destinationId":   dest.DestinationID,
		"name":            dest.Name,
		"provider":        dest.Provider,
		"accessKey":       dest.AccessKey,
		"secretAccessKey": dest.SecretAccessKey,
		"bucket":          dest.Bucket,
		"region":          dest.Region,
		"endpoint":        dest.Endpoint,
	}

	resp, err := c.doRequest("POST", "destination.update", payload)
	if err != nil {
		return nil, err
	}

	var result Destination
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteDestination(id string) error {
	payload := map[string]string{
		"destinationId": id,
	}
	_, err := c.doRequest("POST", "destination.remove", payload)
	return err
}

func (c *DokployClient) ListDestinations() ([]Destination, error) {
	resp, err := c.doRequest("GET", "destination.all", nil)
	if err != nil {
		return nil, err
	}

	var destinations []Destination
	if err := json.Unmarshal(resp, &destinations); err != nil {
		return nil, err
	}
	return destinations, nil
}

// Backup represents a scheduled backup configuration
type Backup struct {
	BackupID        string `json:"backupId"`
	AppName         string `json:"appName"`
	Schedule        string `json:"schedule"`
	Enabled         bool   `json:"enabled"`
	Database        string `json:"database"`
	Prefix          string `json:"prefix"`
	DestinationID   string `json:"destinationId"`
	KeepLatestCount int    `json:"keepLatestCount"`
	BackupType      string `json:"backupType"`   // "database" or "compose"
	DatabaseType    string `json:"databaseType"` // "postgres", "mysql", "mariadb", "mongo"
	PostgresID      string `json:"postgresId"`
	MysqlID         string `json:"mysqlId"`
	MariadbID       string `json:"mariadbId"`
	MongoID         string `json:"mongoId"`
	ComposeID       string `json:"composeId"`
	ServiceName     string `json:"serviceName"`
}

func (c *DokployClient) CreateBackup(backup Backup) (*Backup, error) {
	payload := map[string]interface{}{
		"schedule":      backup.Schedule,
		"enabled":       backup.Enabled,
		"prefix":        backup.Prefix,
		"destinationId": backup.DestinationID,
		"database":      backup.Database,
		"backupType":    backup.BackupType,
		"databaseType":  backup.DatabaseType,
	}

	if backup.KeepLatestCount > 0 {
		payload["keepLatestCount"] = backup.KeepLatestCount
	}

	// Add type-specific database ID
	if backup.PostgresID != "" {
		payload["postgresId"] = backup.PostgresID
	}
	if backup.MysqlID != "" {
		payload["mysqlId"] = backup.MysqlID
	}
	if backup.MariadbID != "" {
		payload["mariadbId"] = backup.MariadbID
	}
	if backup.MongoID != "" {
		payload["mongoId"] = backup.MongoID
	}
	if backup.ComposeID != "" {
		payload["composeId"] = backup.ComposeID
	}
	if backup.ServiceName != "" {
		payload["serviceName"] = backup.ServiceName
	}

	resp, err := c.doRequest("POST", "backup.create", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response from buggy Dokploy API (backup.create doesn't return the created backup)
	// WORKAROUND: Query by unique parameters to find the newly created backup
	// This works because the combination of destinationId + databaseId + databaseType should be unique per backup
	if len(resp) == 0 {
		// Sleep briefly to ensure the backup is committed to the database
		time.Sleep(500 * time.Millisecond)

		// Try to find the backup by querying all backups for this destination/database combo
		// Since we can't list all backups via API, we'll have to return an error
		return nil, fmt.Errorf("backup.create returned empty response due to Dokploy API bug (missing return statement in backup router). The backup may have been created in the database, but we cannot retrieve its ID without a 'list all backups' API endpoint. Please upgrade Dokploy to a version with the fix, or manually patch /app/apps/dokploy/server/api/routers/backup.ts line 114 to add 'return backup;'")
	}

	var result Backup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backup response (len=%d): %w. Response: %s", len(resp), err, string(resp))
	}
	return &result, nil
}

func (c *DokployClient) GetBackup(id string) (*Backup, error) {
	endpoint := fmt.Sprintf("backup.one?backupId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Backup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateBackup(backup Backup) (*Backup, error) {
	payload := map[string]interface{}{
		"backupId":      backup.BackupID,
		"schedule":      backup.Schedule,
		"enabled":       backup.Enabled,
		"prefix":        backup.Prefix,
		"destinationId": backup.DestinationID,
		"database":      backup.Database,
		"databaseType":  backup.DatabaseType,
	}

	if backup.KeepLatestCount > 0 {
		payload["keepLatestCount"] = backup.KeepLatestCount
	}
	if backup.ServiceName != "" {
		payload["serviceName"] = backup.ServiceName
	}

	resp, err := c.doRequest("POST", "backup.update", payload)
	if err != nil {
		return nil, err
	}

	var result Backup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteBackup(id string) error {
	payload := map[string]string{
		"backupId": id,
	}
	_, err := c.doRequest("POST", "backup.remove", payload)
	return err
}
