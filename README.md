# Restic Controller

[![Tests, build and publish](https://github.com/tcardonne/restic-controller/workflows/Tests,%20build%20and%20publish/badge.svg)](https://github.com/tcardonne/restic-controller/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/tcardonne/restic-controller)](https://goreportcard.com/report/github.com/tcardonne/restic-controller)
[![Docker Pulls](https://img.shields.io/docker/pulls/tcardonne/restic-controller)](https://hub.docker.com/r/tcardonne/restic-controller)

## Introduction

Restic Controller is a program that helps to monitor and manage [restic](https://github.com/restic/restic) backup repositories.

This project has three components :
- a Prometheus exporter, allowing to scrape metrics via Prometheus,
- an integrity controller which will check repositories' integrity on a schedule,
- a retention controller which will apply a given retention policy on a schedule.

With this project, you can use a central location to monitor and manage your repositories.


## Get started

Restic Controller is available as a Docker image : [tcardonne/restic-controller](https://hub.docker.com/repository/docker/tcardonne/restic-controller).

Basic usage :
```bash
docker run --rm -it \
    -v "$PWD/config.yml:/app/config.yml" \
    -p "8080:8080" \
    tcardonne/restic-controller
```

Or, with `docker-compose.yml` :
```yaml
version: '3.7'

services:
  controller:
    image: tcardonne/restic-controller:latest
    ports:
      - "8080:8080"
    volumes:
      - "./config.yml:/app/config.yml"
```


Once started, the controller will run scheduled integrity checks and apply retention policies. Exported metrics are available on the `/metrics` path.

## Configuration

Restic-Controller will, by default, look for a `config.yml` file in the current working directory.

Full configuration reference available in the `config.example.yml` file.

### Example :

```yaml
exporter:
  bind_address: ":8080"

repositories:
    - name: "backtothefuture"
      url: "rest:https://user:password@repositories.restic.example/backtothefuture"
      # envFromFile:
      #   RESTIC_REST_USERNAME: /etc/secrets/backtothefuture/username
      #   RESTIC_REST_PASSWORD: /etc/secrets/backtothefuture/password
      password: "password"
      # passwordFile: "/etc/secrets/backtothefuture-repository"
      check:
        schedule: "0 3 * * *"
      retention:
        schedule: "0 4 * * *"
        policy:
          keep_last: 15
```

This configuration will run integrity checks everyday at 3AM and apply the retention policy at 4AM. The policy will keep the last 15 snapshots.

## Grafana & Alertmanager

A Grafana dashboard and a sample Alertmanager rules file are available in the `examples/` directory.

![Grafana example](examples/grafana.png)


## License

This project is licensed under [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause). You can find the complete text in the file `LICENSE`.