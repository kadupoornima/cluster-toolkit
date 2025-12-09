# Google Cluster Toolkit (formerly HPC Toolkit)

## Description

Cluster Toolkit is an open-source software offered by Google Cloud which makes it
easy for customers to deploy AI/ML and HPC environments on Google Cloud.

It allows for the creation of turnkey environments (compute, networking, storage, etc.)
following Google Cloud best-practices. The Toolkit is highly customizable and extensible,
addressing the needs of a broad range of customers.

## Quickstart

The following guide will help you install the Cluster Toolkit and deploy your first cluster.

### Prerequisites

To build and run `gcluster` from source, you need:
- **Go** (v1.24 or higher) - [Install Go](https://go.dev/doc/install)
- **Git** - [Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- **Make** - [Install Make](https://www.gnu.org/software/make/) (Standard on Linux/macOS)
- **Terraform** - [Install Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)
- **Packer** (Optional, for image building) - [Install Packer](https://learn.hashicorp.com/tutorials/packer/get-started-install-cli)

### Installation Guide

#### Linux & macOS

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/GoogleCloudPlatform/cluster-toolkit.git
    cd cluster-toolkit
    ```

2.  **Build the binary:**
    ```bash
    make
    ```
    This will create a `gcluster` binary in the current directory.

3.  **Add to PATH (Optional):**
    ```bash
    export PATH=$PATH:$(pwd)
    ```

#### Windows

We recommend using **WSL 2 (Windows Subsystem for Linux)** for the best experience, following the Linux instructions above.

If you must run natively on Windows:

1.  **Clone the repository:**
    ```powershell
    git clone https://github.com/GoogleCloudPlatform/cluster-toolkit.git
    cd cluster-toolkit
    ```

2.  **Build the binary:**
    ```powershell
    go build -o gcluster.exe gcluster.go
    ```

3.  **Add to PATH:**
    Add the `cluster-toolkit` directory to your System PATH environment variable.

### Usage Examples

#### 1. Basic Deployment

The most common workflow is to deploy a blueprint. A blueprint is a YAML file defining your cluster configuration.

To deploy the standard Slurm example:

```bash
# 1. Create the deployment folder from the blueprint
./gcluster create examples/hpc-slurm.yaml --vars "project_id=<YOUR_PROJECT_ID>"

# 2. Deploy the resources
./gcluster deploy hpc-slurm
```

This will create a `hpc-slurm` directory containing the Terraform code and then deploy it to your Google Cloud project.

#### 2. Single Command Deploy

You can also create and deploy in a single step (Note: `deploy` command accepts a blueprint file directly):

```bash
./gcluster deploy examples/hpc-slurm.yaml --vars "project_id=<YOUR_PROJECT_ID>"
```

#### 3. Customize Blueprint

You can override variables directly via the CLI:

```bash
./gcluster deploy examples/hpc-slurm.yaml \
  --vars "project_id=my-gcp-project" \
  --vars "region=us-west1" \
  --vars "zone=us-west1-b"
```

## Documentation

- **[Tutorials](docs/tutorials/README.md)**: Step-by-step guides.
- **[Examples](examples/README.md)**: Ready-to-use blueprints.
- **[Modules](modules/README.md)**: Documentation for individual infrastructure modules.
- **[Google Cloud Docs](https://cloud.google.com/cluster-toolkit/docs/overview)**: Official product documentation.

## GCP Credentials

### Supplying cloud credentials to Terraform

Terraform needs credentials to authenticate with Google Cloud.

**Recommended: Application Default Credentials**

Run the following command on your workstation:

```shell
gcloud auth application-default login
```

If you encounter quota errors, set your quota project:

```shell
gcloud auth application-default set-quota-project ${PROJECT-ID}
```

For more details on credentials in Cloud Shell or other environments, see [GCP Credentials Details](#gcp-credentials-details) below.

## Troubleshooting

Common deployment issues:

*   **GCP Access**: Ensure your credentials have `Owner` or sufficient permissions on the project.
*   **APIs**: Ensure required APIs (Compute Engine, etc.) are enabled.
*   **Quotas**: Check if you have enough quota for the requested machine types (e.g., N2, H3).

For detailed troubleshooting, refer to:
*   [Slurm Troubleshooting](docs/slurm-troubleshooting.md)
*   [Validation Documentation](docs/blueprint-validation.md)

---

## Detailed Reference

### GCP Credentials Details
Terraform can discover credentials for authenticating to Google Cloud Platform in several ways. We do **not** recommend following Hashicorp's instructions for downloading service account keys. Instead, use `gcloud auth application-default login` as described above.

In virtualized settings (like Cloud Shell or VMs), credentials may be inherited from the environment.

### VM Image Support
The Toolkit supports:
*   HPC Rocky Linux 8
*   Debian 11
*   Ubuntu 20.04 LTS

See [docs/vm-images.md](docs/vm-images.md) for more info.

### Development
For instructions on contributing to the project, setting up the development environment, and running tests, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache License 2.0](LICENSE)
