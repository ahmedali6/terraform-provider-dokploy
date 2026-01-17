package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNewProjectResource(t *testing.T) {
	r := NewProjectResource()
	if r == nil {
		t.Fatal("Expected project resource, got nil")
	}

	_, ok := r.(resource.Resource)
	if !ok {
		t.Error("Expected project resource to implement Resource interface")
	}
}

func TestNewEnvironmentResource(t *testing.T) {
	r := NewEnvironmentResource()
	if r == nil {
		t.Fatal("Expected environment resource, got nil")
	}
}

func TestNewApplicationResource(t *testing.T) {
	r := NewApplicationResource()
	if r == nil {
		t.Fatal("Expected application resource, got nil")
	}
}

func TestNewComposeResource(t *testing.T) {
	r := NewComposeResource()
	if r == nil {
		t.Fatal("Expected compose resource, got nil")
	}
}

func TestNewDomainResource(t *testing.T) {
	r := NewDomainResource()
	if r == nil {
		t.Fatal("Expected domain resource, got nil")
	}
}

func TestNewCertificateResource(t *testing.T) {
	r := NewCertificateResource()
	if r == nil {
		t.Fatal("Expected certificate resource, got nil")
	}
}

func TestNewSSHKeyResource(t *testing.T) {
	r := NewSSHKeyResource()
	if r == nil {
		t.Fatal("Expected SSH key resource, got nil")
	}
}

func TestNewMountResource(t *testing.T) {
	r := NewMountResource()
	if r == nil {
		t.Fatal("Expected mount resource, got nil")
	}
}

func TestNewPortResource(t *testing.T) {
	r := NewPortResource()
	if r == nil {
		t.Fatal("Expected port resource, got nil")
	}
}

func TestNewRedirectResource(t *testing.T) {
	r := NewRedirectResource()
	if r == nil {
		t.Fatal("Expected redirect resource, got nil")
	}
}

func TestNewRegistryResource(t *testing.T) {
	r := NewRegistryResource()
	if r == nil {
		t.Fatal("Expected registry resource, got nil")
	}
}

func TestNewDestinationResource(t *testing.T) {
	r := NewDestinationResource()
	if r == nil {
		t.Fatal("Expected destination resource, got nil")
	}
}

func TestNewBackupResource(t *testing.T) {
	r := NewBackupResource()
	if r == nil {
		t.Fatal("Expected backup resource, got nil")
	}
}

func TestNewServerResource(t *testing.T) {
	r := NewServerResource()
	if r == nil {
		t.Fatal("Expected server resource, got nil")
	}
}

func TestNewRedisResource(t *testing.T) {
	r := NewRedisResource()
	if r == nil {
		t.Fatal("Expected Redis resource, got nil")
	}
}

func TestNewPostgresResource(t *testing.T) {
	r := NewPostgresResource()
	if r == nil {
		t.Fatal("Expected PostgreSQL resource, got nil")
	}
}

func TestNewMySQLResource(t *testing.T) {
	r := NewMySQLResource()
	if r == nil {
		t.Fatal("Expected MySQL resource, got nil")
	}
}

func TestNewMariaDBResource(t *testing.T) {
	r := NewMariaDBResource()
	if r == nil {
		t.Fatal("Expected MariaDB resource, got nil")
	}
}

func TestNewMongoDBResource(t *testing.T) {
	r := NewMongoDBResource()
	if r == nil {
		t.Fatal("Expected MongoDB resource, got nil")
	}
}

func TestNewGitlabProviderResource(t *testing.T) {
	r := NewGitlabProviderResource()
	if r == nil {
		t.Fatal("Expected GitLab provider resource, got nil")
	}
}

func TestNewBitbucketProviderResource(t *testing.T) {
	r := NewBitbucketProviderResource()
	if r == nil {
		t.Fatal("Expected Bitbucket provider resource, got nil")
	}
}

func TestNewGiteaProviderResource(t *testing.T) {
	r := NewGiteaProviderResource()
	if r == nil {
		t.Fatal("Expected Gitea provider resource, got nil")
	}
}

func TestNewOrganizationResource(t *testing.T) {
	r := NewOrganizationResource()
	if r == nil {
		t.Fatal("Expected organization resource, got nil")
	}
}

func TestNewVolumeBackupResource(t *testing.T) {
	r := NewVolumeBackupResource()
	if r == nil {
		t.Fatal("Expected volume backup resource, got nil")
	}
}

func TestNewApiKeyResource(t *testing.T) {
	r := NewApiKeyResource()
	if r == nil {
		t.Fatal("Expected API key resource, got nil")
	}
}

func TestNewUserPermissionsResource(t *testing.T) {
	r := NewUserPermissionsResource()
	if r == nil {
		t.Fatal("Expected user permissions resource, got nil")
	}
}

func TestNewAIResource(t *testing.T) {
	r := NewAIResource()
	if r == nil {
		t.Fatal("Expected AI resource, got nil")
	}
}

