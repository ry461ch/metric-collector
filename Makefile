start-db:
	docker compose --env-file .env.example up -d
stop-db:
	docker compose --env-file .env.example down
