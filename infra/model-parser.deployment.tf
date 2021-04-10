resource "kubernetes_deployment" "ugb-model-parser" {
  metadata {
    name = local.base_name_model-parser

    labels = {
      app = local.base_name_model-parser
    }
  }
  spec {
    selector {
      match_labels = {
        app = local.base_name_model-parser
      }
    }

    template {
      metadata {
        name = local.base_name_model-parser

        labels = {
          app = local.base_name_model-parser
        }
      }
      spec {
        container {
          name = local.base_name_model-parser
          image = "${local.repository}/${local.base_name_model-parser}:${var.DOCKER_IMG_TAG}"
        }

        image_pull_secrets {
          name = "tisserv-hub"
        }
      }
    }
    replicas = 1
  }
}
