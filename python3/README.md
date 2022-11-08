## python3
python3+ module

### Build
1. make image-devel PLATFORMS="linux/amd64 linux/arm64 linux/arm/v7" PYTHON_VERSION=3.11 XFLAGS=--push REGISTRY=baetyltechtest/
2. make image PLATFORMS="linux/amd64 linux/arm64 linux/arm/v7" PYTHON_VERSION=3.7 XFLAGS=--push REGISTRY=baetyltechtest/