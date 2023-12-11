
# MQClient Sample

## What is in the sample?

- **mqclient** contains a sample dockerfile for adding the MQ client into the image.

## Building the sample

First [build the ACE image](../../README.md#Building-a-container-image) or obtain one of the shipped images.

Determine the link for the version of the MQ client to be used, using FixCentral or the standard redistributable
client download site. An example URL would be https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.3.2.0-IBM-MQC-Redist-LinuxX64.tar.gz
for the client.

In this folder, run the docker command as follows (replacing the product version and MQ link as appropriate):

```bash
docker build -t aceappmqclient --build-arg FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.10.0-r1 --build-arg MQ_URL=https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.3.2.0-IBM-MQC-Redist-LinuxX64.tar.gz --file Dockerfile .
```

## Running the sample

To run the container, launch the pod using a command such as:

```bash
docker run -d -p 7600:7600 -p 7800:7800 -e LICENSE=accept aceappmqclient
```

This wll then start an IntegrationServer with the MQ Client installed.
