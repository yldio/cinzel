step "resolve_tag" {
  run = <<-EOF
tag="$${{ github.event.inputs.tag }}"
tag="$${tag#v}"

if [ -n "$tag" ] && [ "$bump" != "none" ]; then
  echo "using tag: $tag"
  echo "tag=$tag" >> "$GITHUB_OUTPUT"
else
  echo "skip=true" >> "$GITHUB_OUTPUT"
fi
EOF
}

job "release" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.resolve_tag]
}

workflow "run_script_roundtrip" {
  filename = "run_script_roundtrip"

  on "workflow_dispatch" {}

  jobs = [job.release]
}
