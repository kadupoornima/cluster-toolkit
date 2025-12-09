# Feature Roadmap

## Short-Term Features

### 1. `gcluster version` Command
**Description:** Add a dedicated `version` command to the CLI. currently, the version is only available via the `--version` flag. A dedicated command allows for more verbose output, such as printing the versions of external dependencies (Terraform, Packer, Ansible) which are critical for the tool's operation.

**Implementation Plan:**
- Create `cmd/version.go`.
- Move and expand the version logic from `cmd/root.go` to this new command.
- Implement checks for `terraform --version`, `packer --version`, and `ansible --version`.
- Register the command in the root command.

**Difficulty:** Low

### 2. `gcluster init` Command
**Description:** A command to bootstrap a new project by creating a basic `blueprint.yaml` file. This helps new users get started quickly without copying examples manually.

**Implementation Plan:**
- Create `cmd/init.go`.
- Define a minimal valid blueprint structure in Go or as an embedded string.
- When `gcluster init` is run, write this content to `blueprint.yaml` in the current directory (fail if file exists).
- Optionally allow selecting from a few templates (e.g., Slurm, GKE).

**Difficulty:** Low/Medium

### 3. JSON Output for `create` and `deploy`
**Description:** Add a `--json` flag to the `create` and `deploy` commands to output the result in a machine-readable format. This is useful for CI/CD pipelines and automation.

**Implementation Plan:**
- Define a struct for the output (e.g., deployment path, status, validation errors).
- Add the `--json` flag to the relevant commands.
- If the flag is set, marshal the result struct to JSON and print it to stdout instead of the usual logs.

**Difficulty:** Medium

## Long-Term Features

### 1. Interactive Blueprint Creator (`gcluster wizard`)
**Description:** A text-based user interface (TUI) to guide users through creating a blueprint. It would ask questions like "What type of workload?", "Which region?", "How many nodes?" and generate the corresponding YAML.

**Implementation Plan:**
- Use a TUI library like `bubbletea` or `survey`.
- Create a decision tree for the questions.
- Generate the `config.Blueprint` object in memory based on answers.
- Serialize the object to YAML.

**Difficulty:** High

### 2. Cost Estimation (`gcluster cost`)
**Description:** Estimate the cost of a blueprint before deployment. This is a high-value feature for HPC users who deploy expensive resources.

**Implementation Plan:**
- Parse the blueprint to identify all resources (VMs, Disks, Filestore, etc.).
- Use the Google Cloud Pricing API or integrate with a tool like `infracost`.
- Map the resources to billing SKUs.
- Calculate and display the estimated hourly/monthly cost.

**Difficulty:** High

### 3. State Visualization / Dashboard
**Description:** A web UI or rich TUI to visualize the state of the deployment. It would show which nodes are up, which jobs are running (if Slurm/K8s integration is added), and monitoring metrics.

**Implementation Plan:**
- Build a lightweight web server embedded in the CLI (`gcluster ui`).
- Use the GCP SDK to query the status of resources.
- Use SSH or Scheduler APIs to get job status.
- Render this information in a dashboard.

**Difficulty:** High
