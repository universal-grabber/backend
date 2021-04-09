variable "DOCKER_IMG_TAG" {}

locals {
  repository = "hub.tisserv.net"

  base_name_api     = "api"
  base_name_storage = "storage"
}
