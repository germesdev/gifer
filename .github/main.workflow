workflow "Test && Publish" {
  on = "pull_request"
  resolves = ["Docker login"]
}

action "Test" {
  uses = "./actions/test"
  args = "go test ."
}

action "Docker login" {
  uses = "fishbullet/golang-actions@master"
  env = {
    MY_NAME = "Bobby"
  }
  args = "go version"
}
