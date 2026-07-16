resource "yandex_vpc_address" "addr" {
  name = var.static_ip.name
  external_ipv4_address {
    zone_id = var.default_zone
  }
}

resource "yandex_vpc_network" "ayn-monitoring-netw" {
  name      = var.vpc.network_name
  folder_id = var.folder_id
}

resource "yandex_vpc_subnet" "ayn-monitoring-subn" {
  name             = var.vpc.subnet_name
  network_id       = yandex_vpc_network.ayn-monitoring-netw.id
  v4_cidr_blocks   = var.vpc.v4_cidr_blocks
}

resource "yandex_vpc_security_group" "aynur-monitoring-sg" {
  name        = var.sg.web.name
  description = var.sg.web.description
  labels      = { forwhom: "armmeh" }

  network_id  = yandex_vpc_network.ayn-monitoring-netw.id
  dynamic "ingress" {
    for_each = var.sg.web.ingress_rules

    content {
      protocol       = lookup(ingress.value, "protocol", null)
      description    = lookup(ingress.value, "description", null)
      port           = lookup(ingress.value, "port", null)
      v4_cidr_blocks = lookup(ingress.value, "v4_cidr", null)
    }
  }

  dynamic "egress" {
    for_each = var.sg.web.egress_rules

    content {
      protocol       = lookup(egress.value, "protocol", null)
      description    = lookup(egress.value, "description", null)
      port           = lookup(egress.value, "port", null)
      v4_cidr_blocks = lookup(egress.value, "v4_cidr", null)
    }
  }
}

