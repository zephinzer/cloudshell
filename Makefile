changes ?= $(shell git status --porcelain | wc -l)
version ?= $(shell git rev-parse HEAD | head -c 8)

-include Makefile.properties

image_namespace ?= zephinzer
image_name ?= cloudshell
image_tag ?= $(version)

init:
	npm install
	go mod vendor
start:
	go run ./cmd/cloudshell
run: build
	docker run -it -p 8376:8376 $(image_namespace)/$(image_name):latest
build:
	docker build \
		--tag $(image_namespace)/$(image_name):latest \
		.
build-example: build
	if [ "${id}" = "" ]; then \
		printf -- '\033[1m\033[31m$${id} was not specified\033[0m\n'; \
		exit 1; \
	fi
	docker build \
		--build-arg IMAGE_NAMESPACE=$(image_namespace) \
		--build-arg IMAGE_NAME=$(image_name) \
		--build-arg IMAGE_TAG=latest \
		--tag $(image_namespace)/$(image_name):${id}-latest \
		--file ./examples/${id}/Dockerfile \
		.

ifeq ($(shell test $(changes) -gt 0; echo $$?),0)
image_tag := $(image_tag)-dev
endif
publish: build
	-docker push $(image_namespace)/$(image_name):latest
	docker tag $(image_namespace)/$(image_name):latest \
		$(image_namespace)/$(image_name):$(image_tag)
	docker push $(image_namespace)/$(image_name):$(image_tag)
publish-example: build-example
	-docker push $(image_namespace)/$(image_name):${id}-latest
	docker tag $(image_namespace)/$(image_name):${id}-latest \
		$(image_namespace)/$(image_name):${id}-$(image_tag)
	docker push $(image_namespace)/$(image_name):${id}-$(image_tag)
