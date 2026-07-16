terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "> 0.9"
    }
  }
  required_version = ">= 1.3.0"
}

provider "yandex" {
  cloud_id                 = var.cloud_id
  zone                     = var.default_zone
  folder_id                = var.folder_id
  service_account_key_file = file(var.keys_path)
}

