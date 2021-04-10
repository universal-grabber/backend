terraform {
  backend "local" {
    path = "/var/tfstate/ugb.tfstate"
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
  config_context = "minikube"
}
