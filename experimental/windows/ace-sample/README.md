# ACE sample container

This directory contains the files required to build a Windows container with a sample
ACE application that runs automatically when the container is started. This container
is built on top of the [ace-basic](../ace-basic) container.

## Building and running the sample

Build as follows: 
```
docker build -t ace-sample:12.0.10.0-windows  .
```

To run the sample after building:
```
docker run --rm -ti -p 7800:7800 ace-sample:12.0.10.0-windows
```
and then `curl http://localhost:7800/test` should return `{"data":"a string from ACE"}`
