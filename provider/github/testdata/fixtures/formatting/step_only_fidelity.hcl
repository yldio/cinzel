step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4" # pinned by policy
  }
}

step "lint" {
  name = "Lint ✅"
  run  = <<-EOT
    echo one
    echo two
  EOT
}
