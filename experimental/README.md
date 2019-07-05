# Experimental docker files

Simple docker images for ACE v11 on Linux (amd64 and s390x) and Windows

ace-minimal-install contains the Dockerfiles for building images containing a minimal ACE v11 install.
ace-minimal contains the Dockerfiles for building images that can run simple servers with a non-root user.
ace-sample contains a sample BAR file and Dockerfiles for building runnable images to serve HTTP clients.

See build-all.sh for details on building the images; setting LICENSE=accept is required for all but the initial image builds.

To run the sample after building:

docker run -e LICENSE=accept --rm -ti ace-sample-alpine

and then curl http://<container IP>:7800/test should return '{"data":"a string from ACE"}'
