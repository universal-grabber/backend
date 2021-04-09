variable "DOCKER_IMG_TAG" {}

variable "base_name" {
  type        = string
  description = "Base name"
  default     = "ugb-processor"
}


locals {
  repository    = "hub.tisserv.net"
  service_image = "${local.repository}/${var.base_name}:${var.DOCKER_IMG_TAG}"
}
