workflow "Test && Publish" {
  on = "pull_request"
  resolves = ["Docker login", "Test"]
}

action "Test" {
  uses = "./actions/shared"
  args = "go test ."
}

action "Docker login" {
  uses = "actions/docker/login@master"
  secrets = ["secrets.$DOCKER_USER", "secrets.$DOCKER_PASS"]
  env = {
    DOCKER_REGISTRY_URL = "pile.mdk.zone"
  }
}
