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

output "instructions_g4-cluster" {
  description = "Generated output from module 'g4-cluster'"
  value       = module.g4-cluster.instructions
}

output "instructions_g4-pool" {
  description = "Generated output from module 'g4-pool'"
  value       = module.g4-pool.instructions
}

output "instructions_nvidia_smi_job_template" {
  description = "Generated output from module 'nvidia_smi_job_template'"
  value       = module.nvidia_smi_job_template.instructions
}
