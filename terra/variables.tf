variable "cloud_id" { type = string }
variable "folder_id" { type = string }
# Zone	Subnet Name	CIDR Block (Example)	Purpose
# ru-central1-a	subnet-a	10.0.1.0/24	Resources in Zone A
# ru-central1-b	subnet-b	10.0.2.0/24	Resources in Zone B
# ru-central1-c	subnet-c	10.0.3.0/24	Resources in Zone C

variable "github_user" { type = string }
variable "github_token" { type = string }
variable "github_repo" { type = string }
variable "ssh_public_key" { type = string }

# Переменные для генерации .env на сервере
variable "tg_bot_token" { type = string }
variable "tg_chat_id" { type = string }
variable "mqtt_broker_url" { type = string }
variable "mqtt_user" { type = string }
variable "mqtt_pass" { type = string }
variable "fleet_backend_url" { type = string }
variable "grafana_admin_pass" { type = string }
variable "grafana_admin_name" { type = string }
variable "grafana_admin_email" { type = string }
variable "postgres_uri" { type = string }
variable "postgres_db" { type = string }
variable "postgres_user" { type = string }
variable "postgres_pass" { type = string }
variable "nginx_htpasswd_content" { type = string }
variable "mqqt_broker_passwd_content" { type = string }
