variable "DOCKER_IMG_TAG" {}

locals {
  repository = "hub.kube.tisserv.net"

  base_name_storage = "ugb-storage"
  base_name_processor = "ugb-processor"
  base_name_model-parser = "ugb-model-parser"
}
