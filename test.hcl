workflow "test" {
  filename = "test"

  on "page_build" {
  }

  on "push" {
    branches = ["main"]
  }

  on "schedule" {
    cron = ["30 5 * * 1,3", "30 5 * * 2,4"]
  }

  on "workflow_call" {
    input "input-1" {
      type        = "string"
      required    = true
      description = "description input 1"
      default     = "default input 1"
    }

    output "output-1" {
      value       = "value output 1"
      description = "description output 1"
    }

    secret "secret-1" {
      description = "description secret 1"
      required    = true
    }
  }

  on "label" {
    types = ["created", "edited"]
  }

  permissions {
    actions         = "read"
    id_token        = "none"
    security_events = "write"
  }

  env {
    name  = "NODE_ENV"
    value = "development"
  }

  env {
    name  = "TOKEN"
    value = "$${{ secrets.token }}"
  }

  defaults {
    run {
      shell             = "bash"
      working_directory = "./scripts"
    }
  }

  concurrency {
    group              = "$${{ github.workflow }}-$${{ github.ref }}"
    cancel_in_progress = true
  }

  jobs = [
    job.test-job-1,
    job.test-job-2,
  ]
}

job "test-job-1" {
  name = "job 1"

  permissions {
    actions         = "read"
    id_token        = "none"
    security_events = "write"
  }

  needs = [
    job.test-job-2
  ]

  if = "$${{ ! startsWith(github.ref, 'refs/tags/') }}"

  runs_on {
    group  = "ubuntu-runners"
    labels = "ubuntu-20.04-16core"
  }

  environment {
    name = "production_environment"
    url  = "$${{ steps.step_id.outputs.url_output }}"
  }

  concurrency {
    group              = "$${{ github.workflow }}-$${{ github.ref }}"
    cancel_in_progress = true
  }

  output {
    name  = "output1"
    value = "-12.4242"
  }

  output {
    name  = "output2"
    value = "$${{ steps.step2.outputs.test }}"
  }

  env {
    name  = "TOKEN"
    value = "$${{ secrets.token }}"
  }

  env {
    name  = "NODE_ENV"
    value = "development"
  }

  defaults {
    run {
      shell             = "bash"
      working_directory = "./scripts"
    }
  }

  timeout_minutes = 30

  strategy {
    matrix {
      include = [{
        color = "green"
        }, {
        animal = "cat"
        color  = "pink"
      }]
      variable {
        name = "node"
        value = [{
          version = 14
          }, {
          env     = "NODE_DEBUG=acto*"
          version = 16
          }, {
          env     = "NODE_OPTIONS=--openssl-legacy-provider"
          version = 20
        }]
      }
      variable {
        name = "os"
        value  = ["ubuntu-latest", "macos-latest"]
      }
      exclude = [{
        environment = "production"
        os          = "macos-latest"
        version     = 12
      }]
    }

    fail_fast = true

    max_parallel = 2
  }

  continue_on_error = "$${{ matrix.experimental }}"

  container {
    image = "node:18"

    credentials {
      username = "$${{ github.actor }}"
      password = "$${{ secrets.github_token }}"
    }

    env {
      name  = "NODE_ENV"
      value = "development"
    }

    ports = [80]

    volumes = ["my_docker_volume:/volume_mount"]

    options = "--cpus 1"
  }

  service "nginx" {
    image = "$${{ options.nginx == true && 'nginx' || '' }}"

    credentials {
      username = "$${{ github.actor }}"
      password = "$${{ secrets.github_token }}"
    }

    env {
      name  = "NODE_ENV"
      value = "development"
    }

    ports = ["8080:80"]

    volumes = ["my_docker_volume:/volume_mount", "/source/directory:/destination/directory"]

    options = "--cpus 1"
  }

  service "redis" {
    image = "redis"

    ports = ["6379/tcp"]
  }
}

step "test-job-1-step-1" {
  id   = "step-1"
  name = "step 1"

  uses {
    action = "actions/checkout"
    version = "v4"
  }

  with {
    name  = "args"
    value = "The $${{ github.event_name }} event triggered this step."
  }

  with {
    name  = "entrypoint"
    value = "/a/different/executable"
  }
}

step "test-job-1-step-2" {
  run = <<EOF
npm ci
npm run build
EOF

  working_directory = "./temp"

  shell = "bash"

  env {
    name  = "GITHUB_TOKEN"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }

  continue_on_error = true

  timeout_minutes = 30
}

job "test-job-2" {
   name = "job 2"

   uses {
    action = "octo-org/another-repo/.github/workflows/workflow.yml"
    version = "v1"
   }

   with {
     name  = "username"
     value = "mona"
   }

   secrets = "inherit"
 }
