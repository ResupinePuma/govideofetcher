build-bot:
	docker compose -f docker-compose.yml build govf
push-bot:
	docker compose -f docker-compose.yml build govf 	

all: build-botpush-bot
