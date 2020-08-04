# Experimental docker files

Simple docker images for ACE v11 on Linux (amd64 and s390x) and Windows

ace-full contains a Dockerfile for building an image that can run the full product, including mqsicreatebar with a virtual X server.

ace-minimal contains the Dockerfiles for building images that can run simple servers with a non-root user.

ace-sample contains a sample BAR file and Dockerfiles for building runnable images to serve HTTP clients.


See build-all.sh for details on building the images; setting LICENSE=accept is required for all but the initial image builds.

To run the sample after building:
```
docker run -e LICENSE=accept --rm -ti ace-sample:11.0.0.9-minimal-alpine
```
and then curl http://<container IP>:7800/test should return '{"data":"a string from ACE"}'

## Various sizes
Local on kenya.hursley.uk.ibm.com (debian 10) with defaults in Dockerfiles:

```
ace-minimal:11.0.0.9-alpine-openjdk14   a224eb57078c        39 seconds ago      682MB
ace-minimal:11.0.0.9-ubuntu             5541e188b068        6 minutes ago       896MB
ace-minimal:11.0.0.9-alpine             c906eaa3ba7e        25 minutes ago      524MB
ace-full:11.0.0.9-ubuntu                e7a8e54f20cf        4 minutes ago       2.48GB
```

Note that the first two have the web UI available on port 7600; removing that capability would leave them at

```
ace-minimal:11.0.0.9-alpine-openjdk14   eabf8622343a        12 minutes ago      462MB
ace-minimal:11.0.0.9-ubuntu             a2cfcf555038        18 minutes ago      676MB
```

Most of these will fit into the IBM Cloud container registry free tier due to compression, but ace-full is too big for that.
