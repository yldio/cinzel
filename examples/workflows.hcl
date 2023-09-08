workflow "workflow_0" {
  jobs = [
    job.job_1
  ]

  on_by_filter {
    event  = "label"
    filter = "types"
    values = ["created"]
  }

  on_by_filter {
    event  = "push"
    filter = "branches"
    values = ["main", "'mmona/octocat'", "'releases/**'"]
  }

  on_by_filter {
    event  = "push"
    filter = "tags"
    values = ["v2", "v1.*"]
  }
}

workflow "workflow_1" {
  name     = "Workflow 1"
  run_name = "Workflow 1 Deployment dev"

  on = "push"

  jobs = [
    job.job_1,
    job.job_2
  ]

  permissions {
    actions = "read"
  }

  env {
    name  = "ENVIRONMENT"
    value = "dev"
  }

  env {
    name  = "VERSION"
    value = "v1.0.0"
  }

  defaults {
    run {
      shell             = "bash"
      working_directory = "./scripts"
    }
  }

  concurrency {
    group              = "group-test-1"
    cancel_in_progress = true
  }
}

workflow "workflow_2" {
  name = "Workflow 2"

  on = "push"

  jobs = []
}

workflow "workflow_3" {
  name = "Workflow 3"

  on_as_list = ["push", "pull_request"]

  jobs = []
}

workflow "workflow_4" {
  name = "Workflow 4"

  on_by_filter {
    event  = "label"
    filter = "types"
    values = ["created"]
  }

  jobs = []
}

workflow "workflow_5" {
  name = "Workflow 5"

  on_by_filter {
    event  = "push"
    filter = "branches"
    values = ["main", "feat/*", 1]
  }

  on_by_filter {
    event  = "pull_request"
    filter = "branches"
    values = ["main"]
  }

  on_by_filter {
    event = "release"
  }

  jobs = []
}
