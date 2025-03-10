step "setup" {
  name = "Setup Node"

  uses {
    action  = "actions/setup-node"
    version = "v4"
  }

  with {
    name  = "node-version"
    value = "20"
  }
}

step "install" {
  name = "Install dependencies"
  run  = "npm ci"
}

step "build" {
  name = "Build"
  run  = "npm run build"

  env {
    name  = "NODE_ENV"
    value = "production"
  }
}

action "my_action" {
  filename    = "composite_action"
  name        = "My Composite Action"
  description = "A sample composite action"

  input "token" {
    description = "GitHub token"
    required    = true
  }

  input "working_directory" {
    description = "Working directory"
    required    = false
    default     = "."
  }

  output "result" {
    description = "Build result"
    value       = "$${{ steps.build.outputs.result }}"
  }

  runs {
    using = "composite"
    steps = [step.setup, step.install, step.build]
  }
}
