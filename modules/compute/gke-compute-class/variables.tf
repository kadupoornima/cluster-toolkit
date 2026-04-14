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

variable "cluster_id" {
  description = "GKE cluster ID provided via 'use: [gke-cluster]'."
  type        = string
}

variable "compute_classes" {
  description = "A map of Custom Compute Classes to deploy to the cluster."
  type = map(object({
    node_pool_auto_creation = optional(bool, false)
    when_unsatisfiable      = optional(string, "DoNotScaleUp")
    priorities = list(object({
      machine_family = optional(string)
      machine_type   = optional(string)
      spot           = optional(bool, false)
      min_cores      = optional(number)
    }))
  }))
  default = {}
}
