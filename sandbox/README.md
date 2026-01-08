# Sandbox for Manual Testing

This directory provides a safe environment for manually testing the Dokploy Terraform provider against a real instance.

## Prerequisites

- Go 1.21+ installed.
- Terraform 1.0+ installed.
- A running Dokploy instance and an API Key.

## Setup & Usage

1.  **Configure Environment**:
    Copy the example variables file and add your credentials:

    ```shell
    cd sandbox
    cp example.terraform.tfvars terraform.tfvars
    ```

2.  **Build the Provider**:
    From the **project root**, build the provider binary:

    ```shell
    go build -o terraform-provider-dokploy .
    ```

3.  **Initialize & Apply**:
    Run Terraform from within the `sandbox` directory, pointing to the local build configuration:

    ```shell
    # (Inside sandbox directory)

    # Export CLI config for local provider overrides
    export TF_CLI_CONFIG_FILE=$(pwd)/dev.tfrc

    # Initialize other providers (tls, github)
    terraform init

    # Apply changes
    terraform apply
    ```

## Cleaning Up

When finished, destroy all created resources:

```shell
export TF_CLI_CONFIG_FILE=$(pwd)/dev.tfrc
terraform destroy
```
