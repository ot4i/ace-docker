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
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" >/dev/null
    nvm install v12.12.0
    nvm use v12.12.0
    ;;
  x86_64)
    ARCHITECTURE=amd64
    PLATFORM=xLinux
	HELM_TAR_FILENAME=helm-linux-amd64-v2.12.3.tar.gz
    ;;
esac

cd ot4i-ace-docker
echo ${ARTIFACTORY_USER} | base64
echo ${ARTIFACTORY_PASS} | base64
# Logging into Artifactory"
docker login -u ${ARTIFACTORY_USER} -p ${ARTIFACTORY_PASS} appconnect-docker-local.artifactory.swg-devops.com

# tag is timestamp, e.g. "20150917-154801"
TAG=${BUILD_TIMESTAMP}

# or "20150917-154801-personal" for personal builds
if [ "$PERSONAL_BUILD" = "true" ]; then
  TAG="$TAG-personal"
  echo "Your personal build tag will be: ${TAG}"
fi

awk -v my_image=${APP_NAME} -v my_tag=${TAG} '/FROM \$BASE_IMAGE/{ print; print "RUN echo \"" my_image ":" my_tag" \" >/etc/debian_chroot"; next }1' $BASE_OS/Dockerfile.aceonly > $BASE_OS/Dockerfile.aceonly_tmp &&
  mv $BASE_OS/Dockerfile.aceonly_tmp $BASE_OS/Dockerfile.aceonly

# Copy in Prod licenses
rm -rf licenses/*
rm LICENSE
cp -rf ../hip-pipeline-common/onprem/licenses-prod/* licenses/
mv licenses/LICENSE LICENSE

# Get OpenTracing files
cd deps
if [ "${MONTIER_USER_EXIT}" = "true" ]; then
  cd OpenTracing/library
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/ACEOpenTracingUserExit.lel
  cd ../config
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/config.yaml
  curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/icp-prereqs/OpenTracing/loggers.properties
  cd ../..
fi

# Get ACE image
curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iiboc/prereqs/builds/$ACE_FOLDER/${PLATFORM}/$ACE_BUILD
cd ..
 cd deps
# Get MQ image
curl -H 'X-JFrog-Art-Api:$ARTIFACTORY_PASS' -O "https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD"
cd ..

image=interim-prod-build
prodimage=ace-server-prod

# build the Prod images
echo "Building the Prod ACE only image"
docker build -f $BASE_OS/Dockerfile.aceonly --build-arg ACE_INSTALL=$ACE_BUILD --tag ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} .

echo "Building the Prod ACE with MQ client image"
docker build -t ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg BASE_IMAGE=${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg MQ_URL=https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD  --build-arg MQ_URL_USER=$ARTIFACTORY_USER  --build-arg MQ_URL_PASS=$ARTIFACTORY_PASS -f $BASE_OS/Dockerfile.mqclient .
docker push ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}

if [ "${BRANCH_TO_BUILD}" == "master" ]; then
  docker tag ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}  ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}
  docker push ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}
fi
# Only build dev edition on amd64. Does not exist on other platforms
if [ "${ARCHITECTURE}" == "amd64" ]; then

  if [ -z "$ACE_DEV_BUILD" ]
  then
      echo "No Developer Edition supplied so skipping building Dev Edition images"
  else
    # Copy in Dev licenses
    rm -rf licenses/*
    rm LICENSE
    cp -rf ../hip-pipeline-common/onprem/licenses-dev/* licenses/
    mv licenses/LICENSE LICENSE

    cd deps
    curl -O --user $ARTIFACTORY_USER:$ARTIFACTORY_PASS https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iiboc/prereqs/builds/$ACE_DEV_FOLDER/${PLATFORM}/$ACE_DEV_BUILD
    cd ..
    cd deps
    # Get MQ image
    curl -H 'X-JFrog-Art-Api:$ARTIFACTORY_PASS' -O "https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD"
    cd ..


    # set the image name in a variable
    image=interim-dev-build
    devimage=ace-server

    # build the Dev images
    echo "Building the Dev ACE only image"
    docker build -f $BASE_OS/Dockerfile.aceonly --build-arg ACE_INSTALL=$ACE_DEV_BUILD --tag ${operatorregistry}/${image}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} .

    echo "Building the Dev ACE with MQ client image"
    docker build -t ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg BASE_IMAGE=${operatorregistry}/${image}:${TAG_VERSION}-${TAG}-${ARCHITECTURE} --build-arg MQ_URL=https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iib/operator/prereqs/builds/$MQ_FOLDER/${PLATFORM}/$MQ_BUILD  --build-arg MQ_URL_USER=$ARTIFACTORY_USER  --build-arg MQ_URL_PASS=$ARTIFACTORY_PASS -f $BASE_OS/Dockerfile.mqclient .
    docker push ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}

    if [ "${BRANCH_TO_BUILD}" == "master" ]; then
      docker tag ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}  ${operatorregistry}/${devimage}:latest-${ARCHITECTURE}
      docker push ${operatorregistry}/${devimage}:latest-${ARCHITECTURE}
    fi

  fi


  sudo pkill dockerd
  sleep 10

  set +x
  echo ""
  echo "Images pushed:"
  if [ -z "$ACE_DEV_BUILD" ]
  then
	 echo "   * ${operatorregistry}/${prodimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}"
     if [ "${BRANCH_TO_BUILD}" == "master" ]; then
		echo "   * ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}"
     fi
  else
	 echo "   * ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}"
     echo "   * ${operatorregistry}/${devimage}:${TAG_VERSION}-${TAG}-${ARCHITECTURE}"
     if [ "${BRANCH_TO_BUILD}" == "master" ]; then
       echo "   * ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}"
       echo "   * ${operatorregistry}/${prodimage}:latest-${ARCHITECTURE}"
     fi
  fi
fi
