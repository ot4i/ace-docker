
# Bars Sample

## What is in the sample?

This directory contains a BAR file with a sample application that doesn't need MQ and a Dockerfile
that builds an image with the application deployed and ready to run. The BAR file (and other BAR 
files placed in this directory) will be copied into the image at build time and unpacked into the 
Integration Server work directory. When the server starts, the application will be started automatically. 

The [Dockerfile](Dockerfile) can also be used with a base image that contains an MQ client or
other pre-requisites; the CustomerDatabaseV1.bar application does not require any pre-reqs, but
some other applications (if any are added into this directory) may do. For more details about the 
CustomerDatabaseV1 application, see the `Using a REST API to manage a set of records` tutorial
in the ACE v12 toolkit.

The image build process compiles maps and schema files to avoid this needing to be done when the
container starts up, and also runs the [ibmint optimize server](https://www.ibm.com/docs/en/app-connect/12.0?topic=commands-ibmint-optimize-server-command)
command to ensure the ACE server will only load components that are needed for the applications.


## Building the sample

First [build the ACE image](../../README.md#Building-a-container-image) or obtain one of the shipped images

In the `sample/bars` folder:

```bash
docker build -t aceapp --build-arg FROMIMAGE=ace:12.0.7.0-r1 --file Dockerfile .
```

## Running the sample

The sample application is a copy of one of the ACE samples called CustomerDB. This provides a RestAPI 
which can be queried over HTTP to find out information about customers.

To run the application, launch the container using a command such as:

```bash
docker run -d --name aceapp -p 7600:7600 -p 7800:7800 -e LICENSE=accept aceapp
```

To exercise the flow, run a command such as:

```bash
curl --request GET \
  --url 'http://96fa448ea76e:7800/customerdb/v1/customers?max=REPLACE_THIS_VALUE' \
  --header 'accept: application/json'
```
