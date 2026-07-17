resource "yandex_dns_zone" "mymeddataru" {
  folder_id           = var.folder_id
  description         = "для целей задания юзаю какой-то своё доменное имя, оно ничего не означает"
  labels              = {}
  name                = "mymeddataru"
  public              = true
  zone                = "mymeddata.ru."
}
resource "yandex_dns_recordset" "mymeddataru_a_record" {
  zone_id = yandex_dns_zone.mymeddataru.id
  name    = "mymeddata.ru."
  type    = "SOA"
  ttl     = 6600
  data = ["ns1.yandexcloud.net. mx.cloud.yandex.net. 1 10800 900 604800 900"]
}
resource "yandex_dns_recordset" "mymeddataru_a_record" {
  zone_id = yandex_dns_zone.mymeddataru.id
  name    = "mymeddata.ru."
  type    = "NS"
  ttl     = 3600
  
  data = ["ns1.yandexcloud.net.", "ns2.yandexcloud.net."]
}
resource "yandex_dns_recordset" "mymeddataru_a_record" {
  zone_id = yandex_dns_zone.mymeddataru.id
  name    = "mymeddata.ru."
  type    = "A"
  ttl     = 600
  
  data    = [yandex_vpc_address.addr.external_ipv4_address[0].address]
}
