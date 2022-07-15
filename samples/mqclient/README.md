
# Bars Sample

## What is in the sample?

- **mqclient** contains a sample dockerfile for adding the MQ client into the image.

## Building the sample

First [build the ACE image](../../README.md#Building-a-container-image) or obtain one of the shipped images.

Download a copy of the MQ redistributable client Example URL: https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.2.0.4-IBM-MQC-Redist-LinuxX64.tar.gz

In the `sample/mqclient` folder:

```bash
docker build -t aceapp --build-arg FROMIMAGE=ace:12.0.4.0-r1 --build-arg MQ_URL=https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.2.0.4-IBM-MQC-Redist-LinuxX64.tar.gz --file Dockerfile .
```

## Running the sample

To run the container, launch the pod using a command such as:

```bash
`docker run -d -p 7600:7600 -p 7800:7800 -e LICENSE=accept aceapp`
```

This wll then start an IntegrationServer with a MQ Client installed.
