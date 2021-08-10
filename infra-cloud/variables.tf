variable "DOCKER_IMG_TAG" {}

locals {
  repository = "hub.kube.tisserv.net"

  base_name_api = "ugb-api"
}
