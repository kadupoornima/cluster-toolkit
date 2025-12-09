mock_provider "google" {}
mock_provider "google-beta" {}

run "verify_c3d_pd_standard_fails" {
  command = plan

  module {
    source = "../"
  }

  variables {
    project_id = "test-project"
    region = "us-central1"
    zone = "us-central1-a"
    deployment_name = "test-deployment"
    machine_type = "c3d-standard-4"
    disk_type = "pd-standard"
    network_self_link = "default"
    labels = {}
    instance_image = {
        project = "test-project"
        family = "test-family"
    }
  }

  expect_failures = [
    google_compute_instance.compute_vm
  ]
}
