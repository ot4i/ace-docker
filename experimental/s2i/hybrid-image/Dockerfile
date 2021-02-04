FROM ace-minimal:11.0.0.11-alpine

LABEL maintainer="Trevor Dolby"

ENV BUILDER_VERSION="0.0.1" LANG="en_US.UTF-8"

LABEL io.k8s.description="App Connect Enterprise S2I Hybrid build/runtime Image" \
      io.k8s.display-name="App Connect Enterprise S2I Hybrid" \
      io.openshift.tags="builder,runtime,ace" \
      io.openshift.s2i.scripts-url=image:///usr/local/s2i \
      io.s2i.scripts-url=image:///usr/local/s2i \
      io.openshift.s2i.destination="/tmp"

ENV STI_SCRIPTS_PATH="/usr/local/s2i" \ 
    WORKDIR="/usr/local/workdir" \
    S2I_DESTINATION="/tmp" 

COPY ./s2i/bin/ /usr/local/s2i

# Buildah seems to look here
COPY ./s2i/bin/ /usr/libexec/s2i

## openshift gets confused by entrypoints
# ENTRYPOINT ["/usr/local/s2i/run"]
# ENTRYPOINT [""]
