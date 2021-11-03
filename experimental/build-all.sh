#!/bin/bash
export PRODUCT_VERSION=12.0.2.0
export PRODUCT_LABEL=ace-${PRODUCT_VERSION}
export DOWNLOAD_URL=http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/12.0.2.0-ACE-LINUX64-DEVELOPER.tar.gz
#export DOWNLOAD_URL=http://kenya.hursley.uk.ibm.com:52367/ace-12.0.2.0.tar.gz

# Exit on error
set -e

cd ace-minimal
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-alpine -f Dockerfile.alpine .
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-ubuntu -f Dockerfile.ubuntu .
# Highly experimental!
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-alpine-openjdk14 -f Dockerfile.alpine-openjdk14 .
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-minimal:${PRODUCT_VERSION}-alpine-openjdk16 -f Dockerfile.alpine-openjdk16 .

cd ../ace-full
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-full:${PRODUCT_VERSION}-ubuntu -f Dockerfile.ubuntu .

cd ../ace-basic
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_LABEL -t ace-basic:${PRODUCT_VERSION}-ubuntu -f Dockerfile.ubuntu .

cd ../sample
# Normally only one of these would be built, and would be tagged with an application version
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-minimal:${PRODUCT_VERSION}-alpine -t ace-sample:${PRODUCT_VERSION}-alpine -f Dockerfile .
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-minimal:${PRODUCT_VERSION}-alpine-openjdk16 -t ace-sample:${PRODUCT_VERSION}-alpine-openjdk16 -f Dockerfile .
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-minimal:${PRODUCT_VERSION}-ubuntu -t ace-sample:${PRODUCT_VERSION}-ubuntu -f Dockerfile .
docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-full:${PRODUCT_VERSION}-ubuntu -t ace-sample:${PRODUCT_VERSION}-full-ubuntu -f Dockerfile .
