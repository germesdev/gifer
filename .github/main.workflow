workflow "Build" {
  on = "pull_request"
  resolves = ["Docker login"]
}

action "Docker login" {
  uses = "fishbullet/golang-actions@master"
  env = {
    MY_NAME = "Bobby"
  }
  args = "go version"
}
