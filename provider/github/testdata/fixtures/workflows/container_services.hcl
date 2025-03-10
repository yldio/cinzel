step "test" {
  run = "npm test"
}

job "integration" {
  runs_on {
    runners = "ubuntu-latest"
  }

  container {
    image = "node:18"

    env {
      name  = "NODE_ENV"
      value = "test"
    }
  }

  service "postgres" {
    image = "postgres:15"

    env {
      name  = "POSTGRES_PASSWORD"
      value = "test"
    }
  }

  service "redis" {
    image = "redis:7"
  }

  steps = [step.test]
}

workflow "integration_wf" {
  filename = "container_services"

  on "push" {}

  jobs = [job.integration]
}
