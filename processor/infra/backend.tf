terraform {
  backend "local" {
    path = "/var/tfstate/ugb-processor.tfstate"
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
  config_context = "minikube"
}
