run:
	# the "-" sign is to ignore errors
	-make down
	docker compose up --build

down:
	docker compose down -v --remove-orphans