func TestNewEnvironmentVariablesResource(t *testing.T) {
	r := NewEnvironmentVariablesResource()
	if r == nil {
		t.Fatal("Expected environment variables resource, got nil")
	}
}

func TestNewServersDataSource(t *testing.T) {
	ds := NewServersDataSource()
	if ds == nil {
		t.Fatal("Expected servers data source, got nil")
	}

	_, ok := ds.(datasource.DataSource)
	if !ok {
		t.Error("Expected servers data source to implement DataSource interface")
	}
}

func TestNewGithubProvidersDataSource(t *testing.T) {
	ds := NewGithubProvidersDataSource()
	if ds == nil {
		t.Fatal("Expected GitHub providers data source, got nil")
	}
}

func TestNewGitlabProvidersDataSource(t *testing.T) {
	ds := NewGitlabProvidersDataSource()
	if ds == nil {
		t.Fatal("Expected GitLab providers data source, got nil")
	}
}

func TestNewBitbucketProvidersDataSource(t *testing.T) {
	ds := NewBitbucketProvidersDataSource()
	if ds == nil {
		t.Fatal("Expected Bitbucket providers data source, got nil")
	}
}

func TestNewGiteaProvidersDataSource(t *testing.T) {
	ds := NewGiteaProvidersDataSource()
	if ds == nil {
		t.Fatal("Expected Gitea providers data source, got nil")
	}
}

func TestNewBackupFilesDataSource(t *testing.T) {
	ds := NewBackupFilesDataSource()
	if ds == nil {
		t.Fatal("Expected backup files data source, got nil")
	}
}

func TestNewOrganizationsDataSource(t *testing.T) {
	ds := NewOrganizationsDataSource()
	if ds == nil {
		t.Fatal("Expected organizations data source, got nil")
	}
}

func TestNewVolumeBackupsDataSource(t *testing.T) {
	ds := NewVolumeBackupsDataSource()
	if ds == nil {
		t.Fatal("Expected volume backups data source, got nil")
	}
}

func TestNewUserDataSource(t *testing.T) {
	ds := NewUserDataSource()
	if ds == nil {
		t.Fatal("Expected user data source, got nil")
	}
}

func TestNewUsersDataSource(t *testing.T) {
	ds := NewUsersDataSource()
	if ds == nil {
		t.Fatal("Expected users data source, got nil")
	}
}

func TestNewAIsDataSource(t *testing.T) {
	ds := NewAIsDataSource()
	if ds == nil {
		t.Fatal("Expected AIs data source, got nil")
	}
}

func TestNewAIModelsDataSource(t *testing.T) {
	ds := NewAIModelsDataSource()
	if ds == nil {
		t.Fatal("Expected AI models data source, got nil")
	}
}

func TestNewApplicationDataSource(t *testing.T) {
	ds := NewApplicationDataSource()
	if ds == nil {
		t.Fatal("Expected application data source, got nil")
	}
}

func TestNewApplicationsDataSource(t *testing.T) {
	ds := NewApplicationsDataSource()
	if ds == nil {
		t.Fatal("Expected applications data source, got nil")
	}
}

func TestNewCertificateDataSource(t *testing.T) {
	ds := NewCertificateDataSource()
	if ds == nil {
		t.Fatal("Expected certificate data source, got nil")
	}
}

func TestNewCertificatesDataSource(t *testing.T) {
	ds := NewCertificatesDataSource()
	if ds == nil {
		t.Fatal("Expected certificates data source, got nil")
	}
}

func TestNewComposeDataSource(t *testing.T) {
	ds := NewComposeDataSource()
	if ds == nil {
		t.Fatal("Expected compose data source, got nil")
	}
}

func TestNewComposesDataSource(t *testing.T) {
	ds := NewComposesDataSource()
	if ds == nil {
		t.Fatal("Expected composes data source, got nil")
	}
}

func TestNewDeploymentsDataSource(t *testing.T) {
	ds := NewDeploymentsDataSource()
	if ds == nil {
		t.Fatal("Expected deployments data source, got nil")
	}
}

func TestNewDestinationDataSource(t *testing.T) {
	ds := NewDestinationDataSource()
	if ds == nil {
		t.Fatal("Expected destination data source, got nil")
	}
}

func TestNewDestinationsDataSource(t *testing.T) {
	ds := NewDestinationsDataSource()
	if ds == nil {
		t.Fatal("Expected destinations data source, got nil")
	}
}

func TestNewDockerContainerDataSource(t *testing.T) {
	ds := NewDockerContainerDataSource()
	if ds == nil {
		t.Fatal("Expected docker container data source, got nil")
	}
}

func TestNewDockerContainersDataSource(t *testing.T) {
	ds := NewDockerContainersDataSource()
	if ds == nil {
		t.Fatal("Expected docker containers data source, got nil")
	}
}
