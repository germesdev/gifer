workflow "Test && Build" {
  on = "pull_request"
  resolves = ["Test", "Docker login"]
}

action "Test" {
  uses = "./actions/shared"
  args = "go test ."
}

action "Docker login" {
  uses = "actions/docker/login@master"
  secrets = ["DOCKER_USER", "DOCKER_PASS"]
  env = {
    DOCKER_REGISTRY_URL = "pile.mdk.zone"
  }
}
