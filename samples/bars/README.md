
# Bars Sample

## What is in the sample?

- **bars_aceonly** contains a BAR for sample applications that don't need MQ. These will be copied into the image (at build time) to `/home/aceuser/bars` and compiled. The Integration Server will pick up and deploy this files on start up. These set of BAR files will be copied when building an image with ACE only or ACE & MQ.

## Building the sample

First [build the ACE image](../README.md#Building-a-container-image) or obtain one of the shipped images

In the `sample/bars` folder:

```bash
docker build -t aceapp --build-arg FROMIMAGE=ace:12.0.4.0-r1 --file Dockerfile .
```

## Running the sample

The sample application is a copy of one of the ACE samples called CustomerDB. This provides a RestAPI which can be queries which will return information about customers.

To run the application launch the container using a command such as:

```bash
`docker run -d -p 7600:7600 -p 7800:7800 -e LICENSE=accept aceapp`
```

To exercise the flow run a command such as:

```bash
curl --request GET \
  --url 'http://96fa448ea76e:7800/customerdb/v1/customers?max=REPLACE_THIS_VALUE' \
  --header 'accept: application/json'
```
