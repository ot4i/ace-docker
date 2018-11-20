# Sample

The sample folder contains Docker files that build a sample image containing sample applications and Integration Server configuration. This requires you to mount the sample `initial-config` folder as well.

## Build the sample image

### Sample based on the ACE & MQ image

First [build the ACE & MQ image](../README.md#build-an-image-with-app-connect-enterprise-and-message-queue).

In the `sample` folder:

```
docker build -t acemqapp --file Dockerfile.acemq .
```

### Sample based on the ACE only image

First [build the ACE only image](../README.md#build-an-image-with-app-connect-enterprise-only).

In the `sample` folder:

```
docker build -t aceapp --file Dockerfile.aceonly .
```

## Run the sample image

### ACE & MQ image

To run a container based on the ACE & MQ image and these settings:
- queue manager name `QMGR`
- listener for MQ on port `1414`
- ACE server name `ACESERVER`
- listener for ACE web ui on port `7600`
- listener for ACE HTTP on port `7600`
- ACE truststore password `truststorepwd`
- ACE keystore password `keystorepwd`

And mounting `sample/initial-config` with the sample configuration into `/home/aceuser/initial-config`.

> **Note**: Always mount any initial config to be processed on start up to `/home/aceuser/initial-config`.

`docker run --name acemqapp -p 7600:7600 -p 7800:7800 -p 7843:7843 -p 1414:1414 --env LICENSE=accept --env MQ_QMGR_NAME=QMGR --env ACE_SERVER_NAME=ACESERVER --mount type=bind,src=/{path to repo}/sample/initial-config,dst=/home/aceuser/initial-config --env ACE_TRUSTSTORE_PASSWORD=truststorepwd --env ACE_KEYSTORE_PASSWORD=keystorepwd acemqapp:latest`

### ACE only image

To run a container based on the ACE only image and these settings:
- ACE server name `ACESERVER`
- listener for ACE web ui on port `7600`
- listener for ACE HTTP on port `7600`
- ACE truststore password `truststorepwd`
- ACE keystore password `keystorepwd`

And mounting `sample/initial-config` with the sample configuration into `/home/aceuser/initial-config`.

> **Note**: Always mount any initial config to be processed on start up to `/home/aceuser/initial-config`.

`docker run --name aceapp -p 7600:7600 -p 7800:7800 -p 7843:7843 --env LICENSE=accept --env ACE_SERVER_NAME=ACESERVER --mount type=bind,src=/{path to repo}/sample/initial-config,dst=/home/aceuser/initial-config --env ACE_TRUSTSTORE_PASSWORD=truststorepwd --env ACE_KEYSTORE_PASSWORD=keystorepwd aceapp:latest`

## What is in the sample?

- **bars_aceonly** contains BAR files for sample applications that don't need MQ. These will be copied into the image (at build time) to `/home/aceuser/bars` and compiled. The Integration Server will pick up and deploy this files on start up. These set of BAR files will be copied when building an image with ACE only or ACE & MQ.
- **bars_mq** contains BAR files for sample applications that need MQ. These will be copied into the image (at build time) to `/home/aceuser/bars` and compiled. The Integration Server will pick up and deploy this files on start up. These set of BAR files will be copied when building an image ACE & MQ.
- **dashboards** contains json defined sample grafana and kibana dashboards. This image has a prometheus exporter, which makes information available to prometheus for statistics data. The grafana dashboard gives a sample visualization of this data. The kibana dashboard is based on the json output from an Integration Server (v11.0.0.2). `LOG_FORMAT` env variable has to be set to `json`.
- **initial-config** is a directory that can be mounted by the container. This contains sample configuration files that the container will process on start up.
- **mqsc** contains a `.mqsc` file that sets up queues used by the applications. This is copied into the image (at build time) to `/etc/mqm` so that the queue manager processes it on start up.
- **PIs** contain Project Interchange files with the source for the applications that will be loaded from `bars_aceonly` and `bars_mq`.
- **Dockerfile.aceonly** the Dockerfile that builds "FROM" the `ace-only` image and builds an application image.
- **Dockerfile.acemq** the Dockerfile that builds "FROM" the `ace-mq` image and builds an application image.

## What are the sample applications?
All of the applications either just echo back a timestamp, or call another flow that echoes a timestamp:

- **HTTPSEcho** - Presents an echo service over a HTTPS endpoint on port 7843. This uses self-signed certificates (from `initial-config/keystore`). You can call this by going to https://localhost:7843/Echo (though you will get an exception about the certificate), or running `curl -k https://localhost:7843/Echo`. This demonstrates how to deploy a flow that presents a HTTPS endpoint, and how to apply a policy and custom `server.conf.yaml` overrides.
- **CallHTTPSEcho** - Wraps a call to `HTTPSEcho` and presents a service over HTTP that is convenient to call. This uses out self-signed CA certificate (from `initial-config/truststore`) to ensure the call to the HTTPS server is trusted. You can call this by going to http://localhost:7800/CallHTTPSEcho, or running `curl http://localhost:7800/CallHTTPSEcho`. This demonstrates how to deploy flow that calls a HTTPS endpoint with specific trust certificates.
- **MQEcho** - Presents an echo service over MQ, using the input queue `ECHOIN` and output queue `ECHOOUT`. You can't call this flow directly as another flow is reading from it's output queue.
- **CallMQEcho** - Wraps a call to `MQEcho` and presents a service over HTTP that is convenient to call. You can call this by going to http://localhost:7800/CallMQEcho, or running `curl http://localhost:7800/CallMQEcho`. This demonstrates how to use MQ for your own flows internally.
- **MQMultiEcho** - Is a duplicate of `MQEcho` (so another flow can call it exclusively), but using the input queue `MULTIECHOIN` and output queue `MULTIECHOOUT`. This is effectively mocking an MQ application for `CallMQMultiEcho`. You can't call this flow directly as another flow is reading from it's output queue.
- **CallMQMultiEcho** - Presents an MQ endpoint that calls `MQMultiEcho` multiple times and aggregates the responses into a single message, using the input queue `MULTIIN` and output queue `MULTIOUT`. Note that this mocks delays in the MQ application so you may have to wait for the response.  This demonstrates how to use message nodes that store state, and how to access the queues that drive those flows.

   Call this from inside the container:
   - Put a message using `echo "x" | /opt/mqm/samp/bin/amqsput MULTIIN QMGR`.
   - Get the response using `/opt/mqm/samp/bin/amqsget MULTIOUT QMGR`.

   Call this from outside the container:
   - Put a message using `MQSAMP_USER_ID=aceuser MQSERVER=ACE.SVRCONN/TCP/localhost(1414) /opt/mqm/samp/bin/amqsputc MULTIIN`, enter the password `passw0rd`, then enter anything for the message text and hit CTRL+D.
   - Get the response using `MQSAMP_USER_ID=aceuser MQSERVER=ACE.SVRCONN/TCP/localhost(1414) /opt/mqm/samp/bin/amqsgetc MULTIOUT QMGR`, and enter the password `passw0rd`.
