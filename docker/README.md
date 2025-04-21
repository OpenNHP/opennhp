# Docker debug env

## build base image

```shell
cd ./docker
docker build -t opennhp-base:latest -f Dockerfile.base ../..
```

## configration

- ./docker/nhp-server/plugins/example/etc/resource.toml 
```
# The value of Addr.Ip can be obtained using the following command
# Addr.Ip = ""

docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nhp-ac

```

## start docker env
```shell
docker compose up -d
```

## Debug
http://localhost:62206/plugins/example?resid=demo&action=login


### vaildate the ipset rule
```
docker exec -it nhp-ac ipset list
```
