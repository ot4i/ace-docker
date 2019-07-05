#!/bin/bash

export DOWNLOAD_URL=http://kenya.hursley.uk.ibm.com:52367/ace-11.0.16836.5.tar.gz
export PRODUCT_VERSION=ace-11.0.16836.5

cd ace-minimal-install
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_VERSION -t ace-minimal-install:11.0.0.5-alpine -f Dockerfile.alpine .
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_VERSION -t ace-minimal-install:11.0.0.5-ubuntu -f Dockerfile.ubuntu .
docker build --build-arg DOWNLOAD_URL --build-arg PRODUCT_VERSION -t ace-minimal-install:11.0.0.5-alpine-openjdk -f Dockerfile.alpine-openjdk .

cd ../ace-minimal
docker build --build-arg LICENSE=accept -t ace-minimal:11.0.0.5-alpine -f Dockerfile.alpine .
docker build --build-arg LICENSE=accept -t ace-minimal:11.0.0.5-ubuntu -f Dockerfile.ubuntu .
docker build --build-arg LICENSE=accept -t ace-minimal:11.0.0.5-alpine-openjdk -f Dockerfile.alpine-openjdk .

cd ../sample
# Normally only one of these would be built, and would be tagged with an application version
docker build --build-arg LICENSE=accept -t ace-sample-alpine -f Dockerfile.alpine .
docker build --build-arg LICENSE=accept -t ace-sample-ubuntu -f Dockerfile.ubuntu .
docker build --build-arg LICENSE=accept -t ace-sample-alpine-openjdk -f Dockerfile.alpine-openjdk .
