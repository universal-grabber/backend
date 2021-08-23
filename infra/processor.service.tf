resource "kubernetes_service" "ugb-processor" {
  metadata {
    name = local.base_name_processor

    labels = {
      app = local.base_name_processor
    }
  }
  spec {
    selector = {
      app = local.base_name_processor
    }

    type = "NodePort"

    port {
      name = "metrics"
      port = 1111
      target_port = 1111
      node_port = 30112
    }
  }
}
