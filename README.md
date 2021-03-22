# GiveDirectly

## Requirements 

The local setup requires that you have `docker` and `docker-compose` installed.

## Quickstart

1. Build the required images: `docker-compose build`

2. Run the images: `docker-compose up`

The `docker-compose up` command will start a Postgres DB server and the apiserver for the library service.

## Testing

Create 2 new book requests:

```shell
curl -X POST -H "Content-Type: application/json" \
    -d '{"email": "test@gmail.com", "title": "testbook"}' \
    localhost:8080/request
    
curl -X POST -H "Content-Type: application/json" \
    -d '{"email": "test@gmail.com", "title": "testbook2"}' \
    localhost:8080/request
```

Get all current requests:

```shell
  curl localhost:8080/request
```

Get a specific request:
```shell
  curl localhost:8080/request/1
```

Delete a request

```shell
  curl -X DELETE localhost:8080/request/1
```

Validate it's been deleted

```shell
  curl localhost:8080/request
```