# Overview

This repository contains some Dockerfiles and some scripts which demonstrate a way in which you might run [IBM App Connect Enterprise](https://www.ibm.com/cloud/app-connect/enterprise) in a [Docker](https://www.docker.com/whatisdocker/) container.

The base image contains a full installation of [IBM App Connect Enterprise for Developers Version 11.0](https://ibm.biz/iibdevedn), as well as some system configuration and user creation.

The demo image extends the base image by copying in a BAR file and performing the necessary setup.

# Docker Hub

A pre-built version of the ACE image is available on Docker Hub as [`ibmcom/ace`](https://hub.docker.com/r/ibmcom/ace/) with the following tags:

  * `11.0.0.0`, `latest` ([Dockerfile](https://github.com/ot4i/ace-docker/blob/master/11.0.0.0/ace/ubuntu-1604/base/Dockerfile))

# Building the image

The image can be built using standard [Docker commands](https://docs.docker.com/userguide/dockerimages/) against the supplied Dockerfile:

~~~
cd 11.0.0.0/ace/ubuntu-1604/base
docker build -t ace:11.0.0.0 .
~~~

To build the demo image you would do the following:

~~~
cd 11.0.0.0/ace/ubuntu-1604/demo
docker build -t ace-demo:11.0.0.0 .
~~~

# Running a container

## Running the base image

After pulling the prebuilt image from Docker Hub, or building a Docker image from the supplied files, you can [run a container](https://docs.docker.com/userguide/usingdocker/) which will start an Integration Server listening on port 7600.

For example:

~~~
docker run --name myAce -e LICENSE=accept -P ibmcom/ace:11.0.0.0
~~~

## Verifying your container is running correctly

Whether you are using the image as provided or if you have customised it, here are a few basic steps that will give you confidence your image has been created properly:

1. Run a container, making sure to expose port 7600 to the host - the container should start without error
2. Connect a browser to your host on the port you exposed in step 1 - the IBM App Connect Enterprise web user interface should be displayed.

## Running the extended image

Start with the provided demo image, but then customised with your own BAR file. Build the image and then run as described above (remembering to reference the image by whatever name you built it with).

For example:

~~~
docker build -t ace-bar:1.0 .
docker run --name myAceBar -e LICENSE=accept -P ace-bar:1.0
~~~

# Issues and contributions

For issues relating specifically to this Docker image, please use the [GitHub issue tracker](https://github.com/ot4i/ace-docker/issues). For more general issues relating to IBM App Connect Enterprise, or to discuss the Docker technical preview, please use the [Integration Community](https://developer.ibm.com/integration/). If you do submit a Pull Request related to this Docker image, please indicate in the Pull Request that you accept and agree to be bound by the terms of the [IBM Contributor License Agreement](CLA.md).

# License

The Dockerfile and associated scripts are licensed under the [Eclipse Public License 2.0](LICENSE). Licenses for the products installed within the images are as follows:

 - IBM App Connect Enterprise for Developers is licensed under the IBM International License Agreement for Non-Warranted Programs. This license may be viewed from the image using the `LICENSE=view` environment variable as described above.
 - License information for Ubuntu packages may be found in `/usr/share/doc/${package}/copyright`

Note that the IBM App Connect Enterprise for Developers license does not permit further distribution.

