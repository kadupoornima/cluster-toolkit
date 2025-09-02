/**
  * Copyright 2023 Google LLC
  *
  * Licensed under the Apache License, Version 2.0 (the "License");
  * you may not use this file except in compliance with the License.
  * You may obtain a copy of the License at
  *
  *      http://www.apache.org/licenses/LICENSE-2.0
  *
  * Unless required by applicable law or agreed to in writing, software
  * distributed under the License is distributed on an "AS IS" BASIS,
  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  * See the License for the specific language governing permissions and
  * limitations under the License.
  */

terraform {
  backend "gcs" {
    bucket = "g4-gke-bucket-1"
    prefix = "gke-g4/poornima-g4-12/primary"
  }
}

module "gke-g4-net-0" {
  source          = "./modules/embedded/modules/network/vpc"
  deployment_name = var.deployment_name
  firewall_rules = [{
    allow = [{
      ports    = ["0-65535"]
      protocol = "tcp"
      }, {
      ports    = ["0-65535"]
      protocol = "udp"
      }, {
      protocol = "icmp"
    }]
    name   = "${var.deployment_name}-internal-0"
    ranges = ["192.168.0.0/16"]
  }]
  labels       = var.labels
  network_name = "${var.deployment_name}-net-0"
  project_id   = var.project_id
  region       = var.region
  secondary_ranges_list = [{
    ranges = [{
      ip_cidr_range = "10.4.0.0/14"
      range_name    = "pods"
      }, {
      ip_cidr_range = "10.0.32.0/20"
      range_name    = "services"
    }]
    subnetwork_name = "${var.deployment_name}-sub-0"
  }]
  subnetworks = [{
    subnet_ip     = "192.168.0.0/18"
    subnet_name   = "${var.deployment_name}-sub-0"
    subnet_region = var.region
  }]
}

module "gke-g4-net-1" {
  source          = "./modules/embedded/modules/network/vpc"
  deployment_name = var.deployment_name
  firewall_rules = [{
    allow = [{
      ports    = ["0-65535"]
      protocol = "tcp"
      }, {
      ports    = ["0-65535"]
      protocol = "udp"
      }, {
      protocol = "icmp"
    }]
    name   = "${var.deployment_name}-internal-1"
    ranges = ["192.168.0.0/16"]
  }]
  labels       = var.labels
  network_name = "${var.deployment_name}-net-1"
  project_id   = var.project_id
  region       = var.region
  subnetworks = [{
    subnet_ip     = "192.168.64.0/18"
    subnet_name   = "${var.deployment_name}-sub-1"
    subnet_region = var.region
  }]
}

module "node_pool_service_account" {
  source          = "./modules/embedded/community/modules/project/service-account"
  deployment_name = var.deployment_name
  name            = "gke-np-sa"
  project_id      = var.project_id
  project_roles   = ["logging.logWriter", "monitoring.metricWriter", "monitoring.viewer", "stackdriver.resourceMetadata.writer", "storage.objectViewer", "artifactregistry.reader"]
}

module "workload_service_account" {
  source          = "./modules/embedded/community/modules/project/service-account"
  deployment_name = var.deployment_name
  name            = "gke-wl-sa"
  project_id      = var.project_id
  project_roles   = ["logging.logWriter", "monitoring.metricWriter", "monitoring.viewer", "stackdriver.resourceMetadata.writer", "storage.objectAdmin", "artifactregistry.reader"]
}

module "g4-cluster" {
  source                         = "./modules/embedded/modules/scheduler/gke-cluster"
  additional_networks            = concat([{ network = module.gke-g4-net-1.network_name, subnetwork = module.gke-g4-net-1.subnetwork_name, subnetwork_project = var.project_id, nic_type = "GVNIC", queue_count = null, network_ip = null, stack_type = null, access_config = [{ nat_ip = null, public_ptr_domain_name = null, network_tier = null }], ipv6_access_config = [], alias_ip_range = [] }])
  configure_workload_identity_sa = true
  deployment_name                = var.deployment_name
  enable_dcgm_monitoring         = true
  enable_private_endpoint        = false
  k8s_service_account_name       = var.k8s_service_account_name
  labels                         = var.labels
  maintenance_exclusions = [{
    end_time        = "2026-04-10T00:00:00Z"
    exclusion_scope = "NO_MINOR_OR_NODE_UPGRADES"
    name            = "no-minor-or-node-upgrades-indefinite"
    start_time      = "2025-08-01T00:00:00Z"
  }]
  master_authorized_networks = [{
    cidr_block   = var.authorized_cidr
    display_name = "kubectl-access-network"
  }]
  network_id                    = module.gke-g4-net-0.network_id
  project_id                    = var.project_id
  region                        = var.region
  release_channel               = "RAPID"
  service_account_email         = module.workload_service_account.service_account_email
  subnetwork_self_link          = module.gke-g4-net-0.subnetwork_self_link
  system_node_pool_disk_size_gb = var.system_node_pool_disk_size_gb
  system_node_pool_machine_type = "e2-standard-16"
  system_node_pool_taints       = []
  version_prefix                = "1.32."
  zone                          = var.zone
}

module "g4-pool" {
  source              = "./modules/embedded/modules/compute/gke-node-pool"
  additional_networks = concat([{ network = module.gke-g4-net-1.network_name, subnetwork = module.gke-g4-net-1.subnetwork_name, subnetwork_project = var.project_id, nic_type = "GVNIC", queue_count = null, network_ip = null, stack_type = null, access_config = [{ nat_ip = null, public_ptr_domain_name = null, network_tier = null }], ipv6_access_config = [], alias_ip_range = [] }])
  auto_upgrade        = true
  cluster_id          = module.g4-cluster.cluster_id
  gke_version         = module.g4-cluster.gke_version
  guest_accelerator = [{
    count = 1
    gpu_driver_installation_config = {
      gpu_driver_version = "LATEST"
    }
    type = "nvidia-rtx-pro-6000"
  }]
  internal_ghpc_module_id = "g4-pool"
  labels                  = var.labels
  machine_type            = "g4-standard-48"
  project_id              = var.project_id
  service_account_email   = module.node_pool_service_account.service_account_email
  static_node_count       = var.static_node_count
  zones                   = [var.zone]
}

module "workload-manager-install" {
  source             = "./modules/embedded/modules/management/kubectl-apply"
  cluster_id         = module.g4-cluster.cluster_id
  gke_cluster_exists = module.g4-cluster.gke_cluster_exists
  jobset = {
    install = true
  }
  kueue = {
    install = true
  }
  project_id = var.project_id
}

module "nvidia_smi_job_template" {
  source                   = "./modules/embedded/modules/compute/gke-job-template"
  allocatable_cpu_per_node = flatten([module.g4-pool.allocatable_cpu_per_node])
  allocatable_gpu_per_node = flatten([module.g4-pool.allocatable_gpu_per_node])
  command                  = ["nvidia-smi"]
  has_gpu                  = flatten([module.g4-pool.has_gpu])
  image                    = "nvidia/cuda:13.0.0-base-ubuntu24.04"
  k8s_service_account_name = var.k8s_service_account_name
  labels                   = var.labels
  name                     = "run-nvidia-smi"
  node_count               = var.static_node_count
  node_pool_names          = flatten([module.g4-pool.node_pool_names])
  tolerations              = flatten([module.g4-pool.tolerations])
}
