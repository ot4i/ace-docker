# Overview

![IBM ACE logo](./app_connect_light_256x256.png)

Run [IBM® App Connect Enterprise](https://developer.ibm.com/integration/docs/app-connect-enterprise/faq/) in a container.

This repo is designed to provide information about how to build a simple ACE container and how to extend it with extra capability as per the requirements of your use case.

If you would like to use pre-built containers please refer to [Pre-Built Containers](#pre-built-containers)

If your looking for information on the previous images that were documented in this repo, please refer to the previous [releases](https://github.com/ot4i/ace-docker/releases). The previous images are designed only for use with the App Connect operator. They are not designed for use in your own non operator led deployment.

## Building a container image

**Important:** Only ACE version **12.0.1.0 or greater** is supported.

Before building the image you must obtain a copy of the relavent build of ACE and make it available on a HTTP endpoint.

When using an insecure http endpoint, build the image using a command such as:

```bash
docker build -t ace --build-arg DOWNLOAD_URL=${DOWNLOAD_URL}  --file ./Dockerfile .
```

If you want to connect to a secure endpoint build the image using a command such as:
i.e.

```bash
docker build -t ace --build-arg USERNAME=<Username> --build-arg PASSWORD=<Password> --build-arg DOWNLOAD_URL=${DOWNLOAD_URL}  --file ./Dockerfile .
```

NOTE: If no DOWNLOAD_URL is provided the build will use a copy of the App Connect Enterprise developer edition as referenced in the Dockerfile

### Running the image

To run the image use a command such as

`docker run -d -p 7600:7600 -p 7800:7800 -e LICENSE=accept ace:latest`

### Extending the image

To add extra artifacts into the container such as server.conf.yaml overrides, bars files etc please refer to the sample on adding  files in [Samples](samples/README.md)

## Pre-Built Containers

Pre-built production images can be found on IBM Container Registry at `cp.icr.io/cp/appc/ace` - [Building a sample IBM App Connect Enterprise image using Docker](https://www.ibm.com/docs/en/app-connect/12.0?topic=cacerid-building-sample-app-connect-enterprise-image-using-docker)

### Fixing security vulnerabilities in the base software

If you find there are vulnerabilities in the base redhat software there are two options to fix this

- Apply the fix yourself using a sample docker file to install all available updates - [Samples/updateBase](samples/updateBase/Dockerfile)
- Pick up the latest version of the image which will include all fixes at the point it was built.

### Fixing issues with ACE

If you find a problem with ACE software, raise a PMR to obtain a fix. Once the fix is provided this can be applied to any existing image using a [Samples](samples/README.md#ifix-sample) dockerfile

## Support

All information provided in this repo as supported as-is.
