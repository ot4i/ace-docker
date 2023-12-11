# ace-minimal docker image

This directory contains files that allow small ACE container images to be
built for use in pipelines and cloud container runtimes.

Main Dockerfiles:

- Dockerfile.alpine builds the smallest image, using Alpine as a base with glibc added
- Dockerfile.ubuntu builds a slightly larger image

## Building

See the [README](../README.md) in the parent directory for download URL update
information (if needed) and then run one of the commands
```
export DOWNLOAD_URL=<someURL>
docker build --build-arg DOWNLOAD_URL -t ace-minimal:12.0.10.0-alpine -f Dockerfile.alpine .
```
or
```
export DOWNLOAD_URL=<someURL>
docker build --build-arg DOWNLOAD_URL -t ace-minimal:12.0.10.0-ubuntu -f Dockerfile.ubuntu .
```

These images can then be used to build other images with build tools, deployed
ACE applications, etc.

