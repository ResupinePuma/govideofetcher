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
		--platform linux/amd64,linux/arm64 \
		--push \
		-t $(REGISTRY)/ytdlapi:latest \
		-f docker/Dockerfile.ytdlapi .

push-api:
	docker push $(REGISTRY)/ytdlapi:latest

build: build-bot build-api

push: push-bot push-api


all: build
