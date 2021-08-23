# Experimental docker files

Simple docker images for ACE v12 on Linux (amd64 and s390x) and Windows

ace-full contains a Dockerfile for building an image that can run the full product, including mqsicreatebar with a virtual X server.

ace-basic contains a Dockerfile for building an image that can run the product server, including all files except the toolkit.

ace-minimal contains the Dockerfiles for building images that can run simple servers with a non-root user.

ace-sample contains a sample BAR file and Dockerfiles for building runnable images to serve HTTP clients.

See build-all.sh for details on building the images; setting LICENSE=accept is required for all but the initial image builds.

To run the sample after building:
```
docker run -e LICENSE=accept --rm -ti ace-sample:12.0.1.0-minimal-alpine
```
and then curl http://[container IP]:7800/test should return '{"data":"a string from ACE"}'

## Various sizes
Local on kenya.hursley.uk.ibm.com (debian 10) with defaults in Dockerfiles:

```
ace-minimal      12.0.1.0-alpine-openjdk16     e7ed561a933e        2 minutes ago       517MB
ace-minimal      12.0.1.0-alpine-openjdk14     6c0eef3c4116        2 minutes ago       527MB
ace-minimal      12.0.1.0-alpine               1d68aaf565fd        4 minutes ago       628MB
ace-minimal      12.0.1.0-ubuntu               874719904be1        5 hours ago         706MB
ace-basic        12.0.1.0-ubuntu               66fdbbf9b010        4 hours ago         1.51GB
ace-full         12.0.1.0-ubuntu               22a2f46d7b31        4 hours ago         3.05GB
```

Most of these will fit into the IBM Cloud container registry free tier due to compression, but ace-full and ace-basic are too big for that.
