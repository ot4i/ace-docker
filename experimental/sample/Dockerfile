# Override for other variants
ARG BASE_IMAGE=ace-minimal:12.0.4.0-ubuntu
FROM $BASE_IMAGE

ARG LICENSE

# docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-full:12.0.4.0-ubuntu -t ace-sample:12.0.4.0-full-ubuntu .
# docker run -e LICENSE=accept -p 7800:7800 --rm -ti ace-sample:12.0.4.0-full-ubuntu

# Switch off the admin REST API for the server run, as we won't be deploying anything after start
RUN sed -i 's/#port: 7600/port: -1/g' /home/aceuser/ace-server/server.conf.yaml 

COPY aceFunction.bar /tmp/aceFunction.bar
RUN bash -c "export LICENSE=${LICENSE} && . /home/aceuser/.bashrc && mqsibar -c -a /tmp/aceFunction.bar -w /home/aceuser/ace-server"

RUN bash -c "export LICENSE=${LICENSE} && . /home/aceuser/.bashrc && ibmint optimize server --work-dir /home/aceuser/ace-server"

# Set entrypoint to run the server
ENTRYPOINT ["bash", "-c", "IntegrationServer -w /home/aceuser/ace-server --no-nodejs"]
