resource "kubernetes_service" "ugb-model-parser" {
  metadata {
    name = local.base_name_model-parser

    labels = {
      app = local.base_name_model-parser
    }
  }
  spec {
    selector = {
      app = local.base_name_model-parser
    }

    type = "NodePort"

    port {
      name = "http"
      port = 443
      target_port = 8443
    }

    port {
      name = "metrics"
      port = 1111
      target_port = 1111
      node_port = 30111
    }
  }
}
