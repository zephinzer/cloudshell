-include Makefile.properties

image_namespace ?= zephinzer
image_name ?= cloudshell
image_tag ?= latest

init:
	npm install
	go mod vendor
start:
	go run ./cmd/cloudshell
build:
	docker build \
		--tag $(image_namespace)/$(image_name):latest \
		.
run: build
	docker run -it -p 8376:8376 $(image_namespace)/$(image_name):latest
publish: build
	-docker push $(image_namespace)/$(image_name):latest
	docker tag $(image_namespace)/$(image_name):latest \
		$(image_namespace)/$(image_name):$(image_tag)
	docker push $(image_namespace)/$(image_name):$(image_tag)
