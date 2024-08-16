workflow "workflow_1" {
  filename = "dummy-file.yaml"
  on {
    event "push" {
      branches = ["main"]
      tags = ["v2"]
    }
  }
  on {
    activity "label" {
      types = ["created"]
    } 
  }
  jobs = [job.job_1, job.job_2]
}

job "job_1" {
  name = "job 1"
  steps = [step.step_1]
}

job "job_2" {
  name = "job 2"

  runs {
    on = "ubuntu-20.04"
  }

  needs = [job.job_1]

  steps = [step.step_2]
}

step "step_1" {
  name = "step 1"
}

step "step_2" {
  name = "step 2"
}