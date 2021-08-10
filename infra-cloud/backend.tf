terraform {
  backend "local" {
    path = "/var/tfstate/ugb-cloud.tfstate"
  }
}

provider "kubernetes" {
  config_path = "~/.kube/kube.tisserv.net.config"
  config_context = "tisserv-kube"
}
