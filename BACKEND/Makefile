# with @, doesn't show the command in the CLI

build:
	@go build -o bin/pos_pharmacy cmd/main.go

test:
	@go test -v ./...
	
run: build
	@./bin/pos_pharmacy

# run:
# 	@go run cmd/main.go

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

migrate-force:
	@migrate -path PATH_TO_YOUR_MIGRATIONS -database YOUR_DATABASE_URL force VERSION