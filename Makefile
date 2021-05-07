changes ?= $(shell git status --porcelain | wc -l)
version ?= $(shell git rev-parse HEAD | head -c 8)
ifeq ($(shell test $(changes) -gt 0; echo $$?),0)
version := $(version)-dev
endif

# use this to override/set settings
-include Makefile.properties

image_namespace ?= zephinzer
image_name ?= cloudshell
image_tag ?= $(version)

init:
	npm install
	go mod vendor
start:
	go run ./cmd/cloudshell
run: package
	docker run -it -p 8376:8376 $(image_namespace)/$(image_name):latest
build:
	CGO_ENABLED=0 \
	go build -a -v \
		-ldflags " \
			-s -w \
			-extldflags 'static' \
			-X main.VersionInfo='$(version)' \
		" \
		-o ./bin/cloudshell ./cmd/cloudshell
package:
	docker build \
		--build-arg VERSION_INFO=$(version) \
		--tag $(image_namespace)/$(image_name):latest \
		.
	docker tag $(image_namespace)/$(image_name):latest \
		$(image_namespace)/$(image_name):$(version)
publish: package
	-docker push $(image_namespace)/$(image_name):latest
	docker push $(image_namespace)/$(image_name):$(version)
package-example: package
	if [ "${id}" = "" ]; then \
		printf -- '\033[1m\033[31m$${id} was not specified\033[0m\n'; \
		exit 1; \
	fi
	docker build \
		--build-arg IMAGE_NAMESPACE=$(image_namespace) \
		--build-arg IMAGE_NAME=$(image_name) \
		--build-arg IMAGE_TAG=$(version) \
		--tag $(image_namespace)/$(image_name):${id}-latest \
		--file ./examples/${id}/Dockerfile \
		.
	docker tag $(image_namespace)/$(image_name):${id}-latest \
		$(image_namespace)/$(image_name):${id}-$(version)
publish-example: package-example
	-docker push $(image_namespace)/$(image_name):${id}-latest
	docker push $(image_namespace)/$(image_name):${id}-$(version)	
