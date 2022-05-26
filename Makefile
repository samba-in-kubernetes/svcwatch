

COMMIT_ID = $(shell git describe --abbrev=40 --always --exclude='*' --dirty=+ 2>/dev/null)
GIT_VERSION = $(shell git describe --match='v[0-9]*.[0-9]' --match='v[0-9]*.[0-9].[0-9]' 2>/dev/null || echo "(unset)")

# Image URL to use all building/pushing image targets
TAG ?= latest
IMG ?= quay.io/samba.org/svcwatch:$(TAG)

CONTAINER_BUILD_OPTS ?=
CONTAINER_CMD ?=
ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD:=$(shell docker version >/dev/null 2>&1 && echo docker)
endif
ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD:=$(shell podman version >/dev/null 2>&1 && echo podman)
endif

build:
	go build -o bin/svcwatch -ldflags "-X main.Version=$(GIT_VERSION) -X main.CommitID=$(COMMIT_ID)"  main.go
.PHONY: build

# Run go vet against code
vet:
	go vet ./...


# Build the container image
image-build:
	$(CONTAINER_CMD) build \
		--build-arg=GIT_VERSION="$(GIT_VERSION)" \
		--build-arg=COMMIT_ID="$(COMMIT_ID)" \
		$(CONTAINER_BUILD_OPTS) $(CONTAINER_BUILD_OPTS) . -t ${IMG} -f Containerfile
.PHONY: image-build


.PHONY: image-build-buildah
image-build-buildah: build
	cn=$$(buildah from registry.access.redhat.com/ubi8/ubi-minimal:latest) && \
	buildah copy $$cn bin/svcwatch /svcwatch && \
	buildah config --cmd='[]' $$cn && \
	buildah config --entrypoint='["/svcwatch"]' $$cn && \
	buildah commit $$cn $(IMG)


# Push the container image
container-push:
	$(CONTAINER_CMD) push ${IMG}


.PHONY: check check-revive check-format

check: check-revive check-format vet

check-format:
	! gofmt $(CHECK_GOFMT_FLAGS) . | sed 's,^,formatting error: ,' | grep 'go$$'

check-revive: revive
	# revive's checks are configured using .revive.toml
	# See: https://github.com/mgechev/revive
	$(REVIVE) -config .revive.toml $$(go list ./... | grep -v /vendor/)

.PHONY: revive
revive:
ifeq (, $(shell command -v revive ;))
	@echo "revive not found in PATH, checking in GOBIN ($(GOBIN))..."
ifeq (, $(shell command -v $(GOBIN)/revive ;))
	@{ \
	set -e ;\
	echo "revive not found in GOBIN, getting revive..." ;\
	REVIVE_TMP_DIR=$$(mktemp -d) ;\
	cd $$REVIVE_TMP_DIR ;\
	go mod init tmp ;\
	go get  github.com/mgechev/revive  ;\
	rm -rf $$REVIVE_TMP_DIR ;\
	}
	@echo "revive installed in GOBIN"
else
	@echo "revive found in GOBIN"
endif
REVIVE=$(shell command -v $(GOBIN)/revive ;)
else
	@echo "revive found in PATH"
REVIVE=$(shell command -v revive ;)
endif
