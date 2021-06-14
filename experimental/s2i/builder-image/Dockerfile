FROM ace-full:12.0.1.0-ubuntu

# docker build -t ace-s2i-builder-image .
#
# docker build -t tdolby/experimental:ace-s2i-builder-image .
# docker push tdolby/experimental:ace-s2i-builder-image

LABEL maintainer="Trevor Dolby"

ENV BUILDER_VERSION="0.0.1" LANG="en_US.UTF-8"

LABEL io.k8s.description="App Connect Enterprise S2I Builder Image" \
      io.k8s.display-name="App Connect Enterprise S2I Builder " \
      io.openshift.tags="builder" \
      io.openshift.s2i.scripts-url=image:///usr/local/s2i \
      io.s2i.scripts-url=image:///usr/local/s2i \
      io.openshift.s2i.destination="/tmp"

ENV STI_SCRIPTS_PATH="/usr/local/s2i" \ 
    WORKDIR="/usr/local/workdir" \
    S2I_DESTINATION="/tmp" 

USER root
RUN apt-get update && apt-get install -y maven

# aceuser
USER 1001

COPY ./s2i/bin/ /usr/local/s2i

ENTRYPOINT [""]
