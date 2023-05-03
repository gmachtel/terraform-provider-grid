terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}

provider "grid" {
}
locals {
  solution_type = "Gaia-X Registry"
  name          = "mygxregistry"
}


resource "grid_network" "net1" {
  solution_type = local.solution_type
  name          = local.name
  nodes         = [8]
  ip_range      = "10.1.0.0/16"
  description   = "newer network"
  add_wg_access = true
}

# Deployment specs
resource "grid_deployment" "d1" {
  solution_type = local.solution_type
  name          = local.name
  node          = 8
  network_name  = grid_network.net1.name

  disks {
    name        = "data"
    size        = 10
    description = "volume holding docker data"
  }

  vms {
    name       = "gxregistry"
    flist      = "https://hub.grid.tf/geertmachtelinckx.3bot/registry.gitlab.com-gaia-x-lab-compliance-gx-registry-latest.flist"
    entrypoint = "/sbin/zinit init"
    publicip   = true
    planetary  = true
    cpu        = 1
    memory     = 1024

    mounts {
      disk_name   = "data"
      mount_point = "/var/lib/docker"
    }

    env_vars = {
      MONGO_HOST=cluster0.dpuol65.mongodb.net,
      MONGO_PORT=27017,
      NODE_ENV=production,
      PORT=3002,
      BASE_URI=http://localhost:3001,
      BASE_URL=http://localhost:3001,
      DB_USERNAME=mongoadmin,
      DB_PASSWORD=z4NHKfAmfVVzPVjc,
      MONGO_DATABASE=trust-anchor-registry
    }
  }
}




# Print deployment info
output "node1_zmachine1_ip" {
  value = grid_deployment.d1.vms[0].ip
}

output "computed_public_ip" {
  value = grid_deployment.d1.vms[0].computedip
}

output "ygg_ip" {
  value = grid_deployment.d1.vms[0].ygg_ip
}