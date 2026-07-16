# --- PROVIDERS ---
variable "cloud_id" {
  type = string
}
variable "folder_id" {
  type = string
}
variable "keys_path" {
  type = string
}
variable "default_zone" {
  type    = string
  default = "ru-central1-a"
}
# Zone	Subnet Name	CIDR Block (Example)	Purpose
# ru-central1-a	subnet-a	10.0.1.0/24	Resources in Zone A
# ru-central1-b	subnet-b	10.0.2.0/24	Resources in Zone B
# ru-central1-c	subnet-c	10.0.3.0/24	Resources in Zone C
variable "default_zone" {
  type    = string
  default = "ru-central1-a"
}

# --- VMs ---
variable "vm" {
  type = object({
    name            = string
    hostname        = string
    user            = string
    ssh_key         = string
    app_folder      = string
    deploy_key_path = string

    image_family   = string
    platform_id    = string
    boot_disk_type = string
    boot_disk_size = number
    preemptible    = bool
    nat            = bool
    resources      = object({
      cores          = number
      memory        = number
      core_fraction = number
    })
  })
}
# ---

# --- VPC ---
variable "static_ip" {
  type = object({
    name = string
  })
}
variable "vpc" {
  type = object({
    network_name   = string
    subnet_name    = string
    v4_cidr_blocks = list(string)
  })
  default = {
    network_name   = "ayn-netw"
    subnet_name    = "ayn-subn"
    v4_cidr_blocks = ["10.0.1.0/16"]
  }
}
variable "sg" {
  type = object({
    web = object({
      name        = string
      description = string
      labels      = map(string)
      ingress_rules = list(object({
        protocol    = string
        v4_cidr     = optional(list(string))
        port        = optional(number)
        from_port   = optional(number)
        to_port     = optional(number)
        description = optional(string)
      }))
      egress_rules = optional(list(object({
        protocol    = string
        v4_cidr     = optional(list(string))
        port        = optional(number)
        from_port   = optional(number)
        to_port     = optional(number)
        description = optional(string)
      })))
    })
  })
}

# --- DNS ---
variable "dns_zone" {
  type = object({
    name        = string
    zone        = string
    is_public   = bool
    description = optional(string)
    recordset = list(object({
      name = string
      data = list(string)
      type = string
      ttl  = number
    }))
  })
}
