step "write_config" {
  run = <<-EOT
    cat > cfg.yaml <<'YAML'
    overrides: {}
    YAML
    echo done
  EOT
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.write_config]
}

workflow "ci" {
  filename = "workflow-parse-script-literal"

  on "push" {}

  jobs = [job.build]
}
