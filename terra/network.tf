resource "yandex_vpc_address" "addr" {
  name = "aynip"
  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}

resource "yandex_vpc_network" "ayn-monitoring-netw" {
  folder_id = var.folder_id
  name      = "ayn-net"
}

resource "yandex_vpc_subnet" "ayn-monitoring-subn" {
  network_id       = yandex_vpc_network.ayn-monitoring-netw.id
  name             = "ayn-subn"
  v4_cidr_blocks   = [ "10.0.2.0/24" ]
}

resource "yandex_vpc_security_group" "aynur-monitoring-sg" {
  name        = "ayn-sg"
  description ="Allow HTTP, HTTPS and SSH"
  labels      = { forwhom: "armmeh" }

  network_id  = yandex_vpc_network.ayn-monitoring-netw.id
  ingress {
    protocol    = "TCP"
    v4_cidr_blocks     = ["0.0.0.0/0"]
    port        = 22
    description = "Allow SSH"
  }
  ingress {
    protocol    = "TCP"
    v4_cidr_blocks     = ["0.0.0.0/0"]
    port        = 80
    description = "Allow HTTP"
  }
  ingress {
    protocol    = "TCP"
    v4_cidr_blocks     = ["0.0.0.0/0"]
    port        = 443
    description = "Allow HTTPS"
  }
  
  egress {
    protocol    = "ANY"
    v4_cidr     = ["0.0.0.0/0"]
    description = "Permit ANY"
  }
}

