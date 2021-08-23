resource "kubernetes_service" "ugb-api" {
  metadata {
    name = local.base_name_api

    labels = {
      app = local.base_name_api
    }

    namespace = local.base_namespace
  }
  spec {
    selector = {
      app = local.base_name_api
    }

    type = "NodePort"

    port {
      name = "http"
      port = 80
      node_port = 30003
      target_port = 8080
    }

    port {
      name = "grpc"
      port = 6565
      node_port = 30004
      target_port = 6565
    }

    port {
      name = "metrics"
      port = 1111
      target_port = 1111
      node_port = 30113
    }
  }
}
