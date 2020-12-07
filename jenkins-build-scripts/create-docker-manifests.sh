#!/bin/bash -ex

edition=$1

export operatorregistry=appconnect-docker-local.artifactory.swg-devops.com/operator

docker version


# Logging into registries
docker login -u ${ARTIFACTORY_USER} -p ${ARTIFACTORY_PASS} appconnect-docker-local.artifactory.swg-devops.com


prodimage="ace-server-prod"
devimage="ace-server"

TAG=${TAG_VERSION}-${BUILD_TIMESTAMP}


ARCH_ARRAY=()
PROD_IMAGE_STRING=
PROD_LATEST_IMAGE_STRING=
DEV_IMAGE_STRING=
DEV_LATEST_IMAGE_STRING=
ZPROD_IMAGE_STRING=
ZPROD_LATEST_IMAGE_STRING=
if [ "$BUILD_PLATFORM" == "amd64-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
  echo Adding amd64 to array of images to build into manifest
  ARCH_ARRAY+=('amd64')
  PROD_IMAGE_STRING+=" ${operatorregistry}/${prodimage}:$TAG-amd64"
  PROD_LATEST_IMAGE_STRING+=" ${operatorregistry}/${prodimage}:latest-amd64"
  DEV_IMAGE_STRING+=" ${operatorregistry}/${devimage}:$TAG-amd64"
  DEV_LATEST_IMAGE_STRING+=" ${operatorregistry}/${devimage}:latest-amd64"
fi
if [ "$BUILD_PLATFORM" == "s390x-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
  echo Adding s390x to array of images to build into manifest
  ARCH_ARRAY+=('s390x')
  ZPROD_IMAGE_STRING+=" ${operatorregistry}/${prodimage}:$TAG-s390x"
  ZPROD_LATEST_IMAGE_STRING+=" ${operatorregistry}/${prodimage}:latest-s390x"
fi
for arch in "${ARCH_ARRAY[@]}"
do
  echo Pulling ${operatorregistry}/${prodimage}:$TAG-$arch locally
  echo Pulling ${operatorregistry}/${devimage}:$TAG-$arch locally

  docker pull ${operatorregistry}/${prodimage}:$TAG-$arch
  if [ "$arch" == "amd64" ] ; then
    docker pull ${operatorregistry}/${devimage}:$TAG-$arch
  fi
  if [ "${BRANCH_TO_BUILD}" == "master" ] ; then
    echo Pulling ${operatorregistry}/${prodimage}:latest-$arch locally
    echo Pulling ${operatorregistry}/${devimage}:latest-$arch locally

    docker pull ${operatorregistry}/${prodimage}:latest-$arch
    if [ "$arch" == "amd64" ] ; then
      docker pull ${operatorregistry}/${devimage}:latest-$arch
    fi
  fi

  echo Finished pulling $arch
done


if [ "$BUILD_PLATFORM" == "amd64-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
  sudo docker manifest create --amend ${operatorregistry}/${prodimage}:$TAG $PROD_IMAGE_STRING
  sudo docker manifest create --amend ${operatorregistry}/${devimage}:$TAG $DEV_IMAGE_STRING
fi
if [ "$BUILD_PLATFORM" == "s390x-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
  sudo docker manifest create --amend ${operatorregistry}/${prodimage}:$TAG $ZPROD_IMAGE_STRING
fi
if [ "${BRANCH_TO_BUILD}" == "master" ] ; then
  if [ "$BUILD_PLATFORM" == "amd64-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
    echo "INFO: creating manifests with 'latest' tag"
    sudo docker manifest create --amend ${operatorregistry}/${prodimage}:latest $PROD_LATEST_IMAGE_STRING
    sudo docker manifest create --amend ${operatorregistry}/${devimage}:latest $DEV_LATEST_IMAGE_STRING
    sudo docker manifest inspect ${operatorregistry}/${devimage}:latest
  fi
  if [ "$BUILD_PLATFORM" == "s390x-only" ] || [ "$BUILD_PLATFORM" == "both" ] ; then
    echo "INFO: Adding zlinux image to prod manifest that has the 'latest' tag"
    sudo docker manifest create --amend ${operatorregistry}/${prodimage}:latest $ZPROD_LATEST_IMAGE_STRING
    sudo docker manifest inspect ${operatorregistry}/${prodimage}:latest
  fi
fi

for arch in "${ARCH_ARRAY[@]}"
do
  echo "Annotating manifest for $arch"
  sudo docker manifest annotate ${operatorregistry}/${prodimage}:$TAG ${operatorregistry}/${prodimage}:$TAG-$arch --os linux --arch $arch
  if [ "$arch" == "amd64" ] ; then
    sudo docker manifest annotate ${operatorregistry}/${devimage}:$TAG ${operatorregistry}/${devimage}:$TAG-$arch --os linux --arch $arch
  fi
  if [ "${BRANCH_TO_BUILD}" == "master" ] ; then
    echo "INFO: Annotating image entry in 'latest' prod manifest for $arch"    
    sudo docker manifest annotate ${operatorregistry}/${prodimage}:latest ${operatorregistry}/${prodimage}:latest-$arch --os linux --arch $arch
    sudo docker manifest inspect ${operatorregistry}/${prodimage}:latest
    if [ "$arch" == "amd64" ]; then
      echo "INFO: Annotating image entry in 'latest' dev manifest for $arch"
      sudo docker manifest annotate ${operatorregistry}/${devimage}:latest ${operatorregistry}/${devimage}:latest-$arch --os linux --arch $arch
      sudo docker manifest inspect ${operatorregistry}/${devimage}:latest
    fi
  fi
done

sudo docker manifest push ${operatorregistry}/${prodimage}:$TAG --purge
skopeo copy --all docker://${operatorregistry}/${prodimage}:$TAG docker://${operatorregistry}/${prodimage}:$TAG
for arch in "${ARCH_ARRAY[@]}"
do
  if [ "$arch" == "amd64" ] ; then
    sudo docker manifest push ${operatorregistry}/${devimage}:$TAG --purge
    skopeo copy --all docker://${operatorregistry}/${devimage}:$TAG docker://${operatorregistry}/${devimage}:$TAG
  fi
done

if [ "${BRANCH_TO_BUILD}" == "master" ] ; then
  sudo docker manifest push ${operatorregistry}/${prodimage}:latest --purge
  skopeo copy --all docker://${operatorregistry}/${prodimage}:latest docker://${operatorregistry}/${prodimage}:latest
  for arch in "${ARCH_ARRAY[@]}"
  do
    if [ "$arch" == "amd64" ] ; then
      sudo docker manifest push ${operatorregistry}/${devimage}:latest --purge
      skopeo copy --all docker://${operatorregistry}/${devimage}:latest docker://${operatorregistry}/${devimage}:latest
    fi
  done
fi
