FROM ace-minimal:12.0.1.0-ubuntu

LABEL maintainer="Trevor Dolby"

# docker build -t tdolby/experimental:ace-s2i-hybrid-image .
# docker push tdolby/experimental:ace-s2i-hybrid-image

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

USER 0

RUN cd /opt && \
    curl -k https://apache.mirrors.nublue.co.uk/maven/maven-3/3.6.3/binaries/apache-maven-3.6.3-bin.tar.gz | tar -xzf - && \
    ln -s /opt/apache-maven-3.6.3/bin/mvn /usr/local/bin/mvn
    
# aceuser
USER 1001

# This is needed because "oc new-app" seems to have issues propagating env vars . . .
ENV LICENSE accept

## openshift gets confused by entrypoints
# ENTRYPOINT ["/usr/local/s2i/run"]
#ENTRYPOINT [""]
