resource "yandex_dns_zone" "mymeddataru" {
  description         = var.dns_zone.description
  folder_id           = var.folder_id
  labels              = {}
  name                = var.dns_zone.name
  public              = var.dns_zone.is_public
  zone                = var.dns_zone.zone
}

resource "yandex_dns_recordset" "mymeddataru_rs" {
  count = length(var.dns_zone.recordset)

  name    = var.dns_zone.recordset[count.index].name
  data    = var.dns_zone.recordset[count.index].data
  type    = var.dns_zone.recordset[count.index].type
  ttl     = var.dns_zone.recordset[count.index].ttl
  zone_id = yandex_dns_zone.mymeddataru.id
}
