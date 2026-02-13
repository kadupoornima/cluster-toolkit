# Code Review Report

## Key Findings
The most critical security risk is the default Service Account scope in the VM instance module, which grants full project access. Additionally, the default network configuration is permissive (public IPs enabled by default, broad internal firewall rules), and some installation scripts use deprecated or insecure package verification methods.

## Prioritized List of Suggested Changes

| File/Line | Issue Type | Recommendation | Reasoning |
| :--- | :--- | :--- | :--- |
| `modules/compute/vm-instance/variables.tf`:150 | **Security (Critical)** | Change default `service_account_scopes` to minimal required scopes (e.g., logging, monitoring) or empty. | Defaulting to `cloud-platform` scope grants full access to all Google Cloud APIs, violating the principle of least privilege. |
| `modules/compute/vm-instance/variables.tf`:108 | **Security (High)** | Change default `disable_public_ips` to `true`. | Defaulting to public IPs increases the attack surface. HPC clusters should ideally be private by default with controlled ingress (e.g., via IAP or Bastion). |
| `modules/compute/vm-instance/variables.tf`:245 | **Security (High)** | Restrict `enable_oslogin` to `ENABLE` or `INHERIT` only, or add a strong warning description. | Allowing `DISABLE` bypasses IAM-based SSH access control, forcing reliance on SSH keys which are harder to audit and revoke. |
| `modules/file-system/pre-existing-network-storage/scripts/install-gcs-fuse.sh`:29 | **Security (High)** | Update GPG key handling to use `signed-by` instead of `apt-key add`. | `apt-key` is deprecated and considers keys trusted for all repositories, which is less secure than scoping keys to specific repositories. |
| `modules/file-system/pre-existing-network-storage/scripts/install-gcs-fuse.sh`:22 | **Security (Medium)** | Enable `repo_gpgcheck=1` if possible, or document why it is disabled. | Disabling GPG check for the repo (`repo_gpgcheck=0`) allows potential man-in-the-middle attacks on package metadata. |
| `modules/network/vpc/main.tf`:141 | **Security (Medium)** | Review `allow_internal_traffic` to be more granular. | Allowing all protocols/ports (`0-65535`) internally facilitates lateral movement if a node is compromised. While common in HPC, it should be configurable. |
| `modules/compute/vm-instance/main.tf` | **Maintainability (Medium)** | Refactor complex `locals` logic (especially `network_interfaces` and `enable_oslogin`) into helper modules. | High cyclomatic complexity in `locals` makes the module hard to debug, test, and maintain. |
| `modules/file-system/*/scripts/install-nfs-client.sh` | **Maintainability (Medium)** | Consolidate duplicate scripts into a shared script library. | Code duplication between `filestore` and `pre-existing-network-storage` increases the maintenance burden and risk of inconsistency. |
| `modules/network/vpc/variables.tf`:38 | **Maintainability (Low)** | Remove deprecated variables (`subnetwork_size`, `primary_subnetwork`, `additional_subnetworks`). | Deprecated code clutters the codebase and confuses users about which variables are active. |

## Missing Best Practices

1.  **CMEK Support**: The `vm-instance` and `filestore` modules currently lack explicit support for **Customer-Managed Encryption Keys (CMEK)**. Adding this would allow enterprise users to manage their own encryption keys for compliance.
2.  **VPC Service Controls**: Ensure all modules are compatible with VPC Service Controls (VPC-SC) to prevent data exfiltration.
3.  **Script Robustness**: Ensure all shell scripts consistently use `set -euo pipefail` to fail fast on errors and undefined variables. Currently, some scripts only use `set -e` or `#!/bin/sh` without strict mode.
4.  **Package Version Pinning**: Installation scripts (e.g., `install-nfs-client.sh`) should pin package versions to ensure reproducibility and prevent unexpected breakage from upstream updates.
