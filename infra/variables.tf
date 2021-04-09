variable "DOCKER_IMG_TAG" {}

locals {
  repository = "hub.tisserv.net"

  base_name_api     = "ugb-api"
  base_name_storage = "ugb-storage"
  base_name_processor = "ugb-processor"
}
