# start docker
docker compose up --build

# migrations 

## get go migrate

```shell
go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate
```

## create migration

```shell
migrate create -ext sql -dir migrations <migration_name>
```

## apply migrations 

```shell
migrate -database 'postgres://postgres:admin@localhost:5432/?sslmode=disable' -path migrations up

```