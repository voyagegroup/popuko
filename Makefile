DIST_NAME := popuko
GIT_REVISION := $(shell git rev-parse --verify HEAD)
BUILD_DATE := $(shell date '+%Y/%m/%d %H:%M:%S %z')

all: help

help:
	@echo "Specify the task"
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@exit 1

clean: ## Remove the exec binary.
	rm -rf $(CURDIR)/$(DIST_NAME)

bootstrap:
	rm -rf vendor/
	dep ensure

build: $(DIST_NAME) ## Build the exec binary for youe machine.

build_linux_x64: ## Just an alias to build for some cloud instance.
	env GOOS=linux GOARCH=amd64 make build -C $(CURDIR)

run: $(DIST_NAME) ## Execute the binary for youe machine.
	$(CURDIR)/$(DIST_NAME)

$(DIST_NAME): clean
	go build -o $(DIST_NAME) -ldflags "-X main.revision=$(GIT_REVISION) -X \"main.builddate=$(BUILD_DATE)\""

test: test_epic test_input test_operation test_queue test_setting
	go test

test_%:
	make test -C $(CURDIR)/$*

travis:
	make bootstrap
	make build -j 8
	make test -j 8
