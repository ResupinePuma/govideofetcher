REGISTRY ?= 192.168.1.2:5000

build-bot:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $(REGISTRY)/vf:latest \
		-f docker/Dockerfile .

push-bot:
	docker push $(REGISTRY)/vf:latest

build-api:
	docker buildx build \
	    --platform linux/arm64,linux/amd64 \
	    --push -t 192.168.1.2:5000/dlapi \
	    -f docker/Dockerfile.ytdlapi .

push-api:
	docker compose push api

build: build-bot build-api

push: push-bot push-api

all: build
