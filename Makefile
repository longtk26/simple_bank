migrateup:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup-mysql:
	migrate -path db/migrations -database "mysql://root:password@tcp(localhost:3306)/simple_bank" -verbose up
migratedown:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose down
sqlc:
	sqlc generate
test:
	go test -v -cover ./...
server:
	go run main.go

.PHONY: migrateup migratedown sqlc test migrateup-mysql server