# Copyright 2026 "Google LLC"
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  # Split the cluster_id to extract project, location, and name
  cluster_id_parts = split("/", var.cluster_id)
  project_id       = local.cluster_id_parts[1]
  location         = local.cluster_id_parts[3]
  cluster_name     = local.cluster_id_parts[5]
}

# Fetch an active access token
data "google_client_config" "default" {}

# Fetch the newly created GKE cluster's details
data "google_container_cluster" "gke_cluster" {
  project  = local.project_id
  location = local.location
  name     = local.cluster_name
}

# Configure the Kubernetes provider using the fetched cluster details
provider "kubernetes" {
  host                   = "https://${data.google_container_cluster.gke_cluster.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.gke_cluster.master_auth[0].cluster_ca_certificate)
}

resource "kubernetes_manifest" "compute_class" {
  for_each = var.compute_classes

  manifest = {
    "apiVersion" = "cloud.google.com/v1"
    "kind"       = "ComputeClass"
    "metadata" = {
      "name" = each.key
    }
    "spec" = {
      "nodePoolAutoCreation" = {
        "enabled" = each.value.node_pool_auto_creation
      }
      "whenUnsatisfiable" = each.value.when_unsatisfiable
      "priorities" = [
        for priority in each.value.priorities : {
          "machineFamily" = priority.machine_family
          "machineType"   = priority.machine_type
          "spot"          = priority.spot
          "minCores"      = priority.min_cores
        }
      ]
    }
  }
}
