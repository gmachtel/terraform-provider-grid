
variable "public_key" {
  type = string
}

variable "presearch_regestration_code" {
  type = string
}

terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}

provider "grid" {
}

resource "grid_network" "net1" {
  nodes         = [33]
  ip_range      = "10.1.0.0/16"
  name          = "network"
  description   = "newer network"
}

# Deployment specs
resource "grid_deployment" "d1" {
  node         = 33
  network_name = grid_network.net1.name

  disks {
    name        = "data"
    size        = 5
    description = "volume holding docker data"
  }

  vms {
    name       = "presearch"
    flist      = "https://hub.grid.tf/tf-official-apps/presearch-v2.2.flist"
    entrypoint = "/sbin/zinit init"
    planetary  = true
    cpu        = 1
    memory     = 1024
    mounts {
      disk_name   = "data"
      mount_point = "/var/lib/docker"
    }
    env_vars = {
      SSH_KEY                     = "${var.public_key}",
      PRESEARCH_REGISTRATION_CODE = "e5083a8d0a6362c6cf7a3078bfac81e3",

    }
  }
}

output "ygg_ip" {
  value = grid_deployment.d1.vms[0].ygg_ip
}