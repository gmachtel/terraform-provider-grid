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
  name = "myvm"
}
resource "grid_network" "net1" {
  nodes         = [1]
  ip_range      = "10.1.0.0/16"
  name          = local.name
  description   = "newer network"
  add_wg_access = true
}

resource "grid_deployment" "swarm1" {
  name         = local.name
  node         = 1
  network_name = grid_network.net1.name
  vms {
    name        = "swarmManager1"
    flist       = "https://hub.grid.tf/tf-official-apps/grid3_ubuntu20.04_debug-latest.flist"
    entrypoint  = "/init.sh"
    cpu         = 2
    memory      = 1024
    rootfs_size = 25000
    env_vars = {
      SSH_KEY = file("~/.ssh/id_rsa.pub")
    }
    planetary = true
  }
  vms {
    name        = "swarmWorker1"
    flist       = "https://hub.grid.tf/tf-official-apps/grid3_ubuntu20.04_debug-latest.flist"
    entrypoint  = "/init.sh"
    cpu         = 2
    memory      = 1024
    rootfs_size = 25000
    env_vars = {
      SSH_KEY = file("~/.ssh/id_rsa.pub")
    }
    planetary = true
  }

  provisioner "remote-exec" {
    inline = [
      "curl -fsSL https://get.docker.com/ | sh",
      "setsid /usr/bin/containerd &",
      "setsid /usr/bin/dockerd -H unix:// --containerd=/run/containerd/containerd.sock &",
      "sleep 10",
      "docker swarm init --advertise-addr ${grid_deployment.swarm1.vms[0].ygg_ip}",
      "docker swarm join-token --quiet worker > /root/token",
    ]
    connection {
      type    = "ssh"
      user    = "root"
      agent   = true
      host    = grid_deployment.swarm1.vms[0].ygg_ip
      timeout = "20s"
    }
  }

  provisioner "local-exec" {
    command = "scp -3 -o StrictHostKeyChecking=no -o NoHostAuthenticationForLocalhost=yes -o UserKnownHostsFile=/dev/null root@[${grid_deployment.swarm1.vms[0].ygg_ip}]:/root/token root@[${grid_deployment.swarm1.vms[1].ygg_ip}]:/root/token"
  }

  provisioner "remote-exec" {
    inline = [
      "curl -fsSL https://get.docker.com/ | sh",
      "setsid /usr/bin/containerd &",
      "setsid /usr/bin/dockerd -H unix:// --containerd=/run/containerd/containerd.sock &",
      "sleep 10",
      "docker swarm join --token $(cat /root/token) [${grid_deployment.swarm1.vms[0].ygg_ip}]:2377"
    ]
    connection {
      type    = "ssh"
      user    = "root"
      agent   = true
      host    = grid_deployment.swarm1.vms[1].ygg_ip
      timeout = "20s"
    }
  }
}



output "node1_zmachine1_ip" {
  value = grid_deployment.swarm1.vms[0].ip
}

output "node1_zmachine2_ip" {
  value = grid_deployment.swarm1.vms[1].ip
}

output "node1_zmachine1_ygg_ip" {
  value = grid_deployment.swarm1.vms[0].ygg_ip
}

output "node1_zmachine2_ygg_ip" {
  value = grid_deployment.swarm1.vms[1].ygg_ip
}

