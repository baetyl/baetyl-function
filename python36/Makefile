MODULE:=python
PLATFORM_ALL:=linux/amd64 linux/arm64 linux/arm/v7

GIT_TAG:=$(shell git tag --contains HEAD)
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

ifeq ($(PLATFORMS),all)
	override PLATFORMS:=$(PLATFORM_ALL)
endif

REGISTRY?=
XFLAGS?=--load
XPLATFORMS:=$(shell echo $(PLATFORMS) | sed 's: :,:g')

.PHONY: image
image:
	@echo "BUILDX: $(REGISTRY)$(MODULE):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)/$(MODULE):3.6-$(VERSION) -f Dockerfile .
