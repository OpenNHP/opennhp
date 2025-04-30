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


### regen certs
```
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
  -keyout server.key -out server.crt -subj "/CN=opennhp.cn" \
  -addext "subjectAltName=DNS:opennhp.cn,IP:127.0.0.1"
```