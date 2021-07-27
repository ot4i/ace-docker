FROM golang:latest as builder

WORKDIR /go/src/github.com/ot4i/ace-docker/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY common/ ./common
RUN go version
RUN go build -ldflags "-X \"main.ImageCreated=$(date --iso-8601=seconds)\"" ./cmd/runaceserver/
RUN go build ./cmd/chkaceready/
RUN go build ./cmd/chkacehealthy/

# Run all unit tests
RUN go test -v ./cmd/runaceserver/
RUN go test -v ./internal/...
RUN go test -v ./common/...
RUN go vet ./cmd/... ./internal/... ./common/...

ARG ACE_INSTALL=ace-12.0.1.0.tar.gz
ARG IFIX_LIST=""
WORKDIR /opt/ibm
COPY deps/$ACE_INSTALL .
COPY ./ApplyIFixes.sh /opt/ibm
RUN mkdir ace-12
RUN tar -xzf $ACE_INSTALL --absolute-names --exclude ace-12.\*/tools --exclude ace-12.\*/server/bin/TADataCollector.sh --exclude ace-12.\*/server/transformationAdvisor/ta-plugin-ace.jar --strip-components 1 --directory /opt/ibm/ace-12 \
  && ./ApplyIFixes.sh $IFIX_LIST \ 
  && rm ./ApplyIFixes.sh

FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV SUMMARY="Integration Server for App Connect Enterprise" \
  DESCRIPTION="Integration Server for App Connect Enterprise" \
  PRODNAME="AppConnectEnterprise" \
  COMPNAME="IntegrationServer"

LABEL summary="$SUMMARY" \
  description="$DESCRIPTION" \
  io.k8s.description="$DESCRIPTION" \
  io.k8s.display-name="Integration Server for App Connect Enterprise" \
  io.openshift.tags="$PRODNAME,$COMPNAME" \
  com.redhat.component="$PRODNAME-$COMPNAME" \
  name="$PRODNAME/$COMPNAME" \
  vendor="IBM" \
  version="REPLACE_VERSION" \
  release="REPLACE_RELEASE" \
  license="IBM" \
  maintainer="Hybrid Integration Platform Cloud" \
  io.openshift.expose-services="" \
  usage=""

# Add required license as text file in Liceses directory (GPL, MIT, APACHE, Partner End User Agreement, etc)
COPY /licenses/ /licenses/

RUN microdnf update && microdnf install findutils util-linux unzip python3 tar procps openssl && microdnf clean all \
   && ln -s /usr/bin/python3 /usr/local/bin/python \
   && mkdir /etc/ACEOpenTracing /opt/ACEOpenTracing /var/log/ACEOpenTracing && chmod 777 /var/log/ACEOpenTracing /etc/ACEOpenTracing

# Force reinstall tzdata package to get zoneinfo files
RUN microdnf reinstall tzdata -y

# Create OpenTracing directories, update permissions and copy in any library or configuration files needed
COPY deps/OpenTracing/library/* ./opt/ACEOpenTracing/
COPY deps/OpenTracing/config/* ./etc/ACEOpenTracing/

WORKDIR /opt/ibm

COPY --from=builder /opt/ibm/ace-12 /opt/ibm/ace-12

# Copy in PID1 process
COPY --from=builder /go/src/github.com/ot4i/ace-docker/runaceserver /usr/local/bin/
COPY --from=builder /go/src/github.com/ot4i/ace-docker/chkace* /usr/local/bin/

# Copy in script files
COPY *.sh /usr/local/bin/

# Install kubernetes cli
COPY ubi/install-kubectl.sh /usr/local/bin/
RUN chmod u+x /usr/local/bin/install-kubectl.sh \
  && install-kubectl.sh 

COPY ubi/generic_invalid/invalid_license.msgflow /home/aceuser/temp/gen
COPY ubi/generic_invalid/InvalidLicenseJava.jar /home/aceuser/temp/gen
COPY ubi/generic_invalid/application.descriptor /home/aceuser/temp

# Create a user to run as, create the ace workdir, and chmod script files
RUN /opt/ibm/ace-12/ace make registry global accept license silently \ 
  && useradd -u 1000 -d /home/aceuser -G mqbrkrs,wheel aceuser \
  && mkdir -p /var/mqsi \
  && mkdir -p /home/aceuser/initial-config \
  && su - -c '. /opt/ibm/ace-12/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server' \
  && chmod -R 777 /home/aceuser \
  && chmod -R 777 /var/mqsi \
  && su - -c '. /opt/ibm/ace-12/server/bin/mqsiprofile && echo $MQSI_JREPATH && chmod g+w $MQSI_JREPATH/lib/security/cacerts' \
  && chmod -R 777 /home/aceuser/temp \
  && chmod 777 /opt/ibm/ace-12/server/ODBC/dsdriver/odbc_cli/clidriver/license

COPY git.commit /home/aceuser/

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/usr/local/bin/ace_env.sh

# Expose ports.  7600, 7800, 7843 for ACE; 9483 for ACE metrics
EXPOSE 7600 7800 7843 9483

WORKDIR /home/aceuser

ENV LOG_FORMAT=basic

# Set user to prevent container running as root by default
USER 1000

# Set entrypoint to run management script

ENTRYPOINT ["runaceserver"]
