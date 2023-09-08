/*
Check images https://github.com/actions/runner-images?tab=readme-ov-file#available-images
*/

file_name = "pull-requests"
single    = true

workflow "pull_requests" {
  name = "Pull Requests"
  on   = "pull_request"

  jobs = [
    job.pull_request
  ]
}

job "pull_request" {
  # ubuntu-20.04 already has GO 1.22.5 
  runs_on = "ubuntu-20.04"

  steps = [
    step.checkout,
    step.tests
  ]
}

step "checkout" {
  name = "Checkout"
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "tests" {
  name = "Tests"
  run  = <<-EOF
make test-ci
EOF
}