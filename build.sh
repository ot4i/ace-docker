#!/bin/sh
echo "Building ACE build container"
buildType=$1

if [ -z "$2" ]
then
      echo "Building with default ACE install parameters"
      docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
        -t ace/builder:11.0.0.2 . -f ./rhel/Dockerfile.build
else
      echo "Building with ACE install $1"
      docker build --build-arg ACE_INSTALL=$2 --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy \
        -t ace/builder:11.0.0.2 . -f ./rhel/Dockerfile.build
fi

docker create --name builder ace/builder:11.0.0.2
docker cp builder:/opt/ibm/ace-11 ./rhel/ace-11
docker cp builder:/go/src/github.com/ot4i/ace-docker/runaceserver ./rhel/runaceserver
docker cp builder:/go/src/github.com/ot4i/ace-docker/chkaceready ./rhel/chkaceready
docker cp builder:/go/src/github.com/ot4i/ace-docker/chkacehealthy ./rhel/chkacehealthy
docker rm -f builder

echo "Building ACE runtime container"

case $buildType in
"ace-dev-only")
   echo "Building ACE only for development"
   docker build -t ace/ace-dev-only -f ./rhel/Dockerfile.acerhel .
   ;;
"ace-only")
   echo "Building ACE only for production"
   docker build -t ace/ace-only -f ./rhel/Dockerfile.acerhel .
   ;;
"ace-mq")
   echo "Building ACE with MQ for production"
   docker build -t ace/ace-mq --build-arg BASE_IMAGE=$3 -f ./rhel/Dockerfile.acemqrhel .
   ;;
"ace-dev-mq-dev")
   echo "Building ACE with MQ for production"
   docker build -t ace/ace-dev-mq-dev --build-arg BASE_IMAGE=$3 -f ./rhel/Dockerfile.acemqrhel .
  ;;
*) echo "Invalid option"
   ;;
esac
