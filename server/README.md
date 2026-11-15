# server

## Install Jzero Framework

```shell
go install github.com/jzero-io/jzero/cmd/jzero@latest

jzero check
```

## Run locally

```shell
# sqlite (default, zero config)
go run main.go server

# mysql (overrides via env vars, see etc/etc.yaml for all placeholders)
DRIVER_NAME=mysql \
DATASOURCE='root:123456@tcp(127.0.0.1:3306)/jzero-admin?charset=utf8mb4&parseTime=True&loc=Local' \
REDIS_MINI_REDIS=false \
REDIS_HOST=127.0.0.1:6379 \
go run main.go server
```

Server listens on `:8001`.

## Generate code

### Generate server code

```shell
jzero gen
```

### Generate swagger code

```shell
jzero gen swagger
```

you can see generated swagger json in `desc/swagger`

## Build docker image

```shell
# add a builder first
docker buildx create --use --name=mybuilder --driver docker-container --driver-opt image=dockerpracticesig/buildkit:master

# build and load
docker buildx build --platform linux/amd64 --progress=plain -t server:latest . --load
```

## Documents

https://docs.jzero.io