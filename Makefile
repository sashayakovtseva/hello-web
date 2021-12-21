# suppress output, run `make XXX V=` to be verbose
V := @

# Common
NAME = hello-web
VCS = github.com
ORG = sashayakovtseva

# Build
OUT_DIR = ./bin
MAIN_PKG = ./cmd/${NAME}
ACTION ?= build
GC_FLAGS = -gcflags 'all=-N -l'
LD_FLAGS = -ldflags "-s -v -w"
BUILD_CMD = go build -o ${OUT_DIR}/${NAME} ${LD_FLAGS} ${MAIN_PKG}

# Docker
DOCKERFILE = deployments/docker/Dockerfile
DOCKER_IMAGE_NAME = ${ORG}/${NAME}

# Other
.DEFAULT_GOAL = build

.PHONY: build
build: clean
	$(V)buf build
	@echo BUILDING PRODUCTION $(NAME)
	$(V)${BUILD_CMD}
	@echo DONE

.PHONY: lint
lint:
	$(V)golangci-lint run
	$(V)buf lint

.PHONY: test
test: GO_TEST_FLAGS += -race
test:
	$(V)go test -mod=vendor $(GO_TEST_FLAGS) --tags=$(GO_TEST_TAGS) ./...

.PHONY: generate
generate:
	$(V)buf generate --path=./api/grpc
	$(V)go generate -x ./...

.PHONY: clean
clean:
	@echo "Removing $(OUT_DIR)"
	$(V)rm -rf $(OUT_DIR)

.PHONY: vendor
vendor:
	$(V)GOPRIVATE=${VCS}/* go mod tidy -compat=1.17
	$(V)GOPRIVATE=${VCS}/* go mod vendor
	$(V)buf mod update
	$(V)git add vendor go.mod go.sum buf.lock

.PHONY: docker-build-push-image
docker-build-push-image:
	$(V)docker build -t ${DOCKER_IMAGE_NAME}:${VERSION} -f ${DOCKERFILE} --build-arg ACTION=${ACTION} .
	$(V)docker push ${DOCKER_IMAGE_NAME}:${VERSION}
