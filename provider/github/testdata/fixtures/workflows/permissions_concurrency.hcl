step "deploy" {
  run = "echo deploying"
}

job "deploy" {
  runs_on {
    runners = "ubuntu-latest"
  }

  permissions {
    contents    = "read"
    deployments = "write"
  }

  concurrency {
    group              = "deploy-prod"
    cancel_in_progress = true
  }

  environment {
    name = "production"
    url  = "https://example.com"
  }

  steps = [step.deploy]
}

workflow "deploy_wf" {
  filename = "permissions_concurrency"

  permissions {
    contents = "read"
    packages = "write"
  }

  concurrency {
    group              = "main-deploy"
    cancel_in_progress = false
  }

  defaults {
    run {
      shell             = "bash"
      working_directory = "./src"
    }
  }

  on "push" {
    branches = ["main"]
  }

  jobs = [job.deploy]
}
