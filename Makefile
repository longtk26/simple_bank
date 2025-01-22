migrateup:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup1:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
migrateup-mysql:
	migrate -path db/migrations -database "mysql://root:password@tcp(localhost:3306)/simple_bank" -verbose up
migratedown:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migrations -database "postgresql://postgres:root@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate
test:
	go test -v -cover ./...
server:
	go run main.go
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/longtk26/simple_bank/db/sqlc Store
newmg:
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: migrateup migratedown sqlc test migrateup-mysql server mock newmg migrateup1 migratedown1