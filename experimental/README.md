# Experimental docker files

Simple docker images for ACE v12 on Linux (amd64 and s390x) and Windows

ace-full contains a Dockerfile for building an image that can run the full product, including mqsicreatebar with a virtual X server.

ace-basic contains a Dockerfile for building an image that can run the product server, including all files except the toolkit.

ace-minimal contains the Dockerfiles for building images that can run simple servers with a non-root user.

ace-sample contains a sample BAR file and Dockerfiles for building runnable images to serve HTTP clients.

See build-all.sh for details on building the images; setting LICENSE=accept is required for all but the initial image builds.

To run the sample after building:
```
docker run -e LICENSE=accept --rm -ti ace-sample:12.0.2.0-minimal-alpine
```
and then curl http://[container IP]:7800/test should return '{"data":"a string from ACE"}'

## Various sizes
Local on kenya.hursley.uk.ibm.com (debian 10) with defaults in Dockerfiles:

```
ace-minimal      12.0.2.0-alpine-openjdk16     2d02c13096c9        24 minutes ago      496MB
ace-minimal      12.0.2.0-alpine-openjdk14     5c1d593ee96f        25 minutes ago      506MB
ace-minimal      12.0.2.0-alpine               6775ce85b5fd        27 minutes ago      604MB
ace-minimal      12.0.2.0-ubuntu               a351cfebbd4d        26 minutes ago      684MB
ace-basic        12.0.2.0-ubuntu               319227027474        19 minutes ago      1.48GB
ace-full         12.0.2.0-ubuntu               73978ff4c598        20 minutes ago      3.02GB
```

Most of these will fit into the IBM Cloud container registry free tier due to compression, but ace-full and ace-basic are too big for that.

