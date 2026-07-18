resource "yandex_dns_zone" "mymeddataru" {
  folder_id           = var.folder_id
  description         = "для целей задания юзаю какой-то своё доменное имя, оно ничего не означает"
  labels              = {}
  name                = "mymeddataru"
  public              = true
  zone                = "mymeddata.ru."
}
resource "yandex_dns_recordset" "mymeddataru_a" {
  depends_on = [
    yandex_compute_instance.vm,
    yandex_vpc_address.addr
  ]
  zone_id = yandex_dns_zone.mymeddataru.id
  name    = "mymeddata.ru."
  type    = "A"
  ttl     = 600
  
  data    = [yandex_vpc_address.addr.external_ipv4_address[0].address]
}
