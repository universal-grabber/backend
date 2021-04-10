resource "kubernetes_deployment" "ugb-processor" {
  metadata {
    name = local.base_name_processor

    labels = {
      app = local.base_name_processor
    }
  }
  spec {
    selector {
      match_labels = {
        app = local.base_name_processor
      }
    }

    template {
      metadata {
        name = local.base_name_processor

        labels = {
          app = local.base_name_processor
        }
      }
      spec {
        container {
          name = local.base_name_processor
          image = "${local.repository}/${local.base_name_processor}:${var.DOCKER_IMG_TAG}"

          env {
            name = "STORAGE_API"
            value = "https://ugb-storage"
          }

          env {
            name = "MODEL_PROCESSOR_API"
            value = "https://ugb-model-parser"
          }

          env {
            name = "BACKEND_API"
            value = "http://ugb-api"
          }

          env {
            name = "BACKEND_GRPC_API"
            value = "ugb-api:6565"
          }

          env {
            name = "LOG_LEVEL"
            value = "6"
          }

          env {
            name = "PARSE_MONGO_URI"
            value = "mongodb://10.0.1.77:27017"
          }

          env {
            name = "ENABLED_TASKS"
            value = "DOWNLOAD,DEEP_SCAN,PARSE,PUBLISH"
          }
        }

        image_pull_secrets {
          name = "tisserv-hub"
        }
      }
    }
    replicas = 1
  }
}
