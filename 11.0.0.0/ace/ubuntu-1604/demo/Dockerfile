FROM ace:11.0.0.0

ENV BAR1=Transformation_Map.bar

# Copy in the bar file to a temporary directory
COPY $BAR1 /tmp

# Unzip the BAR file; need to use bash to make the profile work
RUN bash -c 'mqsicreateworkdir /home/aceuser/ace-server && mqsibar -w /home/aceuser/ace-server -a /tmp/$BAR1 -c'

# Switch off the admin REST API for the server run, as we won't be deploying anything after start
RUN sed -i 's/adminRestApiPort/#adminRestApiPort/g' /home/aceuser/ace-server/server.conf.yaml 

# We inherit the command from the base layer
