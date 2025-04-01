run:
	# the "-" sign is to ignore errors
	-make down
	docker compose up --build

down:
	docker compose down -v --remove-orphans

integration-test:
	go test -v ./internal/infra/database/auction/create_auction_test.go
