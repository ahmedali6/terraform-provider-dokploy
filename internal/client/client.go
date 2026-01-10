package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotFound is returned when a resource is not found (404).
var ErrNotFound = errors.New("resource not found")

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

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, string(respBytes))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBytes))
	}

	return respBytes, nil
}

// --- User ---

// UserDetails represents the nested user object in OrganizationMember.
type UserDetails struct {
	ID                 string   `json:"id"`
	FirstName          string   `json:"firstName"`
	LastName           string   `json:"lastName"`
	Email              string   `json:"email"`
	EmailVerified      bool     `json:"emailVerified"`
	TwoFactorEnabled   bool     `json:"twoFactorEnabled"`
	CreatedAt          string   `json:"createdAt"`
	UpdatedAt          string   `json:"updatedAt"`
	Image              *string  `json:"image"`
	Role               string   `json:"role"`
	IsRegistered       bool     `json:"isRegistered"`
	EnablePaidFeatures bool     `json:"enablePaidFeatures"`
	AllowImpersonation bool     `json:"allowImpersonation"`
	ServersQuantity    int      `json:"serversQuantity"`
	ApiKeys            []ApiKey `json:"apiKeys,omitempty"`
}

// ApiKey represents an API key.
type ApiKey struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Start               string  `json:"start"`
	Key                 string  `json:"key,omitempty"` // Only returned on creation
	UserID              string  `json:"userId"`
	Enabled             bool    `json:"enabled"`
	RateLimitEnabled    bool    `json:"rateLimitEnabled"`
	RateLimitTimeWindow int64   `json:"rateLimitTimeWindow"`
	RateLimitMax        int     `json:"rateLimitMax"`
	RequestCount        int     `json:"requestCount"`
	ExpiresAt           *string `json:"expiresAt"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
	LastRequest         *string `json:"lastRequest"`
	Metadata            *string `json:"metadata"`
}

// OrganizationMember represents a user's membership in an organization.
type OrganizationMember struct {
	ID                      string      `json:"id"` // Member ID
	OrganizationID          string      `json:"organizationId"`
	UserID                  string      `json:"userId"`
	Role                    string      `json:"role"`
	CreatedAt               string      `json:"createdAt"`
	TeamID                  *string     `json:"teamId"`
	IsDefault               bool        `json:"isDefault"`
	CanCreateProjects       bool        `json:"canCreateProjects"`
	CanAccessToSSHKeys      bool        `json:"canAccessToSSHKeys"`
	CanCreateServices       bool        `json:"canCreateServices"`
	CanDeleteProjects       bool        `json:"canDeleteProjects"`
	CanDeleteServices       bool        `json:"canDeleteServices"`
	CanAccessToDocker       bool        `json:"canAccessToDocker"`
	CanAccessToAPI          bool        `json:"canAccessToAPI"`
	CanAccessToGitProviders bool        `json:"canAccessToGitProviders"`
	CanAccessToTraefikFiles bool        `json:"canAccessToTraefikFiles"`
	CanDeleteEnvironments   bool        `json:"canDeleteEnvironments"`
	CanCreateEnvironments   bool        `json:"canCreateEnvironments"`
	AccessedProjects        []string    `json:"accessedProjects"`
	AccessedEnvironments    []string    `json:"accessedEnvironments"`
	AccessedServices        []string    `json:"accessedServices"`
	User                    UserDetails `json:"user"`
}

// User is a simplified user struct for backward compatibility.
type User struct {
	ID             string `json:"userId"`
	Email          string `json:"email"`
	OrganizationID string `json:"organizationId"`
}

// GetUser returns the basic user info (backward compatible).
func (c *DokployClient) GetUser() (*User, error) {
	member, err := c.GetCurrentMember()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:             member.UserID,
		Email:          member.User.Email,
		OrganizationID: member.OrganizationID,
	}, nil
}

// GetCurrentMember returns the full organization member info for the current user.
func (c *DokployClient) GetCurrentMember() (*OrganizationMember, error) {
	resp, err := c.doRequest("GET", "user.get", nil)
	if err != nil {
		return nil, err
	}

	var member OrganizationMember
	if err := json.Unmarshal(resp, &member); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}
	return &member, nil
}

// ListMembers returns all organization members.
func (c *DokployClient) ListMembers() ([]OrganizationMember, error) {
	resp, err := c.doRequest("GET", "user.all", nil)
	if err != nil {
		return nil, err
	}

	var members []OrganizationMember
	if err := json.Unmarshal(resp, &members); err != nil {
		return nil, fmt.Errorf("failed to parse users response: %w", err)
	}
	return members, nil
}

// GetMemberByUserID finds a member by their user ID.
func (c *DokployClient) GetMemberByUserID(userID string) (*OrganizationMember, error) {
	members, err := c.ListMembers()
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.UserID == userID {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("member with user ID %s not found", userID)
}

// GetMemberByID finds a member by their member ID.
func (c *DokployClient) GetMemberByID(memberID string) (*OrganizationMember, error) {
	members, err := c.ListMembers()
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.ID == memberID {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("member with ID %s not found", memberID)
}

// UserPermissionsInput represents the input for assigning permissions.
type UserPermissionsInput struct {
	MemberID                string   `json:"id"`
	AccessedProjects        []string `json:"accessedProjects"`
	AccessedEnvironments    []string `json:"accessedEnvironments"`
	AccessedServices        []string `json:"accessedServices"`
	CanCreateProjects       bool     `json:"canCreateProjects"`
	CanCreateServices       bool     `json:"canCreateServices"`
	CanDeleteProjects       bool     `json:"canDeleteProjects"`
	CanDeleteServices       bool     `json:"canDeleteServices"`
	CanAccessToDocker       bool     `json:"canAccessToDocker"`
	CanAccessToTraefikFiles bool     `json:"canAccessToTraefikFiles"`
	CanAccessToAPI          bool     `json:"canAccessToAPI"`
	CanAccessToSSHKeys      bool     `json:"canAccessToSSHKeys"`
	CanAccessToGitProviders bool     `json:"canAccessToGitProviders"`
	CanDeleteEnvironments   bool     `json:"canDeleteEnvironments"`
	CanCreateEnvironments   bool     `json:"canCreateEnvironments"`
}

// AssignUserPermissions assigns permissions to a member.
func (c *DokployClient) AssignUserPermissions(input UserPermissionsInput) error {
	payload := map[string]interface{}{
		"id":                      input.MemberID,
		"accessedProjects":        input.AccessedProjects,
		"accessedEnvironments":    input.AccessedEnvironments,
		"accessedServices":        input.AccessedServices,
		"canCreateProjects":       input.CanCreateProjects,
		"canCreateServices":       input.CanCreateServices,
		"canDeleteProjects":       input.CanDeleteProjects,
		"canDeleteServices":       input.CanDeleteServices,
		"canAccessToDocker":       input.CanAccessToDocker,
		"canAccessToTraefikFiles": input.CanAccessToTraefikFiles,
		"canAccessToAPI":          input.CanAccessToAPI,
		"canAccessToSSHKeys":      input.CanAccessToSSHKeys,
		"canAccessToGitProviders": input.CanAccessToGitProviders,
		"canDeleteEnvironments":   input.CanDeleteEnvironments,
		"canCreateEnvironments":   input.CanCreateEnvironments,
	}

	_, err := c.doRequest("POST", "user.assignPermissions", payload)
	return err
}

// ApiKeyCreateInput represents the input for creating an API key.
type ApiKeyCreateInput struct {
	Name                string            `json:"name"`
	Metadata            map[string]string `json:"metadata"`
	ExpiresIn           *int64            `json:"expiresIn,omitempty"` // In seconds, min 86400 (1 day)
	RateLimitEnabled    *bool             `json:"rateLimitEnabled,omitempty"`
	RateLimitMax        *int              `json:"rateLimitMax,omitempty"`
	RateLimitTimeWindow *int64            `json:"rateLimitTimeWindow,omitempty"` // In milliseconds
}

// CreateApiKey creates a new API key.
func (c *DokployClient) CreateApiKey(input ApiKeyCreateInput) (*ApiKey, error) {
	payload := map[string]interface{}{
		"name":     input.Name,
		"metadata": input.Metadata,
	}

	if input.ExpiresIn != nil {
		payload["expiresIn"] = *input.ExpiresIn
	}
	if input.RateLimitEnabled != nil {
		payload["rateLimitEnabled"] = *input.RateLimitEnabled
	}
	if input.RateLimitMax != nil {
		payload["rateLimitMax"] = *input.RateLimitMax
	}
	if input.RateLimitTimeWindow != nil {
		payload["rateLimitTimeWindow"] = *input.RateLimitTimeWindow
	}

	resp, err := c.doRequest("POST", "user.createApiKey", payload)
	if err != nil {
		return nil, err
	}

	var apiKey ApiKey
	if err := json.Unmarshal(resp, &apiKey); err != nil {
		return nil, fmt.Errorf("failed to parse API key response: %w", err)
	}
	return &apiKey, nil
}

// DeleteApiKey deletes an API key.
func (c *DokployClient) DeleteApiKey(apiKeyID string) error {
	payload := map[string]string{
		"apiKeyId": apiKeyID,
	}
	_, err := c.doRequest("POST", "user.deleteApiKey", payload)
	return err
}

// GetApiKeyByID retrieves an API key by ID from the current user's API keys.
func (c *DokployClient) GetApiKeyByID(apiKeyID string) (*ApiKey, error) {
	member, err := c.GetCurrentMember()
	if err != nil {
		return nil, err
	}

	for _, key := range member.User.ApiKeys {
		if key.ID == apiKeyID {
			return &key, nil
		}
	}
	return nil, fmt.Errorf("API key with ID %s not found", apiKeyID)
}

// --- AI ---

// AI represents an AI provider configuration.
type AI struct {
	ID             string `json:"aiId"`
	Name           string `json:"name"`
	ApiURL         string `json:"apiUrl"`
	ApiKey         string `json:"apiKey"`
	Model          string `json:"model"`
	IsEnabled      bool   `json:"isEnabled"`
	OrganizationID string `json:"organizationId"`
	CreatedAt      string `json:"createdAt"`
}

// AIModel represents a model available from an AI provider.
type AIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// CreateAI creates a new AI provider configuration.
func (c *DokployClient) CreateAI(name, apiURL, apiKey, model string, isEnabled bool) (*AI, error) {
	payload := map[string]interface{}{
		"name":      name,
		"apiUrl":    apiURL,
		"apiKey":    apiKey,
		"model":     model,
		"isEnabled": isEnabled,
	}

	// Record time before creation to help identify the new resource
	creationTime := time.Now().Add(-1 * time.Second)

	_, err := c.doRequest("POST", "ai.create", payload)
	if err != nil {
		return nil, err
	}

	// API returns empty array on success, need to fetch the created AI
	ais, err := c.ListAIs()
	if err != nil {
		return nil, err
	}

	// Find the newly created AI by name and creation time
	// Look for the most recently created AI with matching name that was created after our timestamp
	var bestMatch *AI
	var bestMatchTime time.Time
	for i := range ais {
		if ais[i].Name == name {
			aiCreatedAt, parseErr := time.Parse(time.RFC3339, ais[i].CreatedAt)
			if parseErr != nil {
				continue
			}
			if aiCreatedAt.After(creationTime) && (bestMatch == nil || aiCreatedAt.After(bestMatchTime)) {
				bestMatch = &ais[i]
				bestMatchTime = aiCreatedAt
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}

	return nil, fmt.Errorf("failed to find created AI configuration")
}

// GetAI retrieves an AI configuration by ID.
func (c *DokployClient) GetAI(aiID string) (*AI, error) {
	endpoint := fmt.Sprintf("ai.get?aiId=%s", aiID)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var ai AI
	if err := json.Unmarshal(resp, &ai); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	return &ai, nil
}

// ListAIs returns all AI configurations.
func (c *DokployClient) ListAIs() ([]AI, error) {
	resp, err := c.doRequest("GET", "ai.getAll", nil)
	if err != nil {
		return nil, err
	}

	var ais []AI
	if err := json.Unmarshal(resp, &ais); err != nil {
		return nil, fmt.Errorf("failed to parse AIs response: %w", err)
	}
	return ais, nil
}

// UpdateAI updates an AI configuration. Note: API requires all fields.
func (c *DokployClient) UpdateAI(ai AI) error {
	payload := map[string]interface{}{
		"aiId":      ai.ID,
		"name":      ai.Name,
		"apiUrl":    ai.ApiURL,
		"apiKey":    ai.ApiKey,
		"model":     ai.Model,
		"isEnabled": ai.IsEnabled,
	}

	_, err := c.doRequest("POST", "ai.update", payload)
	return err
}

// DeleteAI deletes an AI configuration.
func (c *DokployClient) DeleteAI(aiID string) error {
	payload := map[string]string{
		"aiId": aiID,
	}
	_, err := c.doRequest("POST", "ai.delete", payload)
	return err
}

// GetAIModels retrieves available models from an AI provider.
func (c *DokployClient) GetAIModels(apiURL, apiKey string) ([]AIModel, error) {
	// URL encode the parameters to handle special characters safely
	endpoint := fmt.Sprintf("ai.getModels?apiUrl=%s&apiKey=%s", url.QueryEscape(apiURL), url.QueryEscape(apiKey))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var models []AIModel
	if err := json.Unmarshal(resp, &models); err != nil {
		return nil, fmt.Errorf("failed to parse AI models response: %w", err)
	}
	return models, nil
}

// --- Certificate ---

// Certificate represents a TLS certificate in Dokploy.
type Certificate struct {
	ID              string  `json:"certificateId"`
	Name            string  `json:"name"`
	CertificateData string  `json:"certificateData"`
	PrivateKey      string  `json:"privateKey"`
	CertificatePath string  `json:"certificatePath"`
	AutoRenew       *bool   `json:"autoRenew"`
	OrganizationID  string  `json:"organizationId"`
	ServerID        *string `json:"serverId"`
}

// CreateCertificate creates a new TLS certificate.
func (c *DokployClient) CreateCertificate(cert Certificate) (*Certificate, error) {
	payload := map[string]interface{}{
		"name":            cert.Name,
		"certificateData": cert.CertificateData,
		"privateKey":      cert.PrivateKey,
		"organizationId":  cert.OrganizationID,
	}

	if cert.CertificatePath != "" {
		payload["certificatePath"] = cert.CertificatePath
	}
	if cert.AutoRenew != nil {
		payload["autoRenew"] = *cert.AutoRenew
	}
	if cert.ServerID != nil && *cert.ServerID != "" {
		payload["serverId"] = *cert.ServerID
	}

	resp, err := c.doRequest("POST", "certificates.create", payload)
	if err != nil {
		return nil, err
	}

	var result Certificate
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse certificate response: %w", err)
	}
	return &result, nil
}

// GetCertificate retrieves a certificate by ID.
func (c *DokployClient) GetCertificate(id string) (*Certificate, error) {
	endpoint := fmt.Sprintf("certificates.one?certificateId=%s", url.QueryEscape(id))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Certificate
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse certificate response: %w", err)
	}
	return &result, nil
}

// ListCertificates returns all certificates.
func (c *DokployClient) ListCertificates() ([]Certificate, error) {
	resp, err := c.doRequest("GET", "certificates.all", nil)
	if err != nil {
		return nil, err
	}

	var certs []Certificate
	if err := json.Unmarshal(resp, &certs); err != nil {
		return nil, fmt.Errorf("failed to parse certificates response: %w", err)
	}
	return certs, nil
}

// DeleteCertificate deletes a certificate by ID.
func (c *DokployClient) DeleteCertificate(id string) error {
	payload := map[string]string{
		"certificateId": id,
	}
	_, err := c.doRequest("POST", "certificates.remove", payload)
	return err
}

// GetCurrentOrganizationID retrieves the organization ID for the current user.
func (c *DokployClient) GetCurrentOrganizationID() (string, error) {
	resp, err := c.doRequest("GET", "user.get", nil)
	if err != nil {
		return "", err
	}

	var result struct {
		OrganizationID string `json:"organizationId"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("failed to parse user response: %w", err)
	}
	return result.OrganizationID, nil
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
	CleanCache         bool   `json:"cleanCache"`

	// GitHub provider settings (application.saveGithubProvider)
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Owner       string `json:"owner"`
	BuildPath   string `json:"buildPath"`
	GithubId    string `json:"githubId"`
	TriggerType string `json:"triggerType"` // push, tag

	// GitLab provider settings (application.saveGitlabProvider)
	GitlabId            string `json:"gitlabId"`
	GitlabProjectId     int64  `json:"gitlabProjectId"`
	GitlabRepository    string `json:"gitlabRepository"`
	GitlabOwner         string `json:"gitlabOwner"`
	GitlabBranch        string `json:"gitlabBranch"`
	GitlabBuildPath     string `json:"gitlabBuildPath"`
	GitlabPathNamespace string `json:"gitlabPathNamespace"`

	// Bitbucket provider settings (application.saveBitbucketProvider)
	BitbucketId         string `json:"bitbucketId"`
	BitbucketRepository string `json:"bitbucketRepository"`
	BitbucketOwner      string `json:"bitbucketOwner"`
	BitbucketBranch     string `json:"bitbucketBranch"`
	BitbucketBuildPath  string `json:"bitbucketBuildPath"`

	// Gitea provider settings (application.saveGiteaProvider)
	GiteaId         string `json:"giteaId"`
	GiteaRepository string `json:"giteaRepository"`
	GiteaOwner      string `json:"giteaOwner"`
	GiteaBranch     string `json:"giteaBranch"`
	GiteaBuildPath  string `json:"giteaBuildPath"`

	// Docker provider settings (application.saveDockerProvider)
	DockerImage string `json:"dockerImage"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RegistryUrl string `json:"registryUrl"`
	RegistryId  string `json:"registryId"`

	// Build type settings (application.saveBuildType)
	BuildType         string `json:"buildType"` // dockerfile, heroku_buildpacks, paketo_buildpacks, nixpacks, static, railpack
	DockerfilePath    string `json:"dockerfile"`
	DockerContextPath string `json:"dockerContextPath"`
	DockerBuildStage  string `json:"dockerBuildStage"`
	PublishDirectory  string `json:"publishDirectory"`
	Dockerfile        string `json:"dockerfileContent"` // Raw Dockerfile content for drop source
	DropBuildPath     string `json:"dropBuildPath"`     // Build path for "drop" source type
	HerokuVersion     string `json:"herokuVersion"`
	RailpackVersion   string `json:"railpackVersion"`
	IsStaticSpa       bool   `json:"isStaticSpa"`

	// Environment settings (application.saveEnvironment)
	Env           string `json:"env"`
	BuildArgs     string `json:"buildArgs"`
	BuildSecrets  string `json:"buildSecrets"`
	CreateEnvFile bool   `json:"createEnvFile"`

	// Runtime configuration (application.update)
	// Note: The API accepts and returns memoryLimit/memoryReservation/cpuLimit/cpuReservation as strings
	AutoDeploy        bool        `json:"autoDeploy"`
	Replicas          int         `json:"replicas"`
	MemoryLimit       json.Number `json:"memoryLimit"`
	MemoryReservation json.Number `json:"memoryReservation"`
	CpuLimit          json.Number `json:"cpuLimit"`
	CpuReservation    json.Number `json:"cpuReservation"`
	Command           string      `json:"command"`
	Args              string      `json:"args"`
	EntryPoint        string      `json:"entrypoint"`

	// Docker Swarm configuration
	HealthCheckSwarm     map[string]interface{}   `json:"healthCheckSwarm"`
	RestartPolicySwarm   map[string]interface{}   `json:"restartPolicySwarm"`
	PlacementSwarm       map[string]interface{}   `json:"placementSwarm"`
	UpdateConfigSwarm    map[string]interface{}   `json:"updateConfigSwarm"`
	RollbackConfigSwarm  map[string]interface{}   `json:"rollbackConfigSwarm"`
	ModeSwarm            map[string]interface{}   `json:"modeSwarm"`
	LabelsSwarm          map[string]interface{}   `json:"labelsSwarm"`
	NetworkSwarm         []map[string]interface{} `json:"networkSwarm"`
	StopGracePeriodSwarm *int64                   `json:"stopGracePeriodSwarm"`
	EndpointSpecSwarm    map[string]interface{}   `json:"endpointSpecSwarm"`

	// Preview deployments (application.update)
	IsPreviewDeploymentsActive            bool   `json:"isPreviewDeploymentsActive"`
	PreviewEnv                            string `json:"previewEnv"`
	PreviewBuildArgs                      string `json:"previewBuildArgs"`
	PreviewBuildSecrets                   string `json:"previewBuildSecrets"`
	PreviewLabels                         string `json:"previewLabels"`
	PreviewWildcard                       string `json:"previewWildcard"`
	PreviewPort                           int64  `json:"previewPort"`
	PreviewHttps                          bool   `json:"previewHttps"`
	PreviewPath                           string `json:"previewPath"`
	PreviewCertificateType                string `json:"previewCertificateType"`
	PreviewCustomCertResolver             string `json:"previewCustomCertResolver"`
	PreviewLimit                          int64  `json:"previewLimit"`
	PreviewRequireCollaboratorPermissions bool   `json:"previewRequireCollaboratorPermissions"`

	// Rollback configuration
	RollbackActive     bool   `json:"rollbackActive"`
	RollbackRegistryId string `json:"rollbackRegistryId"`

	// Build server configuration
	BuildServerId   string `json:"buildServerId"`
	BuildRegistryId string `json:"buildRegistryId"`

	// Display settings
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Enabled  bool   `json:"enabled"`

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
	// API expects memoryLimit/memoryReservation/cpuLimit/cpuReservation as strings
	if app.MemoryLimit != "" {
		payload["memoryLimit"] = string(app.MemoryLimit)
	}
	if app.MemoryReservation != "" {
		payload["memoryReservation"] = string(app.MemoryReservation)
	}
	if app.CpuLimit != "" {
		payload["cpuLimit"] = string(app.CpuLimit)
	}
	if app.CpuReservation != "" {
		payload["cpuReservation"] = string(app.CpuReservation)
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

// ReadTraefikConfig retrieves the custom Traefik configuration for an application.
func (c *DokployClient) ReadTraefikConfig(appID string) (string, error) {
	endpoint := fmt.Sprintf("application.readTraefikConfig?applicationId=%s", url.QueryEscape(appID))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	// API returns a JSON string (quoted), so we need to unmarshal it
	var config string
	if err := json.Unmarshal(resp, &config); err != nil {
		// If unmarshal fails, it might be null/empty
		if string(resp) == "null" || string(resp) == "" {
			return "", nil
		}
		return "", fmt.Errorf("failed to parse Traefik config response: %w", err)
	}
	return config, nil
}

// UpdateTraefikConfig updates the custom Traefik configuration for an application.
func (c *DokployClient) UpdateTraefikConfig(appID, traefikConfig string) error {
	payload := map[string]string{
		"applicationId": appID,
		"traefikConfig": traefikConfig,
	}
	_, err := c.doRequest("POST", "application.updateTraefikConfig", payload)
	return err
}

// MoveApplication moves an application to a different environment.
func (c *DokployClient) MoveApplication(appID, targetEnvironmentID string) (*Application, error) {
	payload := map[string]string{
		"applicationId":       appID,
		"targetEnvironmentId": targetEnvironmentID,
	}
	resp, err := c.doRequest("POST", "application.move", payload)
	if err != nil {
		return nil, err
	}

	var app Application
	if err := json.Unmarshal(resp, &app); err != nil {
		return nil, fmt.Errorf("failed to parse application response: %w", err)
	}
	return &app, nil
}

// ListApplications retrieves all applications. Uses project.all and extracts applications from all environments.
func (c *DokployClient) ListApplications() ([]Application, error) {
	resp, err := c.doRequest("GET", "project.all", nil)
	if err != nil {
		return nil, err
	}

	var projects []struct {
		Environments []struct {
			Applications []Application `json:"applications"`
		} `json:"environments"`
	}
	if err := json.Unmarshal(resp, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	var apps []Application
	for _, proj := range projects {
		for _, env := range proj.Environments {
			apps = append(apps, env.Applications...)
		}
	}
	return apps, nil
}

// ListApplicationsByEnvironment retrieves all applications in a specific environment.
func (c *DokployClient) ListApplicationsByEnvironment(environmentID string) ([]Application, error) {
	// First get the environment to find its project
	endpoint := fmt.Sprintf("environment.one?environmentId=%s", url.QueryEscape(environmentID))
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var env struct {
		Applications []Application `json:"applications"`
	}
	if err := json.Unmarshal(resp, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment response: %w", err)
	}

	return env.Applications, nil
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

// SaveGitlabProviderInput contains all the fields for the saveGitlabProvider endpoint.
type SaveGitlabProviderInput struct {
	ApplicationID       string
	GitlabId            string
	GitlabProjectId     int64
	GitlabRepository    string
	GitlabOwner         string
	GitlabBranch        string
	GitlabBuildPath     string
	GitlabPathNamespace string
	WatchPaths          []string
	EnableSubmodules    bool
}

// SaveGitlabProvider configures the GitLab provider settings for an application.
// Corresponds to application.saveGitlabProvider endpoint.
func (c *DokployClient) SaveGitlabProvider(input SaveGitlabProviderInput) error {
	payload := map[string]interface{}{
		"applicationId":    input.ApplicationID,
		"enableSubmodules": input.EnableSubmodules,
	}

	if input.GitlabId != "" {
		payload["gitlabId"] = input.GitlabId
	} else {
		payload["gitlabId"] = nil
	}

	if input.GitlabProjectId != 0 {
		payload["gitlabProjectId"] = input.GitlabProjectId
	}
	if input.GitlabRepository != "" {
		payload["gitlabRepository"] = input.GitlabRepository
	}
	if input.GitlabOwner != "" {
		payload["gitlabOwner"] = input.GitlabOwner
	}
	if input.GitlabBranch != "" {
		payload["gitlabBranch"] = input.GitlabBranch
	}
	if input.GitlabBuildPath != "" {
		payload["gitlabBuildPath"] = input.GitlabBuildPath
	}
	if input.GitlabPathNamespace != "" {
		payload["gitlabPathNamespace"] = input.GitlabPathNamespace
	}
	if len(input.WatchPaths) > 0 {
		payload["watchPaths"] = input.WatchPaths
	}

	_, err := c.doRequest("POST", "application.saveGitlabProvider", payload)
	return err
}

// SaveBitbucketProviderInput contains all the fields for the saveBitbucketProvider endpoint.
type SaveBitbucketProviderInput struct {
	ApplicationID       string
	BitbucketId         string
	BitbucketRepository string
	BitbucketOwner      string
	BitbucketBranch     string
	BitbucketBuildPath  string
	WatchPaths          []string
	EnableSubmodules    bool
}

// SaveBitbucketProvider configures the Bitbucket provider settings for an application.
// Corresponds to application.saveBitbucketProvider endpoint.
func (c *DokployClient) SaveBitbucketProvider(input SaveBitbucketProviderInput) error {
	payload := map[string]interface{}{
		"applicationId":    input.ApplicationID,
		"enableSubmodules": input.EnableSubmodules,
	}

	if input.BitbucketId != "" {
		payload["bitbucketId"] = input.BitbucketId
	} else {
		payload["bitbucketId"] = nil
	}

	if input.BitbucketRepository != "" {
		payload["bitbucketRepository"] = input.BitbucketRepository
	}
	if input.BitbucketOwner != "" {
		payload["bitbucketOwner"] = input.BitbucketOwner
	}
	if input.BitbucketBranch != "" {
		payload["bitbucketBranch"] = input.BitbucketBranch
	}
	if input.BitbucketBuildPath != "" {
		payload["bitbucketBuildPath"] = input.BitbucketBuildPath
	}
	if len(input.WatchPaths) > 0 {
		payload["watchPaths"] = input.WatchPaths
	}

	_, err := c.doRequest("POST", "application.saveBitbucketProvider", payload)
	return err
}

// SaveGiteaProviderInput contains all the fields for the saveGiteaProvider endpoint.
type SaveGiteaProviderInput struct {
	ApplicationID    string
	GiteaId          string
	GiteaRepository  string
	GiteaOwner       string
	GiteaBranch      string
	GiteaBuildPath   string
	WatchPaths       []string
	EnableSubmodules bool
}

// SaveGiteaProvider configures the Gitea provider settings for an application.
// Corresponds to application.saveGiteaProvider endpoint.
func (c *DokployClient) SaveGiteaProvider(input SaveGiteaProviderInput) error {
	payload := map[string]interface{}{
		"applicationId":    input.ApplicationID,
		"enableSubmodules": input.EnableSubmodules,
	}

	if input.GiteaId != "" {
		payload["giteaId"] = input.GiteaId
	} else {
		payload["giteaId"] = nil
	}

	if input.GiteaRepository != "" {
		payload["giteaRepository"] = input.GiteaRepository
	}
	if input.GiteaOwner != "" {
		payload["giteaOwner"] = input.GiteaOwner
	}
	if input.GiteaBranch != "" {
		payload["giteaBranch"] = input.GiteaBranch
	}
	if input.GiteaBuildPath != "" {
		payload["giteaBuildPath"] = input.GiteaBuildPath
	}
	if len(input.WatchPaths) > 0 {
		payload["watchPaths"] = input.WatchPaths
	}

	_, err := c.doRequest("POST", "application.saveGiteaProvider", payload)
	return err
}

// SaveDockerProviderInput contains all the fields for the saveDockerProvider endpoint.
type SaveDockerProviderInput struct {
	ApplicationID string
	DockerImage   string
	Username      string
	Password      string
	RegistryUrl   string
	RegistryId    string
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
	if input.RegistryId != "" {
		payload["registryId"] = input.RegistryId
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
	ID            string `json:"composeId"`
	Name          string `json:"name"`
	AppName       string `json:"appName"`
	Description   string `json:"description"`
	ProjectID     string `json:"projectId"`
	EnvironmentID string `json:"environmentId"`
	ServerID      string `json:"serverId"`

	// Compose file content (for raw source type)
	ComposeFile string `json:"composeFile"`
	ComposePath string `json:"composePath"`
	ComposeType string `json:"composeType"` // docker-compose or stack

	// Source configuration
	SourceType string `json:"sourceType"` // github, gitlab, bitbucket, git, raw

	// Custom Git provider settings
	CustomGitUrl       string   `json:"customGitUrl"`
	CustomGitBranch    string   `json:"customGitBranch"`
	CustomGitSSHKeyId  string   `json:"customGitSSHKeyId"`
	CustomGitBuildPath string   `json:"customGitBuildPath"`
	EnableSubmodules   bool     `json:"enableSubmodules"`
	WatchPaths         []string `json:"watchPaths"`

	// GitHub provider settings
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Owner       string `json:"owner"`
	GithubId    string `json:"githubId"`
	TriggerType string `json:"triggerType"`

	// GitLab provider settings
	GitlabId            string `json:"gitlabId"`
	GitlabProjectId     int64  `json:"gitlabProjectId"`
	GitlabRepository    string `json:"gitlabRepository"`
	GitlabOwner         string `json:"gitlabOwner"`
	GitlabBranch        string `json:"gitlabBranch"`
	GitlabBuildPath     string `json:"gitlabBuildPath"`
	GitlabPathNamespace string `json:"gitlabPathNamespace"`

	// Bitbucket provider settings
	BitbucketId         string `json:"bitbucketId"`
	BitbucketRepository string `json:"bitbucketRepository"`
	BitbucketOwner      string `json:"bitbucketOwner"`
	BitbucketBranch     string `json:"bitbucketBranch"`
	BitbucketBuildPath  string `json:"bitbucketBuildPath"`

	// Gitea provider settings
	GiteaId         string `json:"giteaId"`
	GiteaRepository string `json:"giteaRepository"`
	GiteaOwner      string `json:"giteaOwner"`
	GiteaBranch     string `json:"giteaBranch"`
	GiteaBuildPath  string `json:"giteaBuildPath"`

	// Runtime configuration
	AutoDeploy bool `json:"autoDeploy"`
	Replicas   int  `json:"replicas"`

	// Advanced configuration
	Command                   string `json:"command"`
	Suffix                    string `json:"suffix"`
	Randomize                 bool   `json:"randomize"`
	IsolatedDeployment        bool   `json:"isolatedDeployment"`
	IsolatedDeploymentsVolume bool   `json:"isolatedDeploymentsVolume"`

	// Environment
	Env string `json:"env"`

	// Status
	ComposeStatus string `json:"composeStatus"`

	// Webhook token
	RefreshToken string `json:"refreshToken"`

	// Domains
	Domains []Domain `json:"domains"`

	// Timestamps
	CreatedAt string `json:"createdAt"`
}

func (c *DokployClient) CreateCompose(comp Compose) (*Compose, error) {
	// 1. Create compose with serverId
	composeType := comp.ComposeType
	if composeType == "" {
		composeType = "docker-compose"
	}

	payload := map[string]interface{}{
		"environmentId": comp.EnvironmentID,
		"name":          comp.Name,
		"composeType":   composeType,
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

	// 2. Update with Git configuration if necessary.
	updatePayload := map[string]interface{}{
		"composeId":  createdComp.ID,
		"name":       comp.Name,
		"sourceType": comp.SourceType,
		"autoDeploy": comp.AutoDeploy,
	}

	// Description.
	if comp.Description != "" {
		updatePayload["description"] = comp.Description
	}

	// Custom Git provider settings.
	if comp.CustomGitUrl != "" {
		updatePayload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		updatePayload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		updatePayload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.CustomGitBuildPath != "" {
		updatePayload["customGitBuildPath"] = comp.CustomGitBuildPath
	}
	if comp.ComposePath != "" {
		updatePayload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		updatePayload["composeFile"] = comp.ComposeFile
	}
	if comp.EnableSubmodules {
		updatePayload["enableSubmodules"] = comp.EnableSubmodules
	}

	// GitHub provider settings.
	if comp.GithubId != "" {
		updatePayload["githubId"] = comp.GithubId
	}
	if comp.Repository != "" {
		updatePayload["repository"] = comp.Repository
	}
	if comp.Owner != "" {
		updatePayload["owner"] = comp.Owner
	}
	if comp.Branch != "" {
		updatePayload["branch"] = comp.Branch
	}
	if comp.TriggerType != "" {
		updatePayload["triggerType"] = comp.TriggerType
	}

	// GitLab provider settings.
	if comp.GitlabId != "" {
		updatePayload["gitlabId"] = comp.GitlabId
	}
	if comp.GitlabProjectId > 0 {
		updatePayload["gitlabProjectId"] = comp.GitlabProjectId
	}
	if comp.GitlabRepository != "" {
		updatePayload["gitlabRepository"] = comp.GitlabRepository
	}
	if comp.GitlabOwner != "" {
		updatePayload["gitlabOwner"] = comp.GitlabOwner
	}
	if comp.GitlabBranch != "" {
		updatePayload["gitlabBranch"] = comp.GitlabBranch
	}
	if comp.GitlabBuildPath != "" {
		updatePayload["gitlabBuildPath"] = comp.GitlabBuildPath
	}
	if comp.GitlabPathNamespace != "" {
		updatePayload["gitlabPathNamespace"] = comp.GitlabPathNamespace
	}

	// Bitbucket provider settings.
	if comp.BitbucketId != "" {
		updatePayload["bitbucketId"] = comp.BitbucketId
	}
	if comp.BitbucketRepository != "" {
		updatePayload["bitbucketRepository"] = comp.BitbucketRepository
	}
	if comp.BitbucketOwner != "" {
		updatePayload["bitbucketOwner"] = comp.BitbucketOwner
	}
	if comp.BitbucketBranch != "" {
		updatePayload["bitbucketBranch"] = comp.BitbucketBranch
	}
	if comp.BitbucketBuildPath != "" {
		updatePayload["bitbucketBuildPath"] = comp.BitbucketBuildPath
	}

	// Gitea provider settings.
	if comp.GiteaId != "" {
		updatePayload["giteaId"] = comp.GiteaId
	}
	if comp.GiteaRepository != "" {
		updatePayload["giteaRepository"] = comp.GiteaRepository
	}
	if comp.GiteaOwner != "" {
		updatePayload["giteaOwner"] = comp.GiteaOwner
	}
	if comp.GiteaBranch != "" {
		updatePayload["giteaBranch"] = comp.GiteaBranch
	}
	if comp.GiteaBuildPath != "" {
		updatePayload["giteaBuildPath"] = comp.GiteaBuildPath
	}

	// Environment variables.
	if comp.Env != "" {
		updatePayload["env"] = comp.Env
	}

	// Advanced configuration
	if comp.Command != "" {
		updatePayload["command"] = comp.Command
	}
	if comp.Suffix != "" {
		updatePayload["suffix"] = comp.Suffix
	}
	// Always send boolean fields to ensure false values are communicated
	updatePayload["randomize"] = comp.Randomize
	updatePayload["isolatedDeployment"] = comp.IsolatedDeployment
	updatePayload["isolatedDeploymentsVolume"] = comp.IsolatedDeploymentsVolume
	// Send watchPaths if not nil (allows clearing by sending empty array)
	if comp.WatchPaths != nil {
		updatePayload["watchPaths"] = comp.WatchPaths
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

	// Description.
	if comp.Description != "" {
		payload["description"] = comp.Description
	}

	// Custom Git provider settings.
	if comp.CustomGitUrl != "" {
		payload["customGitUrl"] = comp.CustomGitUrl
	}
	if comp.CustomGitBranch != "" {
		payload["customGitBranch"] = comp.CustomGitBranch
	}
	if comp.CustomGitSSHKeyId != "" {
		payload["customGitSSHKeyId"] = comp.CustomGitSSHKeyId
	}
	if comp.CustomGitBuildPath != "" {
		payload["customGitBuildPath"] = comp.CustomGitBuildPath
	}
	if comp.ComposePath != "" {
		payload["composePath"] = comp.ComposePath
	}
	if comp.ComposeFile != "" {
		payload["composeFile"] = comp.ComposeFile
	}
	if comp.EnableSubmodules {
		payload["enableSubmodules"] = comp.EnableSubmodules
	}

	// GitHub provider settings.
	if comp.GithubId != "" {
		payload["githubId"] = comp.GithubId
	}
	if comp.Repository != "" {
		payload["repository"] = comp.Repository
	}
	if comp.Owner != "" {
		payload["owner"] = comp.Owner
	}
	if comp.Branch != "" {
		payload["branch"] = comp.Branch
	}
	if comp.TriggerType != "" {
		payload["triggerType"] = comp.TriggerType
	}

	// GitLab provider settings.
	if comp.GitlabId != "" {
		payload["gitlabId"] = comp.GitlabId
	}
	if comp.GitlabProjectId > 0 {
		payload["gitlabProjectId"] = comp.GitlabProjectId
	}
	if comp.GitlabRepository != "" {
		payload["gitlabRepository"] = comp.GitlabRepository
	}
	if comp.GitlabOwner != "" {
		payload["gitlabOwner"] = comp.GitlabOwner
	}
	if comp.GitlabBranch != "" {
		payload["gitlabBranch"] = comp.GitlabBranch
	}
	if comp.GitlabBuildPath != "" {
		payload["gitlabBuildPath"] = comp.GitlabBuildPath
	}
	if comp.GitlabPathNamespace != "" {
		payload["gitlabPathNamespace"] = comp.GitlabPathNamespace
	}

	// Bitbucket provider settings.
	if comp.BitbucketId != "" {
		payload["bitbucketId"] = comp.BitbucketId
	}
	if comp.BitbucketRepository != "" {
		payload["bitbucketRepository"] = comp.BitbucketRepository
	}
	if comp.BitbucketOwner != "" {
		payload["bitbucketOwner"] = comp.BitbucketOwner
	}
	if comp.BitbucketBranch != "" {
		payload["bitbucketBranch"] = comp.BitbucketBranch
	}
	if comp.BitbucketBuildPath != "" {
		payload["bitbucketBuildPath"] = comp.BitbucketBuildPath
	}

	// Gitea provider settings.
	if comp.GiteaId != "" {
		payload["giteaId"] = comp.GiteaId
	}
	if comp.GiteaRepository != "" {
		payload["giteaRepository"] = comp.GiteaRepository
	}
	if comp.GiteaOwner != "" {
		payload["giteaOwner"] = comp.GiteaOwner
	}
	if comp.GiteaBranch != "" {
		payload["giteaBranch"] = comp.GiteaBranch
	}
	if comp.GiteaBuildPath != "" {
		payload["giteaBuildPath"] = comp.GiteaBuildPath
	}

	// Environment variables.
	if comp.Env != "" {
		payload["env"] = comp.Env
	}

	// Advanced configuration
	if comp.Command != "" {
		payload["command"] = comp.Command
	}
	if comp.Suffix != "" {
		payload["suffix"] = comp.Suffix
	}
	// These boolean fields should always be sent if set
	payload["randomize"] = comp.Randomize
	payload["isolatedDeployment"] = comp.IsolatedDeployment
	payload["isolatedDeploymentsVolume"] = comp.IsolatedDeploymentsVolume
	// Send watchPaths if not nil (allows clearing by sending empty array)
	if comp.WatchPaths != nil {
		payload["watchPaths"] = comp.WatchPaths
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

// MoveCompose moves a compose to a different environment.
func (c *DokployClient) MoveCompose(composeID, targetEnvironmentID string) (*Compose, error) {
	payload := map[string]string{
		"composeId":           composeID,
		"targetEnvironmentId": targetEnvironmentID,
	}
	resp, err := c.doRequest("POST", "compose.move", payload)
	if err != nil {
		return nil, err
	}
	var result Compose
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListComposes retrieves all composes, optionally filtered by environment ID.
func (c *DokployClient) ListComposes(environmentID string) ([]Compose, error) {
	// Composes are retrieved via project.all API
	resp, err := c.doRequest("GET", "project.all", nil)
	if err != nil {
		return nil, err
	}

	var projects []struct {
		Environments []struct {
			EnvironmentID string    `json:"environmentId"`
			Compose       []Compose `json:"compose"`
		} `json:"environments"`
	}
	if err := json.Unmarshal(resp, &projects); err != nil {
		return nil, err
	}

	var composes []Compose
	for _, proj := range projects {
		for _, env := range proj.Environments {
			if environmentID != "" && env.EnvironmentID != environmentID {
				continue
			}
			composes = append(composes, env.Compose...)
		}
	}
	return composes, nil
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
		"databasePassword": password,
		"dockerImage":      dockerImage,
	}

	switch dbType {
	case "postgres":
		endpoint = "postgres.create"
		payload["databaseName"] = name
		payload["databaseUser"] = "postgres"
	case "mysql":
		endpoint = "mysql.create"
		payload["databaseName"] = name
		payload["databaseUser"] = "root"
		payload["databaseRootPassword"] = password // MySQL requires separate root password.
	case "mariadb":
		endpoint = "mariadb.create"
		payload["databaseName"] = name
		payload["databaseUser"] = "root"
		payload["databaseRootPassword"] = password // MariaDB requires separate root password.
	case "mongo":
		// MongoDB does NOT accept databaseName in create API.
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

func (c *DokployClient) UpdateSSHKey(id, name, description string) (*SSHKey, error) {
	payload := map[string]string{
		"sshKeyId":    id,
		"name":        name,
		"description": description,
	}

	_, err := c.doRequest("POST", "sshKey.update", payload)
	if err != nil {
		return nil, err
	}

	// Fetch the updated key to return current state
	return c.GetSSHKey(id)
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
	Command             string `json:"command"`
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

// GitProviderInfo contains the common git provider information nested in responses.
type GitProviderInfo struct {
	GitProviderId  string `json:"gitProviderId"`
	Name           string `json:"name"`
	ProviderType   string `json:"providerType"`
	CreatedAt      string `json:"createdAt"`
	OrganizationID string `json:"organizationId"`
	UserID         string `json:"userId"`
}

type GithubProvider struct {
	ID          string          `json:"githubId"`
	GitProvider GitProviderInfo `json:"gitProvider"`
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

// GetMountsByService fetches all mounts for a service by calling the service-specific endpoint
// and extracting the mounts array from the response.
func (c *DokployClient) GetMountsByService(serviceID, serviceType string) ([]Mount, error) {
	var endpoint string
	switch serviceType {
	case "application":
		endpoint = fmt.Sprintf("application.one?applicationId=%s", serviceID)
	case "postgres":
		endpoint = fmt.Sprintf("postgres.one?postgresId=%s", serviceID)
	case "mysql":
		endpoint = fmt.Sprintf("mysql.one?mysqlId=%s", serviceID)
	case "mariadb":
		endpoint = fmt.Sprintf("mariadb.one?mariadbId=%s", serviceID)
	case "mongo":
		endpoint = fmt.Sprintf("mongo.one?mongoId=%s", serviceID)
	case "redis":
		endpoint = fmt.Sprintf("redis.one?redisId=%s", serviceID)
	case "compose":
		endpoint = fmt.Sprintf("compose.one?composeId=%s", serviceID)
	default:
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response to extract mounts
	var serviceResponse struct {
		Mounts []Mount `json:"mounts"`
	}
	if err := json.Unmarshal(resp, &serviceResponse); err != nil {
		return nil, fmt.Errorf("failed to parse service response: %w", err)
	}

	return serviceResponse.Mounts, nil
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

	// API returns boolean true on success - fetch the created mount from service
	if string(resp) == "true" {
		// Fetch all mounts for the service and find the one we just created
		mounts, err := c.GetMountsByService(mount.ServiceID, mount.ServiceType)
		if err != nil {
			return nil, fmt.Errorf("mount created but failed to fetch mount details: %w", err)
		}

		// Find the mount matching our input (by type, mountPath, and optionally filePath/content)
		// Return the most recently created one that matches
		var bestMatch *Mount
		for i := range mounts {
			m := &mounts[i]
			if m.Type == mount.Type && m.MountPath == mount.MountPath {
				// For file mounts, also check filePath when both sides are non-empty.
				// This allows matching mounts when either the input or returned FilePath is empty.
				if mount.Type == "file" {
					if mount.FilePath != "" && m.FilePath != "" && m.FilePath != mount.FilePath {
						continue
					}
				}
				// For bind mounts, check hostPath
				if mount.Type == "bind" && m.HostPath != mount.HostPath {
					continue
				}
				// For volume mounts, check volumeName
				if mount.Type == "volume" && m.VolumeName != mount.VolumeName {
					continue
				}
				bestMatch = m
			}
		}

		if bestMatch != nil {
			return bestMatch, nil
		}

		return nil, fmt.Errorf("mount created but could not find it in service mounts")
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

	_, err := c.doRequest("POST", "mounts.update", payload)
	if err != nil {
		return nil, err
	}

	// Always fetch fresh data after update since the API returns stale data
	return c.GetMount(mount.ID)
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

// GetPortsByApplication fetches all ports for an application by calling application.one
// and extracting the ports array from the response.
func (c *DokployClient) GetPortsByApplication(applicationID string) ([]Port, error) {
	endpoint := fmt.Sprintf("application.one?applicationId=%s", applicationID)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Parse the application response to extract ports
	var appResponse struct {
		Ports []Port `json:"ports"`
	}
	if err := json.Unmarshal(resp, &appResponse); err != nil {
		return nil, fmt.Errorf("failed to parse application response: %w", err)
	}

	return appResponse.Ports, nil
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

	// API returns boolean true on success - fetch the created port from application
	if string(resp) == "true" {
		// Fetch all ports for the application and find the one we just created
		ports, err := c.GetPortsByApplication(port.ApplicationID)
		if err != nil {
			return nil, fmt.Errorf("port created but failed to fetch port details: %w", err)
		}

		// Find the port matching our input (by publishedPort, targetPort, and protocol if specified)
		for i := range ports {
			p := &ports[i]
			if p.PublishedPort == port.PublishedPort && p.TargetPort == port.TargetPort {
				// If a protocol was specified on creation, also require it to match.
				if port.Protocol != "" && p.Protocol != port.Protocol {
					continue
				}
				return p, nil
			}
		}

		return nil, fmt.Errorf("port created but could not find it in application ports")
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

	_, err := c.doRequest("POST", "port.update", payload)
	if err != nil {
		return nil, err
	}

	// Always fetch fresh data after update since API may return stale data
	return c.GetPort(port.ID)
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

// GetRedirectsByApplication fetches all redirects for an application by calling application.one
// and extracting the redirects array from the response.
func (c *DokployClient) GetRedirectsByApplication(applicationID string) ([]Redirect, error) {
	endpoint := fmt.Sprintf("application.one?applicationId=%s", applicationID)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Parse the application response to extract redirects
	var appResponse struct {
		Redirects []Redirect `json:"redirects"`
	}
	if err := json.Unmarshal(resp, &appResponse); err != nil {
		return nil, fmt.Errorf("failed to parse application response: %w", err)
	}

	return appResponse.Redirects, nil
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

	// API returns boolean true on success - fetch the created redirect from application
	if string(resp) == "true" {
		// Fetch all redirects for the application and find the one we just created
		redirects, err := c.GetRedirectsByApplication(redirect.ApplicationID)
		if err != nil {
			return nil, fmt.Errorf("redirect created but failed to fetch redirect details: %w", err)
		}

		// Find the redirect matching our input (by regex, replacement, permanent)
		// Return the most recently created one that matches
		var bestMatch *Redirect
		for i := range redirects {
			r := &redirects[i]
			if r.Regex == redirect.Regex && r.Replacement == redirect.Replacement && r.Permanent == redirect.Permanent {
				if bestMatch == nil || r.CreatedAt > bestMatch.CreatedAt {
					bestMatch = r
				}
			}
		}

		if bestMatch != nil {
			return bestMatch, nil
		}

		return nil, fmt.Errorf("redirect created but could not find it in application redirects")
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

	// Handle boolean response
	if string(resp) == "true" {
		return c.GetRedirect(redirect.ID)
	}

	var result Redirect
	if err := json.Unmarshal(resp, &result); err != nil {
		// Fallback to fetch
		return c.GetRedirect(redirect.ID)
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

	// Handle boolean response - API returns true on success
	if len(resp) == 0 || string(resp) == "true" {
		return c.GetRegistry(registry.ID)
	}

	var result Registry
	if err := json.Unmarshal(resp, &result); err != nil {
		// If unmarshal fails, try fetching the registry directly
		return c.GetRegistry(registry.ID)
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

// Backup represents a scheduled backup configuration.
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
	// WORKAROUND: Query the database/compose endpoint which includes backups, then find our newly created backup
	if len(resp) == 0 {
		var backups []Backup
		var err error

		if backup.BackupType == "compose" && backup.ComposeID != "" {
			// For compose backups, query the compose endpoint
			backups, err = c.GetBackupsByComposeID(backup.ComposeID)
			if err != nil {
				return nil, fmt.Errorf("backup.create returned empty response, failed to lookup compose backup: %w", err)
			}
		} else {
			// For database backups, query the database endpoint
			var databaseID string
			switch backup.DatabaseType {
			case "postgres":
				databaseID = backup.PostgresID
			case "mysql":
				databaseID = backup.MysqlID
			case "mariadb":
				databaseID = backup.MariadbID
			case "mongo":
				databaseID = backup.MongoID
			}

			if databaseID == "" {
				return nil, fmt.Errorf("backup.create returned empty response and no database ID available to lookup backup")
			}

			backups, err = c.GetBackupsByDatabaseID(databaseID, backup.DatabaseType)
			if err != nil {
				return nil, fmt.Errorf("backup.create returned empty response, failed to lookup backup: %w", err)
			}
		}

		// Find our backup by matching unique parameters
		for _, b := range backups {
			if b.DestinationID == backup.DestinationID &&
				b.Prefix == backup.Prefix &&
				b.Schedule == backup.Schedule {
				return &b, nil
			}
		}

		return nil, fmt.Errorf("backup.create returned empty response and could not find created backup")
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
	// serviceName is required by API schema but can be empty string for database backups.
	// It's only meaningful for compose backups where it specifies the service to backup.
	payload := map[string]interface{}{
		"backupId":      backup.BackupID,
		"schedule":      backup.Schedule,
		"enabled":       backup.Enabled,
		"prefix":        backup.Prefix,
		"destinationId": backup.DestinationID,
		"database":      backup.Database,
		"databaseType":  backup.DatabaseType,
		"serviceName":   backup.ServiceName,
	}

	if backup.KeepLatestCount > 0 {
		payload["keepLatestCount"] = backup.KeepLatestCount
	}

	resp, err := c.doRequest("POST", "backup.update", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response - fetch the backup by ID
	if len(resp) == 0 {
		return c.GetBackup(backup.BackupID)
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

// BackupFile represents a backup file in the destination storage.
type BackupFile struct {
	Key          string `json:"Key"`
	LastModified string `json:"LastModified"`
	Size         int64  `json:"Size"`
	ETag         string `json:"ETag"`
	StorageClass string `json:"StorageClass"`
}

// ListBackupFiles retrieves a list of backup files from a destination.
// search is a required prefix filter for the backup files.
// serverId is optional and filters by server.
func (c *DokployClient) ListBackupFiles(destinationID, search, serverID string) ([]BackupFile, error) {
	endpoint := fmt.Sprintf("backup.listBackupFiles?destinationId=%s&search=%s", destinationID, search)
	if serverID != "" {
		endpoint += fmt.Sprintf("&serverId=%s", serverID)
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []BackupFile
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse backup files response: %w", err)
	}

	return result, nil
}

// GetBackupsByDatabaseID retrieves all backups for a specific database
// by querying the database endpoint which includes backups in its response.
func (c *DokployClient) GetBackupsByDatabaseID(databaseID, databaseType string) ([]Backup, error) {
	var endpoint string
	switch databaseType {
	case "postgres":
		endpoint = fmt.Sprintf("postgres.one?postgresId=%s", databaseID)
	case "mysql":
		endpoint = fmt.Sprintf("mysql.one?mysqlId=%s", databaseID)
	case "mariadb":
		endpoint = fmt.Sprintf("mariadb.one?mariadbId=%s", databaseID)
	case "mongo":
		endpoint = fmt.Sprintf("mongo.one?mongoId=%s", databaseID)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", databaseType)
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// The database response includes a "backups" array
	var result struct {
		Backups []Backup `json:"backups"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse database response: %w", err)
	}

	return result.Backups, nil
}

