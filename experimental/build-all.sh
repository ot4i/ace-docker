#!/bin/bash
export PRODUCT_VERSION=11.0.0.9
export PRODUCT_LABEL=ace-${PRODUCT_VERSION}
export DOWNLOAD_URL=http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/11.0.0.9-ACE-LINUX64-DEVELOPER.tar.gz

cd ace-minimal
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-alpine -f Dockerfile.alpine .
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-ubuntu -f Dockerfile.ubuntu .
# Highly experimental!
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-alpine-openjdk14 -f Dockerfile.alpine-openjdk14 .

cd ../ace-full
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-full:${PRODUCT_VERSION}-ubuntu -f Dockerfile.ubuntu .

cd ../sample
# Normally only one of these would be built, and would be tagged with an application version
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-minimal:${PRODUCT_VERSION}-alpine -t ace-sample:${PRODUCT_VERSION}-alpine -f Dockerfile .
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-minimal:${PRODUCT_VERSION}-ubuntu -t ace-sample:${PRODUCT_VERSION}-ubuntu -f Dockerfile .
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-full:${PRODUCT_VERSION}-ubuntu -t ace-sample:${PRODUCT_VERSION}-full-ubuntu -f Dockerfile .
