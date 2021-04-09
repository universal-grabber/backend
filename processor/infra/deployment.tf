resource "kubernetes_deployment" "backend/processor" {
  metadata {
    name = var.base_name

    labels = {
      app = var.base_name
    }
  }
  spec {
    selector {
      match_labels = {
        app = var.base_name
      }
    }

    template {
      metadata {
        name = var.base_name

        labels = {
          app = var.base_name
        }
      }
      spec {
        container {
          name  = var.base_name
          image = local.service_image

          env {
            name  = "STORAGE_API"
            value = "https://10.0.1.77:30005"
          }

          env {
            name  = "MODEL_PROCESSOR_API"
            value = "https://10.0.1.77:30004"
          }

          env {
            name  = "BACKEND_API"
            value = "http://10.0.1.77:30003"
          }

          env {
            name  = "LOG_LEVEL"
            value = "6"
          }

          env {
            name  = "PARSE_MONGO_URI"
            value = "mongodb://10.0.1.77:27017"
          }

          env {
            name  = "ENABLED_TASKS"
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