// GetBackupsByComposeID retrieves all backups for a specific compose
// by querying the compose endpoint which includes backups in its response.
func (c *DokployClient) GetBackupsByComposeID(composeID string) ([]Backup, error) {
	endpoint := fmt.Sprintf("compose.one?composeId=%s", composeID)

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// The compose response includes a "backups" array
	var result struct {
		Backups []Backup `json:"backups"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse compose response: %w", err)
	}

	return result.Backups, nil
}

// CreateServer creates a new remote server.
func (c *DokployClient) CreateServer(server Server) (*Server, error) {
	payload := map[string]interface{}{
		"name":       server.Name,
		"ipAddress":  server.IPAddress,
		"port":       server.Port,
		"username":   server.Username,
		"sshKeyId":   server.SSHKeyID,
		"serverType": server.ServerType,
	}

	if server.Description != "" {
		payload["description"] = server.Description
	}
	// Note: command is NOT accepted by server.create API, only by server.update.

	resp, err := c.doRequest("POST", "server.create", payload)
	if err != nil {
		return nil, err
	}

	var result Server
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server response: %w", err)
	}
	return &result, nil
}

// UpdateServer updates an existing server.
func (c *DokployClient) UpdateServer(server Server) (*Server, error) {
	payload := map[string]interface{}{
		"serverId":    server.ID,
		"name":        server.Name,
		"ipAddress":   server.IPAddress,
		"port":        server.Port,
		"username":    server.Username,
		"sshKeyId":    server.SSHKeyID,
		"serverType":  server.ServerType,
		"description": server.Description,
		"command":     server.Command,
	}

	resp, err := c.doRequest("POST", "server.update", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response.
	if len(resp) == 0 {
		return c.GetServer(server.ID)
	}

	var result Server
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteServer removes a server by ID.
func (c *DokployClient) DeleteServer(id string) error {
	payload := map[string]string{
		"serverId": id,
	}
	_, err := c.doRequest("POST", "server.remove", payload)
	return err
}

// --- Postgres ---

// Postgres represents a PostgreSQL database instance.
type Postgres struct {
	PostgresID        string `json:"postgresId"`
	Name              string `json:"name"`
	AppName           string `json:"appName"`
	Description       string `json:"description"`
	DatabaseName      string `json:"databaseName"`
	DatabaseUser      string `json:"databaseUser"`
	DatabasePassword  string `json:"databasePassword"`
	DockerImage       string `json:"dockerImage"`
	Command           string `json:"command"`
	Env               string `json:"env"`
	MemoryReservation string `json:"memoryReservation"`
	MemoryLimit       string `json:"memoryLimit"`
	CPUReservation    string `json:"cpuReservation"`
	CPULimit          string `json:"cpuLimit"`
	ExternalPort      int    `json:"externalPort"`
	EnvironmentID     string `json:"environmentId"`
	ApplicationStatus string `json:"applicationStatus"`
	Replicas          int    `json:"replicas"`
	ServerID          string `json:"serverId"`
}

// CreatePostgres creates a new PostgreSQL database instance.
func (c *DokployClient) CreatePostgres(postgres Postgres) (*Postgres, error) {
	payload := map[string]interface{}{
		"name":             postgres.Name,
		"appName":          postgres.AppName,
		"databaseName":     postgres.DatabaseName,
		"databaseUser":     postgres.DatabaseUser,
		"databasePassword": postgres.DatabasePassword,
		"environmentId":    postgres.EnvironmentID,
	}

	if postgres.DockerImage != "" {
		payload["dockerImage"] = postgres.DockerImage
	}
	if postgres.Description != "" {
		payload["description"] = postgres.Description
	}
	if postgres.ServerID != "" {
		payload["serverId"] = postgres.ServerID
	}

	resp, err := c.doRequest("POST", "postgres.create", payload)
	if err != nil {
		return nil, err
	}

	var result Postgres
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal postgres response: %w", err)
	}
	return &result, nil
}

// GetPostgres retrieves a PostgreSQL instance by ID.
func (c *DokployClient) GetPostgres(id string) (*Postgres, error) {
	endpoint := fmt.Sprintf("postgres.one?postgresId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Postgres
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePostgres updates an existing PostgreSQL instance.
func (c *DokployClient) UpdatePostgres(postgres Postgres) (*Postgres, error) {
	payload := map[string]interface{}{
		"postgresId": postgres.PostgresID,
	}

	if postgres.Name != "" {
		payload["name"] = postgres.Name
	}
	if postgres.AppName != "" {
		payload["appName"] = postgres.AppName
	}
	if postgres.Description != "" {
		payload["description"] = postgres.Description
	}
	if postgres.DatabaseName != "" {
		payload["databaseName"] = postgres.DatabaseName
	}
	if postgres.DatabaseUser != "" {
		payload["databaseUser"] = postgres.DatabaseUser
	}
	if postgres.DatabasePassword != "" {
		payload["databasePassword"] = postgres.DatabasePassword
	}
	if postgres.DockerImage != "" {
		payload["dockerImage"] = postgres.DockerImage
	}
	if postgres.Command != "" {
		payload["command"] = postgres.Command
	}
	if postgres.Env != "" {
		payload["env"] = postgres.Env
	}
	if postgres.MemoryReservation != "" {
		payload["memoryReservation"] = postgres.MemoryReservation
	}
	if postgres.MemoryLimit != "" {
		payload["memoryLimit"] = postgres.MemoryLimit
	}
	if postgres.CPUReservation != "" {
		payload["cpuReservation"] = postgres.CPUReservation
	}
	if postgres.CPULimit != "" {
		payload["cpuLimit"] = postgres.CPULimit
	}
	if postgres.ExternalPort > 0 {
		payload["externalPort"] = postgres.ExternalPort
	}
	if postgres.Replicas > 0 {
		payload["replicas"] = postgres.Replicas
	}

	resp, err := c.doRequest("POST", "postgres.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return c.GetPostgres(postgres.PostgresID)
	}

	var result Postgres
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetPostgres(postgres.PostgresID)
	}
	return &result, nil
}

// DeletePostgres removes a PostgreSQL instance by ID.
func (c *DokployClient) DeletePostgres(id string) error {
	payload := map[string]string{
		"postgresId": id,
	}
	_, err := c.doRequest("POST", "postgres.remove", payload)
	return err
}

// --- MySQL ---

// MySQL represents a MySQL database instance.
type MySQL struct {
	MySQLID              string `json:"mysqlId"`
	Name                 string `json:"name"`
	AppName              string `json:"appName"`
	Description          string `json:"description"`
	DatabaseName         string `json:"databaseName"`
	DatabaseUser         string `json:"databaseUser"`
	DatabasePassword     string `json:"databasePassword"`
	DatabaseRootPassword string `json:"databaseRootPassword"`
	DockerImage          string `json:"dockerImage"`
	Command              string `json:"command"`
	Env                  string `json:"env"`
	MemoryReservation    string `json:"memoryReservation"`
	MemoryLimit          string `json:"memoryLimit"`
	CPUReservation       string `json:"cpuReservation"`
	CPULimit             string `json:"cpuLimit"`
	ExternalPort         int    `json:"externalPort"`
	EnvironmentID        string `json:"environmentId"`
	ApplicationStatus    string `json:"applicationStatus"`
	Replicas             int    `json:"replicas"`
	ServerID             string `json:"serverId"`
}

// CreateMySQL creates a new MySQL database instance.
func (c *DokployClient) CreateMySQL(mysql MySQL) (*MySQL, error) {
	payload := map[string]interface{}{
		"name":                 mysql.Name,
		"appName":              mysql.AppName,
		"databaseName":         mysql.DatabaseName,
		"databaseUser":         mysql.DatabaseUser,
		"databasePassword":     mysql.DatabasePassword,
		"databaseRootPassword": mysql.DatabaseRootPassword,
		"environmentId":        mysql.EnvironmentID,
	}

	if mysql.DockerImage != "" {
		payload["dockerImage"] = mysql.DockerImage
	}
	if mysql.Description != "" {
		payload["description"] = mysql.Description
	}
	if mysql.ServerID != "" {
		payload["serverId"] = mysql.ServerID
	}

	resp, err := c.doRequest("POST", "mysql.create", payload)
	if err != nil {
		return nil, err
	}

	var result MySQL
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mysql response: %w", err)
	}
	return &result, nil
}

// GetMySQL retrieves a MySQL instance by ID.
func (c *DokployClient) GetMySQL(id string) (*MySQL, error) {
	endpoint := fmt.Sprintf("mysql.one?mysqlId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MySQL
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMySQL updates an existing MySQL instance.
func (c *DokployClient) UpdateMySQL(mysql MySQL) (*MySQL, error) {
	payload := map[string]interface{}{
		"mysqlId": mysql.MySQLID,
	}

	if mysql.Name != "" {
		payload["name"] = mysql.Name
	}
	if mysql.AppName != "" {
		payload["appName"] = mysql.AppName
	}
	if mysql.Description != "" {
		payload["description"] = mysql.Description
	}
	if mysql.DatabaseName != "" {
		payload["databaseName"] = mysql.DatabaseName
	}
	if mysql.DatabaseUser != "" {
		payload["databaseUser"] = mysql.DatabaseUser
	}
	if mysql.DatabasePassword != "" {
		payload["databasePassword"] = mysql.DatabasePassword
	}
	if mysql.DatabaseRootPassword != "" {
		payload["databaseRootPassword"] = mysql.DatabaseRootPassword
	}
	if mysql.DockerImage != "" {
		payload["dockerImage"] = mysql.DockerImage
	}
	if mysql.Command != "" {
		payload["command"] = mysql.Command
	}
	if mysql.Env != "" {
		payload["env"] = mysql.Env
	}
	if mysql.MemoryReservation != "" {
		payload["memoryReservation"] = mysql.MemoryReservation
	}
	if mysql.MemoryLimit != "" {
		payload["memoryLimit"] = mysql.MemoryLimit
	}
	if mysql.CPUReservation != "" {
		payload["cpuReservation"] = mysql.CPUReservation
	}
	if mysql.CPULimit != "" {
		payload["cpuLimit"] = mysql.CPULimit
	}
	if mysql.ExternalPort > 0 {
		payload["externalPort"] = mysql.ExternalPort
	}
	if mysql.Replicas > 0 {
		payload["replicas"] = mysql.Replicas
	}

	resp, err := c.doRequest("POST", "mysql.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return c.GetMySQL(mysql.MySQLID)
	}

	var result MySQL
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetMySQL(mysql.MySQLID)
	}
	return &result, nil
}

// DeleteMySQL removes a MySQL instance by ID.
func (c *DokployClient) DeleteMySQL(id string) error {
	payload := map[string]string{
		"mysqlId": id,
	}
	_, err := c.doRequest("POST", "mysql.remove", payload)
	return err
}

// --- MariaDB ---

// MariaDB represents a MariaDB database instance.
type MariaDB struct {
	MariaDBID            string `json:"mariadbId"`
	Name                 string `json:"name"`
	AppName              string `json:"appName"`
	Description          string `json:"description"`
	DatabaseName         string `json:"databaseName"`
	DatabaseUser         string `json:"databaseUser"`
	DatabasePassword     string `json:"databasePassword"`
	DatabaseRootPassword string `json:"databaseRootPassword"`
	DockerImage          string `json:"dockerImage"`
	Command              string `json:"command"`
	Env                  string `json:"env"`
	MemoryReservation    string `json:"memoryReservation"`
	MemoryLimit          string `json:"memoryLimit"`
	CPUReservation       string `json:"cpuReservation"`
	CPULimit             string `json:"cpuLimit"`
	ExternalPort         int    `json:"externalPort"`
	EnvironmentID        string `json:"environmentId"`
	ApplicationStatus    string `json:"applicationStatus"`
	Replicas             int    `json:"replicas"`
	ServerID             string `json:"serverId"`
}

// CreateMariaDB creates a new MariaDB database instance.
func (c *DokployClient) CreateMariaDB(mariadb MariaDB) (*MariaDB, error) {
	payload := map[string]interface{}{
		"name":                 mariadb.Name,
		"appName":              mariadb.AppName,
		"databaseName":         mariadb.DatabaseName,
		"databaseUser":         mariadb.DatabaseUser,
		"databasePassword":     mariadb.DatabasePassword,
		"databaseRootPassword": mariadb.DatabaseRootPassword,
		"environmentId":        mariadb.EnvironmentID,
	}

	if mariadb.DockerImage != "" {
		payload["dockerImage"] = mariadb.DockerImage
	}
	if mariadb.Description != "" {
		payload["description"] = mariadb.Description
	}
	if mariadb.ServerID != "" {
		payload["serverId"] = mariadb.ServerID
	}

	resp, err := c.doRequest("POST", "mariadb.create", payload)
	if err != nil {
		return nil, err
	}

	var result MariaDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mariadb response: %w", err)
	}
	return &result, nil
}

// GetMariaDB retrieves a MariaDB instance by ID.
func (c *DokployClient) GetMariaDB(id string) (*MariaDB, error) {
	endpoint := fmt.Sprintf("mariadb.one?mariadbId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MariaDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMariaDB updates an existing MariaDB instance.
func (c *DokployClient) UpdateMariaDB(mariadb MariaDB) (*MariaDB, error) {
	payload := map[string]interface{}{
		"mariadbId": mariadb.MariaDBID,
	}

	if mariadb.Name != "" {
		payload["name"] = mariadb.Name
	}
	if mariadb.AppName != "" {
		payload["appName"] = mariadb.AppName
	}
	if mariadb.Description != "" {
		payload["description"] = mariadb.Description
	}
	if mariadb.DatabaseName != "" {
		payload["databaseName"] = mariadb.DatabaseName
	}
	if mariadb.DatabaseUser != "" {
		payload["databaseUser"] = mariadb.DatabaseUser
	}
	if mariadb.DatabasePassword != "" {
		payload["databasePassword"] = mariadb.DatabasePassword
	}
	if mariadb.DatabaseRootPassword != "" {
		payload["databaseRootPassword"] = mariadb.DatabaseRootPassword
	}
	if mariadb.DockerImage != "" {
		payload["dockerImage"] = mariadb.DockerImage
	}
	if mariadb.Command != "" {
		payload["command"] = mariadb.Command
	}
	if mariadb.Env != "" {
		payload["env"] = mariadb.Env
	}
	if mariadb.MemoryReservation != "" {
		payload["memoryReservation"] = mariadb.MemoryReservation
	}
	if mariadb.MemoryLimit != "" {
		payload["memoryLimit"] = mariadb.MemoryLimit
	}
	if mariadb.CPUReservation != "" {
		payload["cpuReservation"] = mariadb.CPUReservation
	}
	if mariadb.CPULimit != "" {
		payload["cpuLimit"] = mariadb.CPULimit
	}
	if mariadb.ExternalPort > 0 {
		payload["externalPort"] = mariadb.ExternalPort
	}
	if mariadb.Replicas > 0 {
		payload["replicas"] = mariadb.Replicas
	}

	resp, err := c.doRequest("POST", "mariadb.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return c.GetMariaDB(mariadb.MariaDBID)
	}

	var result MariaDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetMariaDB(mariadb.MariaDBID)
	}
	return &result, nil
}

// DeleteMariaDB removes a MariaDB instance by ID.
func (c *DokployClient) DeleteMariaDB(id string) error {
	payload := map[string]string{
		"mariadbId": id,
	}
	_, err := c.doRequest("POST", "mariadb.remove", payload)
	return err
}

// --- MongoDB ---

// MongoDB represents a MongoDB database instance.
type MongoDB struct {
	MongoID           string `json:"mongoId"`
	Name              string `json:"name"`
	AppName           string `json:"appName"`
	Description       string `json:"description"`
	DatabaseUser      string `json:"databaseUser"`
	DatabasePassword  string `json:"databasePassword"`
	ReplicaSets       bool   `json:"replicaSets"`
	DockerImage       string `json:"dockerImage"`
	Command           string `json:"command"`
	Env               string `json:"env"`
	MemoryReservation string `json:"memoryReservation"`
	MemoryLimit       string `json:"memoryLimit"`
	CPUReservation    string `json:"cpuReservation"`
	CPULimit          string `json:"cpuLimit"`
	ExternalPort      int    `json:"externalPort"`
	EnvironmentID     string `json:"environmentId"`
	ApplicationStatus string `json:"applicationStatus"`
	Replicas          int    `json:"replicas"`
	ServerID          string `json:"serverId"`
}

// CreateMongoDB creates a new MongoDB database instance.
func (c *DokployClient) CreateMongoDB(mongo MongoDB) (*MongoDB, error) {
	payload := map[string]interface{}{
		"name":             mongo.Name,
		"appName":          mongo.AppName,
		"databaseUser":     mongo.DatabaseUser,
		"databasePassword": mongo.DatabasePassword,
		"environmentId":    mongo.EnvironmentID,
	}

	if mongo.DockerImage != "" {
		payload["dockerImage"] = mongo.DockerImage
	}
	if mongo.Description != "" {
		payload["description"] = mongo.Description
	}
	if mongo.ServerID != "" {
		payload["serverId"] = mongo.ServerID
	}
	// ReplicaSets defaults to false, only include if true
	if mongo.ReplicaSets {
		payload["replicaSets"] = mongo.ReplicaSets
	}

	resp, err := c.doRequest("POST", "mongo.create", payload)
	if err != nil {
		return nil, err
	}

	var result MongoDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mongo response: %w", err)
	}
	return &result, nil
}

// GetMongoDB retrieves a MongoDB instance by ID.
func (c *DokployClient) GetMongoDB(id string) (*MongoDB, error) {
	endpoint := fmt.Sprintf("mongo.one?mongoId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MongoDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMongoDB updates an existing MongoDB instance.
func (c *DokployClient) UpdateMongoDB(mongo MongoDB) (*MongoDB, error) {
	payload := map[string]interface{}{
		"mongoId": mongo.MongoID,
	}

	if mongo.Name != "" {
		payload["name"] = mongo.Name
	}
	if mongo.AppName != "" {
		payload["appName"] = mongo.AppName
	}
	if mongo.Description != "" {
		payload["description"] = mongo.Description
	}
	if mongo.DatabaseUser != "" {
		payload["databaseUser"] = mongo.DatabaseUser
	}
	if mongo.DatabasePassword != "" {
		payload["databasePassword"] = mongo.DatabasePassword
	}
	// Always include replicaSets in update since it's a boolean
	payload["replicaSets"] = mongo.ReplicaSets
	if mongo.DockerImage != "" {
		payload["dockerImage"] = mongo.DockerImage
	}
	if mongo.Command != "" {
		payload["command"] = mongo.Command
	}
	if mongo.Env != "" {
		payload["env"] = mongo.Env
	}
	if mongo.MemoryReservation != "" {
		payload["memoryReservation"] = mongo.MemoryReservation
	}
	if mongo.MemoryLimit != "" {
		payload["memoryLimit"] = mongo.MemoryLimit
	}
	if mongo.CPUReservation != "" {
		payload["cpuReservation"] = mongo.CPUReservation
	}
	if mongo.CPULimit != "" {
		payload["cpuLimit"] = mongo.CPULimit
	}
	if mongo.ExternalPort > 0 {
		payload["externalPort"] = mongo.ExternalPort
	}
	if mongo.Replicas > 0 {
		payload["replicas"] = mongo.Replicas
	}

	resp, err := c.doRequest("POST", "mongo.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return c.GetMongoDB(mongo.MongoID)
	}

	var result MongoDB
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetMongoDB(mongo.MongoID)
	}
	return &result, nil
}

// DeleteMongoDB removes a MongoDB instance by ID.
func (c *DokployClient) DeleteMongoDB(id string) error {
	payload := map[string]string{
		"mongoId": id,
	}
	_, err := c.doRequest("POST", "mongo.remove", payload)
	return err
}

// --- Redis ---

// Redis represents a Redis database instance.
type Redis struct {
	RedisID           string `json:"redisId"`
	Name              string `json:"name"`
	AppName           string `json:"appName"`
	Description       string `json:"description"`
	DatabasePassword  string `json:"databasePassword"`
	DockerImage       string `json:"dockerImage"`
	Command           string `json:"command"`
	Env               string `json:"env"`
	MemoryReservation string `json:"memoryReservation"`
	MemoryLimit       string `json:"memoryLimit"`
	CPUReservation    string `json:"cpuReservation"`
	CPULimit          string `json:"cpuLimit"`
	ExternalPort      int    `json:"externalPort"`
	EnvironmentID     string `json:"environmentId"`
	ApplicationStatus string `json:"applicationStatus"`
	Replicas          int    `json:"replicas"`
	ServerID          string `json:"serverId"`
}

// CreateRedis creates a new Redis database instance.
// Note: The redis.create API only accepts: name, appName, databasePassword,
// dockerImage, environmentId, description, serverId. Other fields like
// command, env, memoryReservation, memoryLimit, cpuReservation, cpuLimit,
// externalPort, and replicas must be set via redis.update after creation.
func (c *DokployClient) CreateRedis(redis Redis) (*Redis, error) {
	payload := map[string]interface{}{
		"name":             redis.Name,
		"appName":          redis.AppName,
		"databasePassword": redis.DatabasePassword,
		"environmentId":    redis.EnvironmentID,
	}

	// Include optional fields accepted by the create API.
	if redis.DockerImage != "" {
		payload["dockerImage"] = redis.DockerImage
	}
	if redis.Description != "" {
		payload["description"] = redis.Description
	}
	if redis.ServerID != "" {
		payload["serverId"] = redis.ServerID
	}

	resp, err := c.doRequest("POST", "redis.create", payload)
	if err != nil {
		return nil, err
	}

	var result Redis
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redis response: %w", err)
	}
	return &result, nil
}

// GetRedis retrieves a Redis instance by ID.
func (c *DokployClient) GetRedis(id string) (*Redis, error) {
	endpoint := fmt.Sprintf("redis.one?redisId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Redis
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRedis updates an existing Redis instance.
func (c *DokployClient) UpdateRedis(redis Redis) (*Redis, error) {
	payload := map[string]interface{}{
		"redisId": redis.RedisID,
	}

	if redis.Name != "" {
		payload["name"] = redis.Name
	}
	if redis.AppName != "" {
		payload["appName"] = redis.AppName
	}
	if redis.Description != "" {
		payload["description"] = redis.Description
	}
	if redis.DatabasePassword != "" {
		payload["databasePassword"] = redis.DatabasePassword
	}
	if redis.DockerImage != "" {
		payload["dockerImage"] = redis.DockerImage
	}
	if redis.Command != "" {
		payload["command"] = redis.Command
	}
	if redis.Env != "" {
		payload["env"] = redis.Env
	}
	if redis.MemoryReservation != "" {
		payload["memoryReservation"] = redis.MemoryReservation
	}
	if redis.MemoryLimit != "" {
		payload["memoryLimit"] = redis.MemoryLimit
	}
	if redis.CPUReservation != "" {
		payload["cpuReservation"] = redis.CPUReservation
	}
	if redis.CPULimit != "" {
		payload["cpuLimit"] = redis.CPULimit
	}
	if redis.ExternalPort > 0 {
		payload["externalPort"] = redis.ExternalPort
	}
	if redis.Replicas > 0 {
		payload["replicas"] = redis.Replicas
	}

	resp, err := c.doRequest("POST", "redis.update", payload)
	if err != nil {
		return nil, err
	}

	// Handle empty response or non-JSON response (API may return boolean).
	if len(resp) == 0 {
		return c.GetRedis(redis.RedisID)
	}

	var result Redis
	if err := json.Unmarshal(resp, &result); err != nil {
		// API might return a boolean or other non-object response.
		return c.GetRedis(redis.RedisID)
	}
	return &result, nil
}

// DeleteRedis removes a Redis instance by ID.
func (c *DokployClient) DeleteRedis(id string) error {
	payload := map[string]string{
		"redisId": id,
	}
	_, err := c.doRequest("POST", "redis.remove", payload)
	return err
}

// --- GitLab Provider ---

// GitlabProviderListItem is the structure returned by the gitlabProviders list endpoint.
type GitlabProviderListItem struct {
	ID          string          `json:"gitlabId"`
	GitProvider GitProviderInfo `json:"gitProvider"`
	GitlabUrl   string          `json:"gitlabUrl"`
}

// GitlabProvider is the full structure used for create/update operations.
type GitlabProvider struct {
	ID             string `json:"gitlabId"`
	GitProviderId  string `json:"gitProviderId"`
	Name           string `json:"name"`
	GitlabUrl      string `json:"gitlabUrl"`
	ApplicationId  string `json:"applicationId"`
	RedirectUri    string `json:"redirectUri"`
	Secret         string `json:"secret"`
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	GroupName      string `json:"groupName"`
	ExpiresAt      int64  `json:"expiresAt"`
	AuthId         string `json:"authId"`
	OrganizationID string `json:"organizationId"`
	CreatedAt      string `json:"createdAt"`
}

func (c *DokployClient) CreateGitlabProvider(provider GitlabProvider) (*GitlabProvider, error) {
	payload := map[string]interface{}{
		"name":      provider.Name,
		"gitlabUrl": provider.GitlabUrl,
		"authId":    provider.AuthId,
	}

	if provider.ApplicationId != "" {
		payload["applicationId"] = provider.ApplicationId
	}
	if provider.RedirectUri != "" {
		payload["redirectUri"] = provider.RedirectUri
	}
	if provider.Secret != "" {
		payload["secret"] = provider.Secret
	}
	if provider.AccessToken != "" {
		payload["accessToken"] = provider.AccessToken
	}
	if provider.RefreshToken != "" {
		payload["refreshToken"] = provider.RefreshToken
	}
	if provider.GroupName != "" {
		payload["groupName"] = provider.GroupName
	}
	if provider.ExpiresAt != 0 {
		payload["expiresAt"] = provider.ExpiresAt
	}

	resp, err := c.doRequest("POST", "gitlab.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal the response
	var result GitlabProvider
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	// Try wrapper format
	var wrapper struct {
		GitlabProvider GitlabProvider `json:"gitlab"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.GitlabProvider.ID != "" {
		return &wrapper.GitlabProvider, nil
	}

	// If we got here, try to find by name
	return c.findGitlabProviderByName(provider.Name)
}

func (c *DokployClient) findGitlabProviderByName(name string) (*GitlabProvider, error) {
	providers, err := c.ListGitlabProviders()
	if err != nil {
		return nil, fmt.Errorf("gitlab provider created but failed to list providers: %w", err)
	}
	for _, p := range providers {
		if p.GitProvider.Name == name {
			// Fetch the full provider details
			return c.GetGitlabProvider(p.ID)
		}
	}
	return nil, fmt.Errorf("gitlab provider created but not found in list by name: %s", name)
}

func (c *DokployClient) GetGitlabProvider(id string) (*GitlabProvider, error) {
	endpoint := fmt.Sprintf("gitlab.one?gitlabId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GitlabProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateGitlabProvider(provider GitlabProvider) (*GitlabProvider, error) {
	payload := map[string]interface{}{
		"gitlabId": provider.ID,
		"name":     provider.Name,
	}

	if provider.GitlabUrl != "" {
		payload["gitlabUrl"] = provider.GitlabUrl
	}
	if provider.ApplicationId != "" {
		payload["applicationId"] = provider.ApplicationId
	}
	if provider.RedirectUri != "" {
		payload["redirectUri"] = provider.RedirectUri
	}
	if provider.Secret != "" {
		payload["secret"] = provider.Secret
	}
	if provider.AccessToken != "" {
		payload["accessToken"] = provider.AccessToken
	}
	if provider.RefreshToken != "" {
		payload["refreshToken"] = provider.RefreshToken
	}
	if provider.GroupName != "" {
		payload["groupName"] = provider.GroupName
	}
	if provider.ExpiresAt != 0 {
		payload["expiresAt"] = provider.ExpiresAt
	}
	if provider.GitProviderId != "" {
		payload["gitProviderId"] = provider.GitProviderId
	}
	if provider.AuthId != "" {
		payload["authId"] = provider.AuthId
	}

	resp, err := c.doRequest("POST", "gitlab.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 || string(resp) == "true" {
		return c.GetGitlabProvider(provider.ID)
	}

	var result GitlabProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetGitlabProvider(provider.ID)
	}
	return &result, nil
}

func (c *DokployClient) DeleteGitProvider(gitProviderId string) error {
	payload := map[string]string{
		"gitProviderId": gitProviderId,
	}
	_, err := c.doRequest("POST", "gitProvider.remove", payload)
	return err
}

func (c *DokployClient) ListGitlabProviders() ([]GitlabProviderListItem, error) {
	resp, err := c.doRequest("GET", "gitlab.gitlabProviders", nil)
	if err != nil {
		return nil, err
	}

	// Try direct array response
	var providers []GitlabProviderListItem
	if err := json.Unmarshal(resp, &providers); err == nil {
		return providers, nil
	}

	// Try wrapper format
	var wrapper struct {
		Providers []GitlabProviderListItem `json:"providers"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Providers, nil
	}

	// Try gitlabProviders key
	var wrapper2 struct {
		Providers []GitlabProviderListItem `json:"gitlabProviders"`
	}
	if err := json.Unmarshal(resp, &wrapper2); err == nil {
		return wrapper2.Providers, nil
	}

	return nil, fmt.Errorf("failed to parse gitlab providers response")
}

// --- Bitbucket Provider ---

// BitbucketProviderListItem is the structure returned by the bitbucketProviders list endpoint.
type BitbucketProviderListItem struct {
	ID          string          `json:"bitbucketId"`
	GitProvider GitProviderInfo `json:"gitProvider"`
}

// BitbucketProvider is the full structure used for create/update operations.
type BitbucketProvider struct {
	ID                     string `json:"bitbucketId"`
	GitProviderId          string `json:"gitProviderId"`
	Name                   string `json:"name"`
	BitbucketUsername      string `json:"bitbucketUsername"`
	BitbucketEmail         string `json:"bitbucketEmail"`
	AppPassword            string `json:"appPassword"`
	ApiToken               string `json:"apiToken"`
	BitbucketWorkspaceName string `json:"bitbucketWorkspaceName"`
	AuthId                 string `json:"authId"`
	OrganizationID         string `json:"organizationId"`
	CreatedAt              string `json:"createdAt"`
}

func (c *DokployClient) CreateBitbucketProvider(provider BitbucketProvider) (*BitbucketProvider, error) {
	payload := map[string]interface{}{
		"name":   provider.Name,
		"authId": provider.AuthId,
	}

	if provider.BitbucketUsername != "" {
		payload["bitbucketUsername"] = provider.BitbucketUsername
	}
	if provider.BitbucketEmail != "" {
		payload["bitbucketEmail"] = provider.BitbucketEmail
	}
	if provider.AppPassword != "" {
		payload["appPassword"] = provider.AppPassword
	}
	if provider.ApiToken != "" {
		payload["apiToken"] = provider.ApiToken
	}
	if provider.BitbucketWorkspaceName != "" {
		payload["bitbucketWorkspaceName"] = provider.BitbucketWorkspaceName
	}

	resp, err := c.doRequest("POST", "bitbucket.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal the response
	var result BitbucketProvider
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	// Try wrapper format
	var wrapper struct {
		BitbucketProvider BitbucketProvider `json:"bitbucket"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.BitbucketProvider.ID != "" {
		return &wrapper.BitbucketProvider, nil
	}

	// If we got here, try to find by name
	return c.findBitbucketProviderByName(provider.Name)
}

func (c *DokployClient) findBitbucketProviderByName(name string) (*BitbucketProvider, error) {
	providers, err := c.ListBitbucketProviders()
	if err != nil {
		return nil, fmt.Errorf("bitbucket provider created but failed to list providers: %w", err)
	}
	for _, p := range providers {
		if p.GitProvider.Name == name {
			// Fetch the full provider details
			return c.GetBitbucketProvider(p.ID)
		}
	}
	return nil, fmt.Errorf("bitbucket provider created but not found in list by name: %s", name)
}

func (c *DokployClient) GetBitbucketProvider(id string) (*BitbucketProvider, error) {
	endpoint := fmt.Sprintf("bitbucket.one?bitbucketId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result BitbucketProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateBitbucketProvider(provider BitbucketProvider) (*BitbucketProvider, error) {
	payload := map[string]interface{}{
		"bitbucketId":   provider.ID,
		"name":          provider.Name,
		"gitProviderId": provider.GitProviderId,
	}

	if provider.BitbucketUsername != "" {
		payload["bitbucketUsername"] = provider.BitbucketUsername
	}
	if provider.BitbucketEmail != "" {
		payload["bitbucketEmail"] = provider.BitbucketEmail
	}
	if provider.AppPassword != "" {
		payload["appPassword"] = provider.AppPassword
	}
	if provider.ApiToken != "" {
		payload["apiToken"] = provider.ApiToken
	}
	if provider.BitbucketWorkspaceName != "" {
		payload["bitbucketWorkspaceName"] = provider.BitbucketWorkspaceName
	}
	if provider.AuthId != "" {
		payload["authId"] = provider.AuthId
	}

	resp, err := c.doRequest("POST", "bitbucket.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 || string(resp) == "true" {
		return c.GetBitbucketProvider(provider.ID)
	}

	var result BitbucketProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetBitbucketProvider(provider.ID)
	}
	return &result, nil
}

func (c *DokployClient) ListBitbucketProviders() ([]BitbucketProviderListItem, error) {
	resp, err := c.doRequest("GET", "bitbucket.bitbucketProviders", nil)
	if err != nil {
		return nil, err
	}

	// Try direct array response
	var providers []BitbucketProviderListItem
	if err := json.Unmarshal(resp, &providers); err == nil {
		return providers, nil
	}

	// Try wrapper format
	var wrapper struct {
		Providers []BitbucketProviderListItem `json:"providers"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Providers, nil
	}

	// Try bitbucketProviders key
	var wrapper2 struct {
		Providers []BitbucketProviderListItem `json:"bitbucketProviders"`
	}
	if err := json.Unmarshal(resp, &wrapper2); err == nil {
		return wrapper2.Providers, nil
	}

	return nil, fmt.Errorf("failed to parse bitbucket providers response")
}

// --- Gitea Provider ---

// GiteaProviderListItem is the structure returned by the giteaProviders list endpoint.
type GiteaProviderListItem struct {
	ID          string          `json:"giteaId"`
	GitProvider GitProviderInfo `json:"gitProvider"`
}

// GiteaProvider is the full structure used for create/update operations.
type GiteaProvider struct {
	ID                  string `json:"giteaId"`
	GitProviderId       string `json:"gitProviderId"`
	Name                string `json:"name"`
	GiteaUrl            string `json:"giteaUrl"`
	RedirectUri         string `json:"redirectUri"`
	ClientId            string `json:"clientId"`
	ClientSecret        string `json:"clientSecret"`
	AccessToken         string `json:"accessToken"`
	RefreshToken        string `json:"refreshToken"`
	ExpiresAt           int64  `json:"expiresAt"`
	Scopes              string `json:"scopes"`
	LastAuthenticatedAt int64  `json:"lastAuthenticatedAt"`
	GiteaUsername       string `json:"giteaUsername"`
	OrganizationName    string `json:"organizationName"`
	OrganizationID      string `json:"organizationId"`
	CreatedAt           string `json:"createdAt"`
}

func (c *DokployClient) CreateGiteaProvider(provider GiteaProvider) (*GiteaProvider, error) {
	payload := map[string]interface{}{
		"name":     provider.Name,
		"giteaUrl": provider.GiteaUrl,
	}

	if provider.RedirectUri != "" {
		payload["redirectUri"] = provider.RedirectUri
	}
	if provider.ClientId != "" {
		payload["clientId"] = provider.ClientId
	}
	if provider.ClientSecret != "" {
		payload["clientSecret"] = provider.ClientSecret
	}
	if provider.AccessToken != "" {
		payload["accessToken"] = provider.AccessToken
	}
	if provider.RefreshToken != "" {
		payload["refreshToken"] = provider.RefreshToken
	}
	if provider.ExpiresAt != 0 {
		payload["expiresAt"] = provider.ExpiresAt
	}
	if provider.Scopes != "" {
		payload["scopes"] = provider.Scopes
	}
	if provider.LastAuthenticatedAt != 0 {
		payload["lastAuthenticatedAt"] = provider.LastAuthenticatedAt
	}
	if provider.GiteaUsername != "" {
		payload["giteaUsername"] = provider.GiteaUsername
	}
	if provider.OrganizationName != "" {
		payload["organizationName"] = provider.OrganizationName
	}

	resp, err := c.doRequest("POST", "gitea.create", payload)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal the response
	var result GiteaProvider
	if err := json.Unmarshal(resp, &result); err == nil && result.ID != "" {
		return &result, nil
	}

	// Try wrapper format
	var wrapper struct {
		GiteaProvider GiteaProvider `json:"gitea"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil && wrapper.GiteaProvider.ID != "" {
		return &wrapper.GiteaProvider, nil
	}

	// If we got here, try to find by name
	return c.findGiteaProviderByName(provider.Name)
}

func (c *DokployClient) findGiteaProviderByName(name string) (*GiteaProvider, error) {
	providers, err := c.ListGiteaProviders()
	if err != nil {
		return nil, fmt.Errorf("gitea provider created but failed to list providers: %w", err)
	}
	for _, p := range providers {
		if p.GitProvider.Name == name {
			// Fetch the full provider details
			return c.GetGiteaProvider(p.ID)
		}
	}
	return nil, fmt.Errorf("gitea provider created but not found in list by name: %s", name)
}

func (c *DokployClient) GetGiteaProvider(id string) (*GiteaProvider, error) {
	endpoint := fmt.Sprintf("gitea.one?giteaId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GiteaProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateGiteaProvider(provider GiteaProvider) (*GiteaProvider, error) {
	payload := map[string]interface{}{
		"giteaId": provider.ID,
		"name":    provider.Name,
	}

	if provider.GiteaUrl != "" {
		payload["giteaUrl"] = provider.GiteaUrl
	}
	if provider.RedirectUri != "" {
		payload["redirectUri"] = provider.RedirectUri
	}
	if provider.ClientId != "" {
		payload["clientId"] = provider.ClientId
	}
	if provider.ClientSecret != "" {
		payload["clientSecret"] = provider.ClientSecret
	}
	if provider.AccessToken != "" {
		payload["accessToken"] = provider.AccessToken
	}
	if provider.RefreshToken != "" {
		payload["refreshToken"] = provider.RefreshToken
	}
	if provider.ExpiresAt != 0 {
		payload["expiresAt"] = provider.ExpiresAt
	}
	if provider.Scopes != "" {
		payload["scopes"] = provider.Scopes
	}
	if provider.LastAuthenticatedAt != 0 {
		payload["lastAuthenticatedAt"] = provider.LastAuthenticatedAt
	}
	if provider.GiteaUsername != "" {
		payload["giteaUsername"] = provider.GiteaUsername
	}
	if provider.OrganizationName != "" {
		payload["organizationName"] = provider.OrganizationName
	}
	if provider.GitProviderId != "" {
		payload["gitProviderId"] = provider.GitProviderId
	}

	resp, err := c.doRequest("POST", "gitea.update", payload)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 || string(resp) == "true" {
		return c.GetGiteaProvider(provider.ID)
	}

	var result GiteaProvider
	if err := json.Unmarshal(resp, &result); err != nil {
		return c.GetGiteaProvider(provider.ID)
	}
	return &result, nil
}

func (c *DokployClient) ListGiteaProviders() ([]GiteaProviderListItem, error) {
	resp, err := c.doRequest("GET", "gitea.giteaProviders", nil)
	if err != nil {
		return nil, err
	}

	// Try direct array response
	var providers []GiteaProviderListItem
	if err := json.Unmarshal(resp, &providers); err == nil {
		return providers, nil
	}

	// Try wrapper format
	var wrapper struct {
		Providers []GiteaProviderListItem `json:"providers"`
	}
	if err := json.Unmarshal(resp, &wrapper); err == nil {
		return wrapper.Providers, nil
	}

	// Try giteaProviders key
	var wrapper2 struct {
		Providers []GiteaProviderListItem `json:"giteaProviders"`
	}
	if err := json.Unmarshal(resp, &wrapper2); err == nil {
		return wrapper2.Providers, nil
	}

	return nil, fmt.Errorf("failed to parse gitea providers response")
}

// --- Organization ---

type Organization struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Slug      *string `json:"slug"`
	Logo      *string `json:"logo"`
	CreatedAt string  `json:"createdAt"`
	OwnerID   string  `json:"ownerId"`
}

func (c *DokployClient) CreateOrganization(name string, logo *string) (*Organization, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if logo != nil && *logo != "" {
		payload["logo"] = *logo
	}

	resp, err := c.doRequest("POST", "organization.create", payload)
	if err != nil {
		return nil, err
	}

	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetOrganization(id string) (*Organization, error) {
	endpoint := fmt.Sprintf("organization.one?organizationId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateOrganization(org Organization) (*Organization, error) {
	payload := map[string]interface{}{
		"organizationId": org.ID,
		"name":           org.Name,
	}
	if org.Logo != nil && *org.Logo != "" {
		payload["logo"] = *org.Logo
	}

	resp, err := c.doRequest("POST", "organization.update", payload)
	if err != nil {
		return nil, err
	}

	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteOrganization(id string) error {
	payload := map[string]string{
		"organizationId": id,
	}
	_, err := c.doRequest("POST", "organization.delete", payload)
	return err
}

func (c *DokployClient) ListOrganizations() ([]Organization, error) {
	resp, err := c.doRequest("GET", "organization.all", nil)
	if err != nil {
		return nil, err
	}

	var result []Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Volume Backup ---

type VolumeBackup struct {
	VolumeBackupID  string  `json:"volumeBackupId"`
	Name            string  `json:"name"`
	VolumeName      string  `json:"volumeName"`
	Prefix          string  `json:"prefix"`
	ServiceType     string  `json:"serviceType"`
	AppName         string  `json:"appName"`
	ServiceName     *string `json:"serviceName"`
	TurnOff         bool    `json:"turnOff"`
	CronExpression  string  `json:"cronExpression"`
	KeepLatestCount int     `json:"keepLatestCount"`
	Enabled         bool    `json:"enabled"`
	DestinationID   string  `json:"destinationId"`
	CreatedAt       string  `json:"createdAt"`
	// Service IDs (only one will be set based on serviceType)
	ApplicationID *string `json:"applicationId"`
	PostgresID    *string `json:"postgresId"`
	MariadbID     *string `json:"mariadbId"`
	MongoID       *string `json:"mongoId"`
	MysqlID       *string `json:"mysqlId"`
	RedisID       *string `json:"redisId"`
	ComposeID     *string `json:"composeId"`
}

func (c *DokployClient) CreateVolumeBackup(backup VolumeBackup) (*VolumeBackup, error) {
	payload := map[string]interface{}{
		"name":           backup.Name,
		"volumeName":     backup.VolumeName,
		"prefix":         backup.Prefix,
		"cronExpression": backup.CronExpression,
		"destinationId":  backup.DestinationID,
		"serviceType":    backup.ServiceType,
		"appName":        backup.AppName,
	}

	if backup.ServiceName != nil && *backup.ServiceName != "" {
		payload["serviceName"] = *backup.ServiceName
	}
	if backup.KeepLatestCount > 0 {
		payload["keepLatestCount"] = backup.KeepLatestCount
	}
	payload["turnOff"] = backup.TurnOff
	payload["enabled"] = backup.Enabled

	// Set the appropriate service ID based on service type
	if backup.ApplicationID != nil {
		payload["applicationId"] = *backup.ApplicationID
	}
	if backup.PostgresID != nil {
		payload["postgresId"] = *backup.PostgresID
	}
	if backup.MysqlID != nil {
		payload["mysqlId"] = *backup.MysqlID
	}
	if backup.MariadbID != nil {
		payload["mariadbId"] = *backup.MariadbID
	}
	if backup.MongoID != nil {
		payload["mongoId"] = *backup.MongoID
	}
	if backup.RedisID != nil {
		payload["redisId"] = *backup.RedisID
	}
	if backup.ComposeID != nil {
		payload["composeId"] = *backup.ComposeID
	}

	resp, err := c.doRequest("POST", "volumeBackups.create", payload)
	if err != nil {
		return nil, err
	}

	var result VolumeBackup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) GetVolumeBackup(id string) (*VolumeBackup, error) {
	endpoint := fmt.Sprintf("volumeBackups.one?volumeBackupId=%s", id)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result VolumeBackup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) UpdateVolumeBackup(backup VolumeBackup) (*VolumeBackup, error) {
	payload := map[string]interface{}{
		"volumeBackupId": backup.VolumeBackupID,
		"name":           backup.Name,
		"volumeName":     backup.VolumeName,
		"prefix":         backup.Prefix,
		"cronExpression": backup.CronExpression,
		"destinationId":  backup.DestinationID,
	}

	if backup.ServiceName != nil && *backup.ServiceName != "" {
		payload["serviceName"] = *backup.ServiceName
	}
	if backup.KeepLatestCount > 0 {
		payload["keepLatestCount"] = backup.KeepLatestCount
	}
	payload["turnOff"] = backup.TurnOff
	payload["enabled"] = backup.Enabled

	resp, err := c.doRequest("POST", "volumeBackups.update", payload)
	if err != nil {
		return nil, err
	}

	var result VolumeBackup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *DokployClient) DeleteVolumeBackup(id string) error {
	payload := map[string]string{
		"volumeBackupId": id,
	}
	_, err := c.doRequest("POST", "volumeBackups.delete", payload)
	return err
}

func (c *DokployClient) ListVolumeBackups(serviceID, serviceType string) ([]VolumeBackup, error) {
	endpoint := fmt.Sprintf("volumeBackups.list?id=%s&volumeBackupType=%s", serviceID, serviceType)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result []VolumeBackup
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
