resource "kubernetes_deployment" "ugb-storage" {
  metadata {
    name = local.base_name_storage

    labels = {
      app = local.base_name_storage
    }
  }
  spec {
    selector {
      match_labels = {
        app = local.base_name_storage
      }
    }

    template {
      metadata {
        name = local.base_name_storage

        labels = {
          app = local.base_name_storage
        }
      }
      spec {
        container {
          name  = local.base_name_storage
          image = "${local.repository}/${local.base_name_storage}:${var.DOCKER_IMG_TAG}"
        }

        image_pull_secrets {
          name = "tisserv-hub"
        }
      }
    }
    replicas = 1
  }
}
