


build:
	go build .

run: build
	./givedirectly -pg-password "test1234"

start-db:
	docker run --name libraryStorePG -d -p 5432:5432 --rm -e POSTGRES_PASSWORD="test1234" -e POSTGRES_USER="librarystore" -e POSTGRES_DB="librarystore"  postgres:13