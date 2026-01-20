package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestInferSourceType(t *testing.T) {
	tests := []struct {
		name     string
		plan     *ApplicationResourceModel
		expected string
	}{
		{
			name: "docker image",
			plan: &ApplicationResourceModel{
				DockerImage: types.StringValue("nginx:latest"),
			},
			expected: "docker",
		},
		{
			name: "custom git url",
			plan: &ApplicationResourceModel{
				CustomGitUrl: types.StringValue("https://github.com/user/repo.git"),
			},
			expected: "git",
		},
		{
			name: "gitlab provider",
			plan: &ApplicationResourceModel{
				GitlabId: types.StringValue("gitlab-123"),
			},
			expected: "gitlab",
		},
		{
			name: "bitbucket provider",
			plan: &ApplicationResourceModel{
				BitbucketId: types.StringValue("bitbucket-456"),
			},
			expected: "bitbucket",
		},
		{
			name: "gitea provider",
			plan: &ApplicationResourceModel{
				GiteaId: types.StringValue("gitea-789"),
			},
			expected: "gitea",
		},
		{
			name: "default to github",
			plan: &ApplicationResourceModel{
				DockerImage:  types.StringNull(),
				CustomGitUrl: types.StringNull(),
				GitlabId:     types.StringNull(),
				BitbucketId:  types.StringNull(),
				GiteaId:      types.StringNull(),
			},
			expected: "github",
		},
		{
			name: "empty string values default to github",
			plan: &ApplicationResourceModel{
				DockerImage:  types.StringValue(""),
				CustomGitUrl: types.StringValue(""),
				GitlabId:     types.StringValue(""),
				BitbucketId:  types.StringValue(""),
				GiteaId:      types.StringValue(""),
			},
			expected: "github",
		},
		{
			name: "docker image takes precedence over git",
			plan: &ApplicationResourceModel{
				DockerImage:  types.StringValue("nginx:latest"),
				CustomGitUrl: types.StringValue("https://github.com/user/repo.git"),
			},
			expected: "docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferSourceType(tt.plan)
			if result.ValueString() != tt.expected {
				t.Errorf("Expected source type %s, got %s", tt.expected, result.ValueString())
			}
		})
	}
}

func TestInferSourceTypeWithUnknown(t *testing.T) {
	tests := []struct {
		name     string
		plan     *ApplicationResourceModel
		expected string
	}{
		{
			name: "unknown docker image",
			plan: &ApplicationResourceModel{
				DockerImage: types.StringUnknown(),
			},
			expected: "github",
		},
		{
			name: "unknown custom git url",
			plan: &ApplicationResourceModel{
				CustomGitUrl: types.StringUnknown(),
			},
			expected: "github",
		},
		{
			name: "unknown gitlab id",
			plan: &ApplicationResourceModel{
				GitlabId: types.StringUnknown(),
			},
			expected: "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferSourceType(tt.plan)
			if result.ValueString() != tt.expected {
				t.Errorf("Expected source type %s, got %s", tt.expected, result.ValueString())
			}
		})
	}
}

func TestInferSourceTypePriority(t *testing.T) {
	tests := []struct {
		name     string
		plan     *ApplicationResourceModel
		expected string
	}{
		{
			name: "docker has highest priority",
			plan: &ApplicationResourceModel{
				DockerImage:  types.StringValue("nginx:latest"),
				CustomGitUrl: types.StringValue("https://github.com/user/repo.git"),
				GitlabId:     types.StringValue("gitlab-123"),
				BitbucketId:  types.StringValue("bitbucket-456"),
				GiteaId:      types.StringValue("gitea-789"),
			},
			expected: "docker",
		},
		{
			name: "git has second priority",
			plan: &ApplicationResourceModel{
				CustomGitUrl: types.StringValue("https://github.com/user/repo.git"),
				GitlabId:     types.StringValue("gitlab-123"),
				BitbucketId:  types.StringValue("bitbucket-456"),
				GiteaId:      types.StringValue("gitea-789"),
			},
			expected: "git",
		},
		{
			name: "gitlab has third priority",
			plan: &ApplicationResourceModel{
				GitlabId:    types.StringValue("gitlab-123"),
				BitbucketId: types.StringValue("bitbucket-456"),
				GiteaId:     types.StringValue("gitea-789"),
			},
			expected: "gitlab",
		},
		{
			name: "bitbucket has fourth priority",
			plan: &ApplicationResourceModel{
				BitbucketId: types.StringValue("bitbucket-456"),
				GiteaId:     types.StringValue("gitea-789"),
			},
			expected: "bitbucket",
		},
		{
			name: "gitea has fifth priority",
			plan: &ApplicationResourceModel{
				GiteaId: types.StringValue("gitea-789"),
			},
			expected: "gitea",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferSourceType(tt.plan)
			if result.ValueString() != tt.expected {
				t.Errorf("Expected source type %s, got %s", tt.expected, result.ValueString())
			}
		})
	}
}

