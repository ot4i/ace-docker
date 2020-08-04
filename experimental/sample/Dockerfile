ARG BASE_IMAGE=ace-minimal:11.0.0.9-ubuntu
FROM $BASE_IMAGE

ARG LICENSE

# docker build --build-arg LICENSE=accept --build-arg BASE_IMAGE=ace-full:11.0.0.9-ubuntu -t ace-sample:11.0.0.9-full-ubuntu .
# docker run -e LICENSE=accept -p 7800:7800 --rm -ti ace-sample:11.0.0.9-full-ubuntu

# Switch off the admin REST API for the server run, as we won't be deploying anything after start
RUN sed -i 's/#port: 7600/port: -1/g' /home/aceuser/ace-server/server.conf.yaml 

COPY aceFunction.bar /tmp/aceFunction.bar
RUN bash -c "export LICENSE=${LICENSE} && . /home/aceuser/.bashrc && mqsibar -c -a /tmp/aceFunction.bar -w /home/aceuser/ace-server"

# Set entrypoint to run the server
ENTRYPOINT ["bash", "-c", "IntegrationServer -w /home/aceuser/ace-server --no-nodejs"]
