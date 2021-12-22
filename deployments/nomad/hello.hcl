job "hello" {
  type        = "service"
  datacenters = ["dc1"]

  constraint {
    operator  = "distinct_property"
    attribute = "${attr.unique.network.ip-address}"
    value     = "2"
  }

  spread {
    attribute = "${node.unique.id}"
  }

  update {
    max_parallel = 1

    auto_revert = true
    canary      = "1"
  }

  group "hello" {
    count          = "1"
    shutdown_delay = "10s"

    restart {
      interval = "1m"
      mode     = "delay"
    }

    task "hello" {
      driver = "docker"
      config {
        image = "docker.io/sashayakovtseva/hello-web:v0.1.0"
        ports = ["grpc", "http"]
      }

      env {
        GRPC_PORT = "50051"
        HTTP_PORT = "6060"
      }

      vault {
        policies    = ["reader"]
        change_mode = "noop"
        env         = false
      }

      template {
        data = <<EOH
          {{with secret "secret/hello"}}
          DEFAULT_FIRST_NAME="{{.Data.data.default_first_name}}"
          {{end}}
          EOH

        destination = "secrets/file.env"
        change_mode = "restart"
        env         = true
      }

      template {
        data = <<EOH
           {{ key "hello" }}
        EOH

        destination   = "local/default_last_name.json"
        change_mode   = "signal"
        change_signal = "SIGHUP"
      }

      service {
        name         = "grpc-hello"
        port         = "grpc"
        address_mode = "host"

        tags = [
          "green",
          "v0.1.0"
        ]

        canary_tags = [
          "blue",
          "v0.1.0"
        ]

        meta {
          protocol = "grpc"
          dc       = "${attr.consul.datacenter}"
        }

        check {
          type         = "grpc"
          port         = "grpc"
          interval     = "5s"
          timeout      = "1s"
          grpc_service = "sashayakovtseva.hello.v1.HelloService"
        }
      }


      service {
        name         = "http-hello"
        port         = "http"
        address_mode = "host"

        tags = [
          "green",
          "v0.1.0"
        ]

        canary_tags = [
          "blue",
          "v0.1.0"
        ]

        meta {
          protocol = "http"
          dc       = "${attr.consul.datacenter}"
        }

        check {
          type     = "http"
          port     = "http"
          interval = "5s"
          timeout  = "1s"
          path     = "/check"
        }
      }

      resources {
        cpu    = 300
        memory = 100
      }
    }

    network {
      port "grpc" {
        to = 50051
      }
      port "http" {
        to = 6060
      }
    }
  }
}
