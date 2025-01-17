variable "public_key" {
  type = string
}

terraform {
  required_providers {
    grid = {
      source  = "threefoldtechdev.com/providers/grid"
      version = "0.2"
    }
  }
}

provider "grid" {
  network = "dev"
}

# resource "grid_scheduler" "sched" {
#   requests {
#     name = "node1"
#     cru  = 2
#     sru  = 512
#     mru  = 1024
#     public_config = true
#   }
# }

# this data source is used to break circular dependency in cases similar to the following:
# vm: needs to know the domain in its init script
# gateway_name: needs the ip of the vm to use as backend.
# - the fqdn can be computed from grid_gateway_domain for the vm
# - the backend can reference the vm ip directly 
data "grid_gateway_domain" "domain" {
  node = 14
  name = "examp123456"
}

locals {
  name = "vmtesting"
}


resource "grid_network" "net1" {
  nodes       = [14]
  ip_range    = "10.1.0.0/16"
  name        = local.name
  description = "newer network"
}
resource "grid_deployment" "d1" {
  name         = local.name
  node         = 14
  network_name = grid_network.net1.name
  vms {
    name       = "vm1"
    flist      = "https://hub.grid.tf/tf-official-apps/base:latest.flist"
    cpu        = 2
    memory     = 1024
    entrypoint = "/sbin/zinit init"
    env_vars = {
      SSH_KEY = "${var.public_key}"
    }
    planetary = true
  }
}

locals {
  ygg_ip = try(length(grid_deployment.d1.vms[0].ygg_ip), 0) > 0 ? grid_deployment.d1.vms[0].ygg_ip : ""
}

resource "grid_name_proxy" "p1" {
  node            = 14
  name            = "examp123456"
  backends        = [format("http://%s:9000", local.ygg_ip)]
  network         = grid_network.net1.name
  tls_passthrough = false
}

output "fqdn" {
  value = data.grid_gateway_domain.domain.fqdn
}

output "ygg_ip" {
  value = grid_deployment.d1.vms[0].ygg_ip
}
