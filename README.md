# Overview


<img src="./app_connect_light_256x256.png" width="100" alt="IBM ACE logo"/>

Run [IBM® App Connect Enterprise](https://www.ibm.com/cloud/app-connect/enterprise) in a container.

You can build an image containing one of the following combinations:
- IBM App Connect Enterprise
- IBM App Connect Enterprise with IBM MQ Advanced
- IBM App Connect Enterprise for Developers with IBM MQ Advanced for Developers
- IBM App Connect Enterprise for Developers

# Building a container image

Download a copy of App Connect Enterprise (ie. `ace-11.0.0.6.tar.gz`) and place it in the `deps` folder. When building the image use `build-arg` to specify the name of the file: `--build-arg ACE_INSTALL=ace-11.0.0.6.tar.gz`

- **Important:** Only ACE version **11.0.0.6 or greater** is supported.

Choose if you want to have an image with just App Connect Enterprise or an image with both App Connect Enterprise and IBM MQ Advanced.

### Building a container image which contains an IBM Service provided fix for ACE

You may have been provided with a fix for App Connect Enterprise by IBM Support, this fix will have a name of the form `11.0.0.X-ACE-LinuxX64-TF12345.tar.gz`. In order to apply this fix follow these steps.
 - On a local system extract the App Connect Enterprise archive
   `tar -xvf ace-11.0.0.6.tar.gz`
 - Extract the fix package into expanded App Connect Enterprise installation
   `tar -xvf /path/to/11.0.0.6-ACE-LinuxX64-TF12345.tar.gz --directory ace-11.0.0.6`
 - Tar and compress the resulting App Connect Enterprise installation
   `tar -cvf ace-11.0.0.5_with_IT12345.tar ace-11.0.0.6`
   `gzip ace-11.0.0.5_with_IT12345.tar`
 - Place the resulting `ace-11.0.0.5_with_IT12345.tar.gz` file in the `deps` folder and when building using the `build-arg` to specify the name of the file: `--build-arg ACE_INSTALL=ace-11.0.0.5_with_IT12345.tar.gz`

### Using App Connect Enterprise for Developers

Get [ACE for Developers edition](https://www.ibm.com/marketing/iwm/iwm/web/pick.do?source=swg-wmbfd). Then place it in the `deps` folder as mentioned above.

### Using MQ production image

When building an image with both ACE and MQ, the docker file uses the [MQ Advanced for Developers image in docker registry](https://hub.docker.com/r/ibmcom/mq/) as base image.

When building a production image with MQ, follow the [MQ instructions](https://github.com/ibm-messaging/mq-container/blob/master/docs/building.md#building-a-production-image) to build your own production MQ image. Then, when building the ACE with MQ image use `build-arg` to set the `BASE_IMAGE` to your production MQ image. More details below.

## Build an image with App Connect Enterprise and MQ

[Info on how to get the Developers or production image for MQ](#using-mq-production-image)

The `deps` folder must contain a copy of ACE, **version 11.0.0.6 or greater**. If using ACE for Developers, download it from [here](https://www.ibm.com/marketing/iwm/iwm/web/pick.do?source=swg-wmbfd).
Then set the build argument `ACE_INSTALL` to the name of the ACE file placed in `deps`.

1. ACE production with MQ Advanced production
   - `docker build -t ace-mq --build-arg BASE_IMAGE={MQ-image} --build-arg ACE_INSTALL={ACE-file-in-deps-folder} --file ubi/Dockerfile.acemq .`
2. ACE for Developers with MQ Advanced for Developers:
   -`docker build -t ace-dev-mq-dev --build-arg ACE_INSTALL={ACE-dev-file-in-deps-folder} --file ubi/Dockerfile.acemq .`

**Note:** As mentioned before, the docker file will download the **[Development version of IBM MQ](https://hub.docker.com/r/ibmcom/mq/)** by default unless `BASE_IMAGE` is changed.

## Build an image with App Connect Enterprise only

The `deps` folder must contain a copy of ACE, **version 11.0.0.6 or greater**. If using ACE for Developers, download it from [here](https://www.ibm.com/marketing/iwm/iwm/web/pick.do?source=swg-wmbfd).
Then set the build argument `ACE_INSTALL` to the name of the ACE file placed in `deps`.

1. ACE for Developers only:
   - `docker build -t ace-dev-only --build-arg ACE_INSTALL={ACE-dev-file-in-deps-folder} --file ubi/Dockerfile.aceonly .`
2. ACE production only: 
   - `docker build -t ace-only --build-arg ACE_INSTALL={ACE-file-in-deps-folder} --file ubi/Dockerfile.aceonly .`

## Build an image with App Connect Enterprise and MQ Client

Follow the instructions above for building an image with App Connect Enterprise Only.

Add the MQ Client libraries to your existing image by running `docker build -t ace-mqclient --build-arg BASE_IMAGE=<AceOnlyImageTag> --file ubi/Dockerfile.mqclient .`

`<AceOnlyImageTag>` is the tag of the image you want to add the client libs to i.e. ace-only. You can supply a customer URL for the MQ binaries by setting the argument MQ_URL

# Usage

### Accepting the License

In order to use the image, it is necessary to accept the terms of the IBM App Connect Enterprise license. This is achieved by specifying the environment variable `LICENSE` equal to `accept` when running the image. You can also view the license terms by setting this variable to `view`. Failure to set the variable will result in the termination of the container with a usage statement. You can view the license in a different language by also setting the `LANG` environment variable.

### Red Hat OpenShift SecurityContextConstraints Requirements

This chart requires a SecurityContextConstraints to be bound to the target namespace prior to installation. To meet this requirement there may be cluster scoped as well as namespace scoped pre and post actions that need to occur.


#### Running an ACE Only Integration Server

The predefined SecurityContextConstraints name: [`ibm-anyuid-scc`](https://ibm.biz/cpkspec-scc) has been verified for this chart when creating an ACE & MQ integration server, if your target namespace is bound to this SecurityContextConstraints resource you can proceed to install the chart.

Run the following command to add the service account of the Integration server to the anyuid scc - `oc adm policy add-scc-to-user ibm-anyuid-scc system:serviceaccount:<namespace>:<releaseName>-ibm-ace-server-prod-serviceaccount` i.e.
``` 
oc adm policy add-scc-to-user ibm-anyuid-scc system:serviceaccount:default:ace-nomq-ibm-ace-server-rhel-prod-serviceaccount
```

#### Running an ACE & MQ Integration Server

The predefined SecurityContextConstraints name: [`ibm-anyuid-scc`](https://ibm.biz/cpkspec-scc) has been verified for this chart when creating an ACE & MQ integration server, if your target namespace is bound to this SecurityContextConstraints resource you can proceed to install the chart.

Run the following command to add the service account of the Integration server to the anyuid scc. - `oc adm policy add-scc-to-user ibm-anyuid-scc system:serviceaccount:<namespace>:<releaseName>-ibm-ace-server-mq-prod-serviceaccount` i.e.
```
oc adm policy add-scc-to-user ibm-anyuid-scc system:serviceaccount:ace:ace-mq-ibm-ace-server-mq-prod-serviceaccount
```

### ACE & MQ image

To run a container with ACE and MQ with default configuration and these settings:
- queue manager name `QMGR`
- listener for MQ on port `1414`
- ACE server name `ACESERVER`
- listener for ACE web ui on port `7600`
- listener for ACE HTTP on port `7600`
run the following command:

`docker run --name acemqserver -p 7600:7600 -p 7800:7800 -p 7843:7843 -p 1414:1414 --env LICENSE=accept --env MQ_QMGR_NAME=QMGR --env ACE_SERVER_NAME=ACESERVER ace-mq:latest`

Once the console shows that the integration server is listening on port 7600, you can go to the ACE UI at http://localhost:7600/. To stop the container, run `docker stop acemqserver` and the container will shut down cleanly, stopping the integration server, then the queue manager.

### ACE only image

To run a container with ACE only with default configuration and these settings:
- ACE server name `ACESERVER`
- listener for ACE web ui on port `7600`
- listener for ACE HTTP on port `7600`
run the following command:

`docker run --name aceserver -p 7600:7600 -p 7800:7800 -p 7843:7843 --env LICENSE=accept --env ACE_SERVER_NAME=ACESERVER ace-only:latest`

Once the console shows that the integration server is listening on port 7600, you can go to the ACE UI at http://localhost:7600/. To stop the container, run `docker stop aceserver` and the container will shut down cleanly, stopping the integration server.

### Sample image

In the `sample` folder there is an example on how to build a server image with a set of configuration and BAR files based on a previously built ACE only or ACE & MQ image. **[How to use the sample.](sample/README.md)**

### Environment variables supported by this image

- **LICENSE** - Set this to `accept` to agree to the App Connect Enterprise license. If you wish to see the license you can set this to `view`.
- **LANG** - Set this to the language you would like the license to be printed in.
- **LOG_FORMAT** - Set this to change the format of the logs which are printed on the container's stdout.  Set to "json" to use JSON format (JSON object per line); set to "basic" to use a simple human-readable format.  Defaults to "basic".
- **USE_QMGR** - Set to `true` to start a Queue Manager and set the Integration Server to use it.
- **ACE_ENABLE_METRICS** - Set this to `true` to generate Prometheus metrics for your Integration Server.
- **ACE_SERVER_NAME** - Set this to the name you want your Integration Server to run with.
- **ACE_TRUSTSTORE_PASSWORD** - Set this to the password you wish to use for the trust store (if using one).
- **ACE_KEYSTORE_PASSWORD** - Set this to the password you wish to use for the key store (if using one).

- **ACE_ADMIN_SERVER_SECURITY** - Set to `true` if you intend to secure your Integration Server using SSL.
- **ACE_ADMIN_SERVER_NAME** - Set this to the DNS name of your Integration Server for SSL SAN checking.
- **ACE_ADMIN_SERVER_CA** - Set this to your Integration Server SSL CA certificate.
- **ACE_ADMIN_SERVER_CERT** - Set this to your Integration Server SSL certificate.
- **ACE_ADMIN_SERVER_KEY** - Set this to your Integration Server SSL key certificate.

The following environment variables are used by MQ Advanced if being used:

- **LICENSE** - Set this to `accept` to agree to the App Connect Enterprise license. If you wish to see the license you can set this to `view`.
- **LANG** - Set this to the language you would like the license to be printed in.
- **MQ_QMGR_NAME** - Set this to the name you want your Queue Manager to be created with.
- **LOG_FORMAT** - Set this to change the format of the logs which are printed on the container's stdout.  Set to "json" to use JSON format (JSON object per line); set to "basic" to use a simple human-readable format.  Defaults to "basic".
- **MQ_ENABLE_METRICS** - Set this to `true` to generate Prometheus metrics for your Queue Manager.

When using MQ Advanced for Developers, there are extra environment variables available - check [MQ container configuration information](https://github.com/ibm-messaging/mq-container/blob/master/docs/developer-config.md#environment-variables) for more info.

## How to dynamically configure the ACE Integration Server

To enable dynamic configuration of the ACE Integration Server, this setup supports configuration injected into the image as files.

Before the Integration Server starts, the container is checked for the folder `/home/aceuser/initial-config`. For each folder in `/home/aceuser/initial-config` a script called `ace_config_{folder-name}.sh` will be run to process the information in the folder.
Shell scripts are supplied for the list of folders below, but you can extend this mechanism by adding your own folders and associated shell scripts.

- **Note**: The work dir for the Integration Server in the image is `/home/aceuser/ace-server`.
- **Note**: An example `initial-config` directory with data can be found in the `sample` folder, as well as the [command on how to mount it when running the image]((sample/README.md#run-the-sample-image).

You can mount the following file structure at `/home/aceuser/initial-config`. Missing folders will be skipped, but *empty* folders will cause an error:
- `/home/aceuser/initial-config/keystore`
   - A text file containing a certificate file in PEM format. This will be imported into the keystore file, along with the private key. The filename must be the *alias* for the certificate in the keystore, with the suffix `.crt`. The alias must not contain any whitespace characters.
   - A text file containing a private key file in PEM format. This will be imported into the keystore file, along with the certificate. The filename must be the *alias* for the certificate in the keystore, with the suffix `.key`.
   - If the private key is encrypted, then the passphrase may be specified in a file with the filename of *alias* with the suffix `.pass`.
   - The keystore file that will be created for these files needs a password. You must set the keystore password using the environment variable `ACE_KEYSTORE_PASSWORD`.
   - You can place multiple sets of files, each with a different file name/alias; each `.crt` file must have an associated `.key` file, and a `.pass` file must be present if the private key has a passphrase.
- `/home/aceuser/initial-config/odbcini`
   - A text file called `odbc.ini`. This must be an `odbc.ini` file suitable for the Integration Server to use when connecting to a database.  This will be copied to `/home/aceuser/ace-server/odbc.ini`.
- `/home/aceuser/initial-config/policy`
   - A set of `.policyxml` files, each with the suffix `.policyxml`, and a single `policy.descriptor` file.  These will be copied to `/home/aceuser/ace-server/overrides/DefaultPolicies/`. They should be specified in the `server.conf.yaml` section in order to be used.
- `/home/aceuser/initial-config/serverconf`
   - A text file called `server.conf.yaml` that contains a `server.conf.yaml` overrides file. This will be copied to `/home/aceuser/ace-server/overrides/server.conf.yaml`
- `/home/aceuser/initial-config/setdbparms`
   - For any parameters that need to be set via `mqsisetdbparms` include a text file called `setdbparms.txt` This supports 2 formats:
      ```
      # Lines starting with a "#" are ignored
      # Each line which starts mqsisetdbparms will be run as written 
      # Alternatively each line should specify the <resource> <userId> <password>, separated by a single space
      # Each line will be processed by calling...
      #   mqsisetdbparms ${ACE_SERVER_NAME} -n <resource> -u <userId> -p <password>
      resource1 user1 password1
      resource2 user2 password2
      mqsisetdbparms -w /home/aceuser/ace-server -n salesforce::SecurityIdentity -u myUsername -p myPassword -c myClientID -s myClientSecret
      ```
- `/home/aceuser/initial-config/truststore`
   - A text file containing a certificate file in PEM format. This will be imported into the truststore file as a trusted Certificate Authority's certificate. The filename must be the *alias* for the certificate in the keystore, with the suffix `.crt`. The alias must not contain any whitespace characters.
   - The truststore file that will be created for these files needs a password. You must set a truststore password using the environment variable `ACE_TRUSTSTORE_PASSWORD`
   - You can place multiple files, each with a different file name/alias.
- `/home/aceuser/initial-config/webusers`
   - A text file called `admin-users.txt`. It contains a list of users to be created as `admin` users using the command `mqsiwebuseradmin`. These users will have READ, WRITE and EXECUTE access on the Integration Server. The file has the following format:
     ```
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ", "WRITE" and "EXECUTE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r admin
     admin1 password1
     admin2 password2
     ```
     - A text file called `operator-users.txt`. It contains a list of users to be created as `operator` users using the command `mqsiwebuseradmin`. These users will have READ and EXECUTE access on the Integration Server. The file has the following format:
     ```
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" and "EXECUTE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r operator
     operator1 password1
     operator2 password2
     ```
     - A text file called `editor-users.txt`. It contains a list of users to be created as `editor` users using the command `mqsiwebuseradmin`. These users will have READ and WRITE access on the Integration Server. The file has the following format:
     ```
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" and "WRITE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r editor
     editor1 password1
     editor2 password2
     ```
     - A text file called `audit-users.txt`. It contains a list of users to be created as `audit` users using the command `mqsiwebuseradmin`. These users will have READ access on the Integration Server. The file has the following format:
     ```
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r audit
     audit1 password1
     audit2 password2
     ```
     - A text file called `viewer-users.txt`. It contains a list of users to be created as `viewer` users using the command `mqsiwebuseradmin`. These users will have READ access on the Integration Server. The file has the following format:
     ```
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r viewer
     viewer1 password1
     viewer2 password2
     ```
- `/home/aceuser/initial-config/mqsc`
   - A text file called `config.mqsc`. It contains a list of mqsc commands which will be processed on start by `runmqsc` command. Further details can be found in the [MQ Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.1.0/com.ibm.mq.adm.doc/q020670_.htm)
- `/home/aceuser/initial-config/agent`
   - A json file called 'switch.json' containing configuration information for the switch, this will be copied into the appropriate iibswitch directory
   - A json file called 'agentx.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
   - A json file called 'agentc.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
   - A json file called 'agentp.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
- `/home/aceuser/initial-config/extensions`
   - A zip file called `extensions.zip` will be extracted into the directory `/home/aceuser/ace-server/extensions`. This allows you to place extra files into a directory you can then reference in, for example, the server.conf.yaml
- `/home/aceuser/initial-config/ssl`
   - A pem file called 'ca.crt' will be extracted into the directory `/home/aceuser/ace-server/ssl`
   - A pem file called 'tls.key' will be extracted into the directory `/home/aceuser/ace-server/ssl`
   - A pem file called 'tls.cert' will be extracted into the directory `/home/aceuser/ace-server/ssl`

## Logging

The logs from the integration server running within the container, and MQ Queue Manager (if present), will log to standard out. The log entries can be output in two format:
- basic: human-headable for use in development when using `docker logs` or `kubectl logs`
- json: for pushing into ELK stack for searching and visualising in Kibana

The output format is controlled by the `LOG_FORMAT` environment variable, which also controls the output format of the MQ Queue Manager logs (if present).

A sample Kibana dashboard is available at sample/dashboards/ibm-ace-kibana-dashboard.json

## Monitoring

The accounting and statistics feature in IBM App Connect Enterprise provides the component level data with detailed insight into the running message flows to enabled problem determination, profiling, capacity planning, situation alert monitoring and charge-back modelling.

A Prometheus exporter runs on port 9483 if `ACE_ENABLE_METRICS` is set to `true` - the exporter listens for accounting and statistics, and resource statistics, data on a websocket from the integration server, then aggregates this data to make available to Prometheus when requested.

A sample Grafana dashboard is available at sample/dashboards/ibm-ace-grafana-dashboard.json

If `MQ_ENABLE_METRICS` is set to `true` and an MQ Queue Manager is present an additional exporter is run on port 9157 to provide MQ statistics to Prometheus: https://github.com/ibm-messaging/mq-container/blob/master/docs/internals.md#prometheus-metrics

# License

The Dockerfile and associated scripts are licensed under the [Eclipse Public License 2.0](LICENSE). Licenses for the products installed within the images are as follows:

 - IBM App Connect Enterprise for Developers is licensed under the IBM International License Agreement for Non-Warranted Programs. This license may be viewed from the image using the `LICENSE=view` environment variable as described above.

Note that the IBM App Connect Enterprise for Developers license does not permit further distribution.

# Copyright

© Copyright IBM Corporation 2015, 2018
