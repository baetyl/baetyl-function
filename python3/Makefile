MODULE:=python3
BIN:=baetyl-$(MODULE)
SRC_FILES:=$(shell find . -type f -name '*.py')
PLATFORM_ALL:=darwin/amd64 linux/amd64 linux/arm64 linux/arm/v7

export DOCKER_CLI_EXPERIMENTAL=enabled

GIT_TAG:=$(shell git tag --contains HEAD|awk 'END {print}')
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

GO_OS:=$(shell go env GOOS)
GO_ARCH:=$(shell go env GOARCH)
GO_ARM:=$(shell go env GOARM)

ifndef PLATFORMS
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

OUTPUT     :=../output
OUTPUT_DIRS:=$(PLATFORMS:%=$(OUTPUT)/%/$(BIN))
OUTPUT_BINS:=$(OUTPUT_DIRS:%=%/$(BIN))
PKG_PLATFORMS := $(shell echo $(PLATFORMS) | sed 's:/:-:g')
OUTPUT_PKGS:=$(PKG_PLATFORMS:%=$(OUTPUT)/$(BIN)_%_$(VERSION).zip)

.PHONY: image
image:
	@echo "BUILDX: $(REGISTRY)$(MODULE):3.6-$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE):3.6-$(VERSION) .

.PHONY: image-opencv
image-opencv:
	@echo "BUILDX: $(REGISTRY)$(MODULE)-opencv:3.6-$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE)-opencv:3.6-$(VERSION) -f Dockerfile-opencv .

.PHONY: build
build: $(OUTPUT_BINS)

$(OUTPUT_BINS): $(SRC_FILES)
	@echo "BUILD $@"
	@mkdir -p $(dir $@)
	@rm -rf temp && mkdir temp && unzip lib/$(shell echo $(@:$(OUTPUT)/%/$(BIN)/$(BIN)=%) | sed 's:/:-:g').zip -d temp
	@cp -r temp/* $^ program.yml $(dir $@) && rm -rf temp

.PHONY: package
package: build $(OUTPUT_PKGS)

$(OUTPUT_PKGS):
	@echo "PACKAGE $@"
	@cd $(OUTPUT)/$(shell echo $(@:$(OUTPUT)/$(BIN)_%_$(VERSION).zip=%) | sed 's:-:/:g')/$(BIN) && zip -q -r $(notdir $@) *