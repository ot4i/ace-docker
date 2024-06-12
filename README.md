# Overview

![IBM ACE logo](./app_connect_light_256x256.png)

Run [IBM® App Connect Enterprise](https://developer.ibm.com/integration/docs/app-connect-enterprise/faq/) in a container.

This repo is designed to provide information about how to build a simple ACE container and how to extend it with extra capability as per the requirements of your use case.

If you would like to use pre-built containers please refer to [Pre-Built Containers](#pre-built-containers)

If you are looking for information on the previous images that were documented in this repo, please refer to the previous [releases](https://github.com/ot4i/ace-docker/releases). The previous images are designed only for use with the App Connect operator. They are not designed for use in your own non-operator deployment.

## Building a container image

**Important:** Only ACE version **12.0.1.0 or greater** is supported.

Before building the image you must obtain a copy of the relevant build of ACE and make sure it is available on an HTTP endpoint.

When using an insecure http endpoint, build the image using a command such as:

```bash
docker build -t ace --build-arg DOWNLOAD_URL=<download URL>  --file ./Dockerfile .
```

If you want to connect to a secure endpoint build the image using a command such as:
i.e.

```bash
docker build -t ace --build-arg USERNAME=<Username> --build-arg PASSWORD=<Password> --build-arg DOWNLOAD_URL=<download URL>  --file ./Dockerfile .
```

### DOWNLOAD_URL

The link is likely to be of the form
```
https://iwm.dhe.ibm.com/sdfdl/v2/regs2/mbford/Xa.2/Xb.WJL1cUPI9gANEhP8GuPD_qX1rj6x5R4yTUM7s_C2ue8/Xc.12.0.10.0-ACE-LINUX64-DEVELOPER.tar.gz/Xd./Xf.LpR.D1vk/Xg.12164875/Xi.swg-wmbfd/XY.regsrvs/XZ.pPVETUejcqPsVfDVKbdNu6IRpo4TkyKu/12.0.10.0-ACE-LINUX64-DEVELOPER.tar.gz
```
Use this link as the DOWNLOAD_URL build parameter, adjusting the version numbers in the other files and parameters as needed.

### Running the image

To run the image use a command such as

`docker run -d -p 7600:7600 -p 7800:7800 -e LICENSE=accept -e ACE_SERVER_NAME=myserver ace:latest`

where `ACE_SERVER_NAME` is the name of the Integration Server that will be running.

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
