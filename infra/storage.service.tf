resource "kubernetes_service" "ugb-storage" {
  metadata {
    name = local.base_name_storage

    labels = {
      app = local.base_name_storage
    }
  }
  spec {
    selector = {
      app = local.base_name_storage
    }

    type = "NodePort"

    port {
      name = "http"
      port = 443
      target_port = 8443
    }

    port {
      name = "grpc"
      port = 6565
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
