step "step_1" {
  name = "Step 1"

  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "step_2" {
  name = "Step 2"

  // TODO: make it read relative...
  run = import_script("../examples/steps/echo_hello.sh")
}