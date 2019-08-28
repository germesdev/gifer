workflow "Build" {
  on = "pull_request"
  resolves = ["Docker login"]
}

action "Docker login" {
  uses = "./docker-login"
  env = {
    MY_NAME = "Bobby"
  }
  args = "go version"
}
