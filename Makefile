.PHONY: build

BUILD_MAJOR := $(shell date -u +%y%m%d%H%M)
BUILD_MINOR := $(shell cat build_number)

build:
	@echo
	@echo "⋮⋮ Building..."
	go build -ldflags "-X main.version=$(BUILD_MAJOR).$(BUILD_MINOR)" ./cmd/jasmine/jasmine.go

init:
	@echo
	@echo "⋮⋮ Init..."
	git config core.hooksPath .githooks

test: 
	@echo
	@echo "⋮⋮ Testing..."
	go test -count=1 ./... -coverprofile cover.out

review:
	@echo
	@echo "⋮⋮ Reviewing tests..."
	go tool cover -html cover.out

package:
	@echo
	@echo "⋮⋮ Packaging container..."
	docker build -f ./build/package/Dockerfile --build-arg BUILD_NUM="$(BUILD_MAJOR).$(BUILD_MINOR)" -t giuseppe007/jasmine:local .

run: package
	@echo
	@echo "⋮⋮ Running the container locally..."
	docker run -e JASMINE_JIRAAPIKEY=$(JASMINE_JIRAAPIKEY) -p 2112:2112 -v ./configs/jasmine/config.yaml:/config.yaml:ro giuseppe007/jasmine:local --config=/config.yaml

publish:
	@echo
	@echo "⋮⋮ Publish the container..."
	docker tag giuseppe007/jasmine:local giuseppe007/jasmine:$(BUILD_MAJOR).$(BUILD_MINOR)
	docker push giuseppe007/jasmine:$(BUILD_MAJOR).$(BUILD_MINOR)

local:
	@echo
	@echo "⋮⋮ Launching containers for local use..."
	# docker compose -f ./deployments/docker-compose.yaml --project-name jasmine up -d --force-recreate
	docker compose -f ./deployments/docker-compose.yaml --project-name jasmine up -d 
	docker ps | grep jasmine

clean-local:
	docker compose -f ./deployments/docker-compose.yaml --project-name jasmine down


all: build test package
	@echo

