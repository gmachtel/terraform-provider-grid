terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}

provider "grid" {
  mnemonics = "aisle boring weekend scrub weapon vivid pass that amount negative entire usage" 
    network = "test"
}

locals {
  solution_type = "Gaia-X Registry"
  name          = "mygxregistry"
}


resource "grid_network" "net1" {
  solution_type = local.solution_type
  name          = "gxnetwork"
  nodes         = [1]
  ip_range      = "10.1.0.0/16"
  description   = "mygxnetwork"
  add_wg_access = true
}

# Deployment specs
resource "gx_registry" "d1" {
  solution_type = local.solution_type
  name          = "registry"
  node          = 1
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
      PORT=3000,
      BASE_URL=http://localhost:3000,
      DB_USERNAME=mongoadmin,
      DB_PASSWORD=z4NHKfAmfVVzPVjc,
      MONGO_DATABASE=trust-anchor-registry
    }
  }
}
resource "gx_compliance" "d2" {
  solution_type = local.solution_type
  name          = local.name
  node          = 1
  network_name  = test.grid.tf

  disks {
    name        = "data"
    size        = 10
    description = "volume holding docker data"
  }
  
  vms {
    name       = "gxcomplioance"
    flist      = "https://hub.grid.tf/geertmachtelinckx.3bot/registry.gitlab.com-gaia-x-lab-compliance-gx-compliance-latest.flist"
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
      X509_CERTIFICATE='-----BEGIN CERTIFICATE----- MIIDkDCCAngCCQDdxKvpzlMCHDANBgkqhkiG9w0BAQsFADCBiTELMAkGA1UEBhMCRVMxDzANBgNVBAgMBkJpbGJhbzEPMA0GA1UEBwwGQmlsYmFvMQ8wDQYDVQQKDAZH
YWlhLXgxDDAKBgNVBAsMA2N0bzEMMAoGA1UEAwwDY3RvMSswKQYJKoZIhvcNAQkB
Fhx2YWxlbnRpbi5taXNpYXN6ZWtAZ2FpYS14LmV1MB4XDTIzMDUwMzEyNDIzOFoX
DTI0MDUwMjEyNDIzOFowgYkxCzAJBgNVBAYTAkVTMQ8wDQYDVQQIDAZCaWxiYW8x
DzANBgNVBAcMBkJpbGJhbzEPMA0GA1UECgwGR2FpYS14MQwwCgYDVQQLDANjdG8x
DDAKBgNVBAMMA2N0bzErMCkGCSqGSIb3DQEJARYcdmFsZW50aW4ubWlzaWFzemVr
QGdhaWEteC5ldTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKlxFWSp
tRmc9xl0kEZa0RfPyOaIWdKh3n1DrLNOQkiqvg1o2LU7Q9ztzTHSpwyFCwf0J7f0
lS0LDWJRZg5OrGVKW6f6O4Jkv862ffs5NvTSgUJb8nNO/lz7BMhzyb1PKOh3NG7I
HcMwXLSYILqrVI/JC046QUyO2+e3fHEA9nj+yiyzWeYaE4FUvycdLZCCjEJDVk3D
/pGW5HIHavM3LYB+5DayxtEb/u/QpBHA2X62aNp+kpVxU7ZPzKn+/EqXxNQm9lpd
fIbnd9u5qjHNGdst+SZmK8EGNdnuZSaL32mJGQK+UfVnbcz/k+XgWEks+2si6TJI
kfTbrHnNqdgIxWUCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAco2ghUiyPulgv37B
7DHN1VUYl6Z6y5jkmgPs4JBYHPks9gcrU5NRcDhVkwKaYsmgcuzLyz6wwCGKjZLm
f1UTXA1lSUDRY1SD+Hms2VAHchNUP9OQW8PH8eR900EsvkRkDP9UDiY6wqZWIloX
FhjYhBULn5gpLO9O4yj6zDf/TT4yFU2UaquiLEn4a+Vw/RMtn4AK59aHht3ZAPG0
wd3HIKaI0VVAuFSeevhD6UnheLOT178Wd9dO+4tJBIy1/QlTB/ftipuZw5Js4kNF
1i1e3HPPXYynNUqa7PxnW5594hNXVqHY0xVynZJC9GvjXFGB2ZlKg/CTb2vjvCrJ
lF9efw== -----END CERTIFICATE-----', 
      privateKey='-----BEGIN PRIVATE KEY----- MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCpcRVkqbUZnPcZdJBGWtEXz8jmiFnSod59Q6yzTkJIqr4NaNi1O0Pc7c0x0qcMhQsH9Ce39JUtCw1i
UWYOTqxlSlun+juCZL/Otn37OTb00oFCW/JzTv5c+wTIc8m9TyjodzRuyB3DMFy0
mCC6q1SPyQtOOkFMjtvnt3xxAPZ4/soss1nmGhOBVL8nHS2QgoxCQ1ZNw/6RluRy
B2rzNy2AfuQ2ssbRG/7v0KQRwNl+tmjafpKVcVO2T8yp/vxKl8TUJvZaXXyG53fb
uaoxzRnbLfkmZivBBjXZ7mUmi99piRkCvlH1Z23M/5Pl4FhJLPtrIukySJH026x5
zanYCMVlAgMBAAECggEASSeoq+BVbyyExrm7vJRjKBuuylFeLoFydLSuMG/+UC9f
hJyay4w93XnSGMuxEcezHoj8SQDREzRtX+By5oRzC/xRnDF+Veq3oUDLHZbzMjpc
UlEuWThmu7AovX64QAKYT91/hsDhkK8lp1by8oUcKkQLnIesP6iCKwXeNU/MGip7
Gi9qTL1VAoq01s19WItjqQeN4eDrQS7LnH34c5TnX0wTrAyfbYWMKqt5e+eTnjMz
ATyAmtyEcTG9OIBSURcEkVf3tZf/MOehLTB3VykBBRLBDvBLgTxxHFLOH6HokcEB
hWEiSVZaz75Lae2Tn9zw3PTtgrD1IbqafJZYia7rlQKBgQDYJiTXg09xDQqx/pPE
AK3J+odBcM2mmbZn7UtL3GGxr2XQ2WHuqxdLZUdgGV47xoGzdupjckaJLZDK7Qsd
RR6TZIBC+RjgRYKRh8t6g0D4Zbf6Bz7MtgSfh0nm4EsNKPsPM52mMuBvmMUIRkzT
11d5IUB/JYkh/qB4hFvW79fzfwKBgQDIrnAkMx1G6+Z2NGUmAzwbZ1hHaGVIhfDa
8eyd4qmhc+Ut5AfquMdVsGmUY+7GmSysKBPh6dKhFLpaRBvc6NpgMxdXj3xIJDMJ
u4TU+9G7VF35Y2oMLk0V8xCt/LbtDdRDRopEipSipF82sS7N43mY9nDfXROeAfDw
WEW5nB1pGwKBgFqEVsdduQX51THHD7B0LFM6NCwmhiivlGCfUfn513umnTfB4S4m
vgGX7zUCSqEztxslwHBW6c0GpO0mKZJfYLeOzdu+HznEQjKZsI+kqa+cQVhWcmke
DH+WE0+5PBTNxZ2PGwT/p1d0nYaikgfg1UPnte/JE02GKUNvAlrXZBMXAoGBAKfr
fp5rvtW3UrIaCYETN9peUTn/GrDikrVBtZIvNW2Jgn0xz4YSc4k6Aj5OmF/Jj93F
800X0E1FAOHDF+VzWjcgySlVQNNEpwg/xlhJFie/4pppGzVyEMKLDqvnSFF3PuPE
RLIxm0m0bI9hFx7kdr0NiUj9owqV6TvTQRSckud9AoGBAMzlIGTvfjiY7jfLI/LN
DtyNO/rRBtU9rZ7AAsYkXsckXT0bFR/OtKvJXBb0TROSqfeTBopoad3j1KPXI7qh
b7NgkUwavdjmOqQ1J/n9QUHh+yXZyfFkM/IkXREz8aist4afvcVYeEvc1ZiYQvOs
vAvU2uvodx8bye1cPk8s+O6E -----END PRIVATE KEY-----',
      REGISTRY_URL='http://localhost:3000/',
      BASE_URL='http://localhost:3001/',
      BASE_URI='http://localhost:3001/',
      APP_PATH=''
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
