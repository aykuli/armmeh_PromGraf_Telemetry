data "yandex_compute_image" "container-optimized-image" {
  family = "container-optimized-image"
}

resource "yandex_compute_instance" "vm" {
  folder_id   = var.folder_id
  name        = "aynurvm"
  hostname    = "aynurhost"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.container-optimized-image.id
      size     = 40
      type     = "network-ssd"
    }
  }
  network_interface {
    security_group_ids = [yandex_vpc_security_group.aynur-monitoring-sg.id]
    subnet_id          = yandex_vpc_subnet.ayn-monitoring-subn.id
    nat                = true
    nat_ip_address     = yandex_vpc_address.addr.external_ipv4_address[0].address
  }

  resources {
    cores         = 2
    memory        = 4
    core_fraction = 100
  }
  scheduling_policy {
    preemptible = true
  }

  metadata = {
    ssh-keys = "${var.github_user}:${var.ssh_public_key}"

    user-data = templatefile("cloud-init.yml", {
      github_user  = var.github_user,
      github_token = var.github_token,
      github_repo  = var.github_repo,

      tg_bot_token        = var.tg_bot_token,
      tg_chat_id          = var.tg_chat_id,
      mqtt_broker_url     = var.mqtt_broker_url,
      mqtt_user           = var.mqtt_user,
      mqtt_pass           = var.mqtt_pass,
      fleet_backend_url   = var.fleet_backend_url,
      grafana_admin_pass  = var.grafana_admin_pass,
      grafana_admin_name  = var.grafana_admin_name,
      grafana_admin_email = var.grafana_admin_email,
      postgres_uri        = var.postgres_uri,
      postgres_db         = var.postgres_db,
      postgres_user       = var.postgres_user,
      postgres_pass       = var.postgres_pass,
      htpasswd_content    = var.nginx_htpasswd_content
    })
  }
}

output "public_ip" {
  value = yandex_compute_instance.vms.0.network_interface[0].nat_ip_address
}