#!/bin/bash -ex


APP_NAME="ace-docker"
BASE_OS="ubi"

# ======================================"
echo "Preparing to build $APP_NAME"
# ======================================"

# unset DOCKER_HOST

# The version of docker used in this build now requires us to explicitly
# add our docker registry as a valid one.
export operatorregistry=appconnect-docker-local.artifactory.swg-devops.com/operator

case $(uname -i) in
  s390x)
    ARCHITECTURE=s390x
    PLATFORM=zLinux
	  HELM_TAR_FILENAME=helm-linux-s390x-v2.13.1.tar.gz
    IFIX_LIST=$(echo $IFIX_LIST_S390X | tr ' ' ',' | sed 's|,$||g')
    ;;
  x86_64)
    ARCHITECTURE=amd64
    PLATFORM=xLinux
    HELM_TAR_FILENAME=helm-linux-amd64-v2.12.3.tar.gz
    IFIX_LIST=$(echo $IFIX_LIST_AMD64 | tr ' ' ',' | sed 's|,$||g')
    ;;
esac


# Logging into Artifactory"
docker login -u ${ARTIFACTORY_USER} -p ${ARTIFACTORY_PASS} appconnect-docker-local.artifactory.swg-devops.com

# tag is timestamp, e.g. "20150917-154801"
TAG=${BUILD_TIMESTAMP}


# Copy in Prod licenses
rm -rf licenses/*
cp -rf ${WORKSPACE}/hip-pipeline-common/onprem/licenses-prod/* licenses/

# Get OpenTracing files
cd ${WORKSPACE}/deps
if [ "${MONTIER_USER_EXIT}" = "true" ]; then
  cd ${WORKSPACE}/deps/OpenTracing/library
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/ACEOpenTracingUserExit.lel
  cd ${WORKSPACE}/deps/OpenTracing/config
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/config.yaml
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/loggers.properties
fi


cd ${WORKSPACE}/deps
# Get ACE image
curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iiboc/prereqs/builds/$ACE_FOLDER/${PLATFORM}/$ACE_BUILD
# Get MQ image
curl -H 'X-JFrog-Art-Api:$ARTIFACTORY_PASS' -O "https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD"


prodimage=ace-server-prod

# build the Prod images
echo "Building the Prod ACE only image"
cd ${WORKSPACE}
docker build --no-cache -t ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f $BASE_OS/Dockerfile.aceonly --build-arg ACE_INSTALL=$ACE_BUILD --build-arg IFIX_LIST=$IFIX_LIST .

echo "Building the Prod ACE with connectors image"
docker build -t ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f ${BASE_OS}/Dockerfile.connectors --build-arg BASE_IMAGE=${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} .

echo "Building the Prod ACE with MQ client image"
docker build -t ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f ${BASE_OS}/Dockerfile.mqclient --build-arg BASE_IMAGE=${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg MQ_URL=https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD  --build-arg MQ_URL_USER=$ARTIFACTORY_USER  --build-arg MQ_URL_PASS=$ARTIFACTORY_PASS .
docker push ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}

if [ "${BRANCH_TO_BUILD}" == "master" ]; then
  docker tag ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}  ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}
  docker push ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}
fi

if [ "${ARCHITECTURE}" == "s390x" ]; then
  echo "Production zlinux images built"
  exit 0
fi

if [ -z "$ACE_DEV_BUILD" ]; then
  echo "No Developer Edition supplied so skipping building Dev Edition images"
  exit 0
fi 

##
## Now building amd64 dev images. 
##

# Copy in Dev licenses
rm -rf licenses/*
cp -rf ${WORKSPACE}/hip-pipeline-common/onprem/licenses-dev/* licenses/

cd ${WORKSPACE}/deps
curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iiboc/prereqs/builds/$ACE_DEV_FOLDER/${PLATFORM}/$ACE_DEV_BUILD

# Get MQ image
curl -H 'X-JFrog-Art-Api:$ARTIFACTORY_PASS' -O "https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD"
cd ${WORKSPACE}


# set the image name in a variable
devimage=ace-server

# build the Dev images
echo "Building the Dev ACE only image"
docker build --no-cache -t ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f ${BASE_OS}/Dockerfile.aceonly --build-arg ACE_INSTALL=$ACE_DEV_BUILD --build-arg IFIX_LIST=$IFIX_LIST .

echo "Building the Dev ACE with connectors image"
docker build -t ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f ${BASE_OS}/Dockerfile.connectors --build-arg BASE_IMAGE=${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} .

echo "Building the Dev ACE with MQ client image"
docker build -t ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} -f ${BASE_OS}/Dockerfile.mqclient --build-arg BASE_IMAGE=${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg MQ_URL=https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD  --build-arg MQ_URL_USER=$ARTIFACTORY_USER  --build-arg MQ_URL_PASS=$ARTIFACTORY_PASS .
docker push ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}

if [ "${BRANCH_TO_BUILD}" == "master" ]; then
  docker tag ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}  ${operatorregistry}/${devimage}:latest-${ARCHITECTURE}
  docker push ${operatorregistry}/${devimage}:latest-${ARCHITECTURE}
fi
