MODULE:=function-node
BIN:=baetyl-$(MODULE)
SRC_FILES:=function_grpc_pb.js function_pb.js runtime.js
PLATFORM_ALL:=linux/amd64 linux/arm64 linux/arm/v7

export DOCKER_CLI_EXPERIMENTAL=enabled

GIT_TAG:=$(shell git tag --contains HEAD|awk 'END {print}')
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

ifndef PLATFORMS
	GO_OS:=$(shell go env GOOS)
	GO_ARCH:=$(shell go env GOARCH)
	GO_ARM:=$(shell go env GOARM)
	PLATFORMS:=$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))
	ifeq ($(GO_OS),darwin)
		PLATFORMS+=linux/amd64
	endif
else ifeq ($(PLATFORMS),all)
	override PLATFORMS:=$(PLATFORM_ALL)
endif

REGISTRY?=
XFLAGS?=--load
XPLATFORMS:=$(shell echo $(filter-out darwin/amd64,$(PLATFORMS)) | sed 's: :,:g')

OUTPUT:=../output
OUTPUT_PKGS:=$(PLATFORMS:%=$(OUTPUT)/%/baetyl/$(BIN)-$(VERSION).zip)

.PHONY: image
image:
	@echo "BUILDX: $(REGISTRY)$(MODULE):10.23-$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE):10.23-$(VERSION) .

$(OUTPUT_PKGS): $(SRC_FILES)
	@echo "PACKAGE $@"
	@mkdir -p $(dir $@)
	@install -d -m 0755 __package
	@npm install
	@cp -r $^ program.yml node_modules __package && cd __package && zip -q -r $(notdir $@) .
	@cp __package/$(notdir $@) $(dir $@)
	@rm -rf __package

.PHONY: package
package: $(OUTPUT_PKGS)
