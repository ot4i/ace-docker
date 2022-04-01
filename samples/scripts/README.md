
# Scripts Sample

## What is in the sample?

This sample includes copying in a script which should run before the IntegrationServer starts up. These scripts are used to process a mounted file which contains some credentials.

## Building the sample

First [build the ACE image](../README.md#Building-a-container-image) or obtain one of the shipped images

In the `sample/scripts` folder:

```bash
docker build -t aceapp --build-arg FROMIMAGE=ace:12.0.4.0-r1 --file Dockerfile .
```

## Running the sample

To run the application launch the container using a command such as:

```bash
`docker run --name aceapp -d -p 7600:7600 -p 7800:7800 -v <fullyQualifiedPath>/setdbparms:/home/aceuser/initial-config/setdbparms -e LICENSE=accept aceapp`
```

On startup the logs will show that the configured scripts are run before starting the integration server. These credentials can then be used by flows referencing the appropriate resource
