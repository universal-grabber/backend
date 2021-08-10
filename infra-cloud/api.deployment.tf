resource "kubernetes_deployment" "ugb-api" {
  metadata {
    name = local.base_name_api

    labels = {
      app = local.base_name_api
    }

    namespace = local.base_namespace
  }
  spec {
    selector {
      match_labels = {
        app = local.base_name_api
      }
    }

    template {
      metadata {
        name = local.base_name_api

        labels = {
          app = local.base_name_api
        }

        namespace = local.base_namespace
      }
      spec {
        container {
          name = local.base_name_api
          image = "${local.repository}/${local.base_name_api}:${var.DOCKER_IMG_TAG}"
        }

        image_pull_secrets {
          name = "tisserv-hub"
        }
      }
    }
    replicas = 1
  }
}
