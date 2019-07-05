#!/bin/bash -ex
echo "Building ACE build container"
buildType=$1
buildTag=$2
aceInstall=$3
mqImage=$4


if [ -z "$aceInstall" ]
then
      echo "Building temporary container with default ACE install parameters"
      docker build --build-arg  -t ace/builder:11.0.0.4 . -f ./rhel/Dockerfile.build
else
      echo "Building temporary container with ACE install $buildType"
      docker build --build-arg ACE_INSTALL=$aceInstall -t ace/builder:11.0.0.4 . -f ./rhel/Dockerfile.build
fi

docker create --name builder ace/builder:11.0.0.4
docker cp builder:/opt/ibm/ace-11 ./rhel/ace-11
docker cp builder:/go/src/github.com/ot4i/ace-docker/runaceserver ./rhel/runaceserver
docker cp builder:/go/src/github.com/ot4i/ace-docker/chkaceready ./rhel/chkaceready
docker cp builder:/go/src/github.com/ot4i/ace-docker/chkacehealthy ./rhel/chkacehealthy
docker rm -f builder

echo "Building ACE runtime container"

# Replace the FROM statement to use the MQ container
sed -i "s%^FROM .*%FROM $mqImage%" ./rhel/Dockerfile.acemqrhel

case $buildType in
"ace-dev-only")
   echo "Building ACE only for development"
   docker build -t $buildTag -f ./rhel/Dockerfile.acerhel .
   ;;
"ace-only")
   echo "Building ACE only for production"
   docker build -t $buildTag -f ./rhel/Dockerfile.acerhel .
   ;;
"ace-mq")
   echo "Building ACE with MQ for production"
   docker build -t $buildTag --build-arg BASE_IMAGE=$mqImage -f ./rhel/Dockerfile.acemqrhel .
   ;;
"ace-dev-mq-dev")
   echo "Building ACE with MQ for production"
   docker build -t $buildTag --build-arg BASE_IMAGE=$mqImage -f ./rhel/Dockerfile.acemqrhel .
  ;;
*) echo "Invalid option"
   ;;
esac
