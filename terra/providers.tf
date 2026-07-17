terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "> 0.9"
    }
  }
  backend "s3" {
    endpoint = "https://storage.yandexcloud.net"
    bucket = var.s3bucket
    key    = "terraform.tfstate"
    region = "ru-central1"

    # Специфичные настройки для совместимости Yandex Cloud с протоколом S3
    skip_region_validation      = true
    skip_credentials_validation = true
    skip_requesting_account_id  = true
    skip_s3_express_support     = true
  }
  required_version = ">= 1.3.0"
}

provider "yandex" {
  cloud_id                 = var.cloud_id
  folder_id                = var.folder_id
  zone                     = "ru-central1-a"
  service_account_key_file = "key.json"
}

