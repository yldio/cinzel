step "echo" {
  run = "echo scheduled"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "schedule_workflow" {
  filename = "schedule_workflow"

  on "schedule" {
    cron = ["0 0 * * *", "0 12 * * *"]
  }

  jobs = [job.build]
}
