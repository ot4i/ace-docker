# FROM ace-minimal:12.0.1.0-alpine-openjdk14
FROM ace-minimal:12.0.1.0-alpine

# docker build -t ace-s2i-runtime-image .

# docker build -t tdolby/experimental:ace-s2i-runtime-image .
# docker push tdolby/experimental:ace-s2i-runtime-image

LABEL maintainer="Trevor Dolby"

LABEL io.k8s.description="App Connect Enterprise S2I Runtime Image" \
      io.k8s.display-name="App Connect Enterprise S2I Runtime" \
      io.openshift.tags="ace,appconnect,minimal" \
      io.openshift.s2i.scripts-url=image:///usr/local/s2i \
      io.s2i.scripts-url=image:///usr/local/s2i \
      io.openshift.expose-services="7800/tcp:http, 7843/tcp:https" \
      io.openshift.s2i.destination="/tmp"

WORKDIR "/tmp"

ENV STI_SCRIPTS_PATH="/usr/local/s2i" \
    WORKDIR="/usr/local/workdir" \
    S2I_DESTINATION="/tmp"

COPY ./s2i/bin/ /usr/local/s2i

USER root

RUN chmod +x /usr/local/s2i/assemble && \
    chmod +x /usr/local/s2i/assemble-runtime && \
    chmod +x /usr/local/s2i/run && \
    chmod +x /usr/local/s2i/usage

USER 1001

# Default setting for the verbose option
# ARG VERBOSE=false
ENTRYPOINT [""]
CMD ["/usr/local/s2i/run"]
