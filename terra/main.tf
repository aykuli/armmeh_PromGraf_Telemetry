data "yandex_compute_image" "container-optimized-image" {
  family = var.vm.image_family
}

resource "yandex_compute_instance" "vms" {
  count = 1

  name        = var.vm.name
  hostname    = var.vm.hostname
  folder_id   = var.folder_id
  platform_id = var.vm.platform_id
  zone        = var.default_zone

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.container-optimized-image.id
      size     = var.vm.boot_disk_size
      type     = var.vm.boot_disk_type
    }
  }
  network_interface {
    security_group_ids = [yandex_vpc_security_group.aynur-monitoring-sg.id]
    subnet_id          = yandex_vpc_subnet.ayn-monitoring-subn.id
    nat                = var.vm.nat
  }

  resources {
    cores         = 2
    memory        = 4
    core_fraction = 100
  }
  scheduling_policy {
    preemptible = var.vm.preemptible
  }

  metadata = {
    user-data = templatefile("cloud-init.yml", {
      vm_user        = var.web_vm.user,
      ssh_public_key = var.web_vm.ssh_key,
      app_folder     = var.web_vm.app_folder,

      deploy_key     = indent(6, file(var.web_vm.deploy_key_path)),
      db_name = var.db_name,
      db_pwd  = yandex_lockbox_secret_version.v1.entries[0].text_value,
      db_user = var.db_user,
      db_port = var.db_port,
      db_host = yandex_mdb_postgresql_cluster.pg_cluster.host[0].fqdn
    })
  }
}

output "public_ip" {
  value = yandex_compute_instance.vms.0.network_interface[0].nat_ip_address
}