action "greet" {
  filename    = "node_action"
  name        = "Greet Action"
  description = "A sample Node.js action"

  input "who_to_greet" {
    description = "Who to greet"
    required    = true
    default     = "World"
  }

  output "time" {
    description = "The time we greeted you"
  }

  runs {
    using = "node20"
    main  = "index.js"
    pre   = "setup.js"
    post  = "cleanup.js"
  }
}
