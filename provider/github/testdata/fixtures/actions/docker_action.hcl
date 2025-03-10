action "hello" {
  filename    = "docker_action"
  name        = "Hello Docker Action"
  description = "A sample Docker action"

  input "who_to_greet" {
    description = "Who to greet"
    required    = true
    default     = "World"
  }

  runs {
    using = "docker"
    image = "Dockerfile"
    args  = ["--greeting", "hello"]
  }

  branding {
    icon  = "award"
    color = "green"
  }
}