func TestInferComposeSourceType(t *testing.T) {
	tests := []struct {
		name     string
		plan     *ComposeResourceModel
		expected string
	}{
		{
			name: "raw compose file",
			plan: &ComposeResourceModel{
				ComposeFileContent: types.StringValue("version: '3'\nservices:\n  web:\n    image: nginx"),
			},
			expected: "raw",
		},
		{
			name: "custom git url",
			plan: &ComposeResourceModel{
				CustomGitUrl: types.StringValue("https://github.com/user/compose-repo.git"),
			},
			expected: "git",
		},
		{
			name: "gitlab provider",
			plan: &ComposeResourceModel{
				GitlabId: types.StringValue("gitlab-123"),
			},
			expected: "gitlab",
		},
		{
			name: "bitbucket provider",
			plan: &ComposeResourceModel{
				BitbucketId: types.StringValue("bitbucket-456"),
			},
			expected: "bitbucket",
		},
		{
			name: "gitea provider",
			plan: &ComposeResourceModel{
				GiteaId: types.StringValue("gitea-789"),
			},
			expected: "gitea",
		},
		{
			name: "default to github",
			plan: &ComposeResourceModel{
				ComposeFileContent: types.StringNull(),
				CustomGitUrl:       types.StringNull(),
				GitlabId:           types.StringNull(),
				BitbucketId:        types.StringNull(),
				GiteaId:            types.StringNull(),
			},
			expected: "github",
		},
		{
			name: "empty string values default to github",
			plan: &ComposeResourceModel{
				ComposeFileContent: types.StringValue(""),
				CustomGitUrl:       types.StringValue(""),
				GitlabId:           types.StringValue(""),
				BitbucketId:        types.StringValue(""),
				GiteaId:            types.StringValue(""),
			},
			expected: "github",
		},
		{
			name: "raw compose takes precedence over git",
			plan: &ComposeResourceModel{
				ComposeFileContent: types.StringValue("version: '3'"),
				CustomGitUrl:       types.StringValue("https://github.com/user/repo.git"),
			},
			expected: "raw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferComposeSourceType(tt.plan)
			if result.ValueString() != tt.expected {
				t.Errorf("Expected source type %s, got %s", tt.expected, result.ValueString())
			}
		})
	}
}

func TestInferComposeSourceTypeWithUnknown(t *testing.T) {
	tests := []struct {
		name     string
		plan     *ComposeResourceModel
		expected string
	}{
		{
			name: "unknown compose file content",
			plan: &ComposeResourceModel{
				ComposeFileContent: types.StringUnknown(),
			},
			expected: "github",
		},
		{
			name: "unknown custom git url",
			plan: &ComposeResourceModel{
				CustomGitUrl: types.StringUnknown(),
			},
			expected: "github",
		},
		{
			name: "unknown gitlab id",
			plan: &ComposeResourceModel{
				GitlabId: types.StringUnknown(),
			},
			expected: "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferComposeSourceType(tt.plan)
			if result.ValueString() != tt.expected {
				t.Errorf("Expected source type %s, got %s", tt.expected, result.ValueString())
			}
		})
	}
}
