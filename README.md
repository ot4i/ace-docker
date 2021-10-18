# Overview

<img src="./app_connect_light_256x256.png" width="100" alt="IBM ACE logo"/>

Run [IBM® App Connect Enterprise](https://developer.ibm.com/integration/docs/app-connect-enterprise/faq/) in a container.

You can build an image containing one of the following combinations:
- IBM App Connect Enterprise 
- IBM App Connect Enterprise with IBM MQ Client
- IBM App Connect Enterprise for Developers
- IBM App Connect Enterprise for Developers with IBM MQ Client

For information of building images with IBM MQ Advanced please refer to [IBM App Connect Enterprise for Developers with IBM MQ Advanced for Developers](README-MQ.md)

The IBM App Connect operator now supports a single image which includes both the ACE server runtime as well as an MQ client. This readme will describe how you can build an equivalent image.

Pre-built developer and production edition image can be found on IBM Container Registry - [Obtaining the IBM App Connect Enterprise server image from the IBM Cloud Container Registry](https://www.ibm.com/support/knowledgecenter/en/SSTTDS_11.0.0/com.ibm.ace.icp.doc/certc_install_obtaininstallationimageser.html)


## Building a container image

Download a copy of App Connect Enterprise (ie. `ace-12.0.1.0.tar.gz`) and place it in the `deps` folder. When building the image use `build-arg` to specify the name of the file: `--build-arg ACE_INSTALL=ace-12.0.1.0.tar.gz`

**Important:** Only ACE version **12.0.1.0 or greater** is supported.

Choose if you want to have an image with just App Connect Enterprise or an image with both App Connect Enterprise and the IBM MQ Client libraries. The second of these is used by the IBM App Connect operator.

### Building a container image which contains an IBM Service provided fix for ACE

You may have been provided with a fix for App Connect Enterprise by IBM Support, this fix will have a name of the form `12.0.X.Y-ACE-LinuxX64-TF12345.tar.gz`. This fix can be used to create a container image in one of two different ways:

#### Installation during container image build
This method builds a new container image derived from an existing ACE container image and applies the ifix using the standard `mqsifixinst.sh` script. The ifix image can be built from any existing ACE container image, e.g. `ace-only`, `ace-mqclient`, or another ifix image. Simply build `Dockerfile.ifix` passing in the full `BASE_IMAGE` name and the `IFIX_ID` arguments set:

```bash
docker build -t ace-server:12.0.x.y-r1-tfit12345 --build-arg BASE_IMAGE=ace-server:12.0.x.y-1 --build-arg IFIX_ID=12.0.X.Y-ACE-LinuxX64-TFIT12345 --file ubi/Dockerfile.ifix path/to/folder/containing/ifix
```

#### Pre-applying the fix to the ACE install image
This method applies the ifix directly to the ACE installation image that is consumed to make the full container image. **NB**: Only follow these instructions if you have been instructed by IBM Support to "manually install" the ifix, or that the above method is not applicable to your issue. If you follow these instructions then the ifix ID will _not_ appear in the output of `mqsiservice -v`.

In order to apply this fix manually follow these steps.
 - On a local system extract the App Connect Enterprise archive
   `tar -xvf ace-12.0.1.0.tar.gz`
 - Extract the fix package into expanded App Connect Enterprise installation
   `tar -xvf /path/to/12.0.1.0-ACE-LinuxX64-TF12345.tar.gz --directory ace-12.0.1.0`
 - Tar and compress the resulting App Connect Enterprise installation
   `tar -cvf ace-12.0.1.0_with_IT12345.tar ace-12.0.1.0`
   `gzip ace-12.0.1.0_with_IT12345.tar`
 - Place the resulting `ace-12.0.1.0_with_IT12345.tar.gz` file in the `deps` folder and when building using the `build-arg` to specify the name of the file: `--build-arg ACE_INSTALL=ace-12.0.1.0_with_IT12345.tar.gz`

### Using App Connect Enterprise for Developers

Get [ACE for Developers edition](https://www.ibm.com/marketing/iwm/iwm/web/pick.do?source=swg-wmbfd). Then place it in the `deps` folder as mentioned above.

### Build an image with App Connect Enterprise only

NOTE: The current dockerfiles are tailored towards use by the App Connect Operator and as a result may have function removed from it if we are no longer using it in our operator. If you prefer to use the old dockerfiles for building your containers please use the `Dockerfile-legacy.aceonly` file


The `deps` folder must contain a copy of ACE, **version 12.0.1.0 or greater**. If using ACE for Developers, download it from [here](https://www.ibm.com/marketing/iwm/iwm/web/pick.do?source=swg-wmbfd).
Then set the build argument `ACE_INSTALL` to the name of the ACE file placed in `deps`.

1. ACE for Developers only:
   - `docker build -t ace-dev-only --build-arg ACE_INSTALL={ACE-dev-file-in-deps-folder} --file ubi/Dockerfile.aceonly .`
2. ACE production only: 
   - `docker build -t ace-only --build-arg ACE_INSTALL={ACE-file-in-deps-folder} --file ubi/Dockerfile.aceonly .`

### Build an image with App Connect Enterprise and MQ Client

Follow the instructions above for building an image with App Connect Enterprise Only.

Add the MQ Client libraries to your existing image by running `docker build -t ace-mqclient --build-arg BASE_IMAGE=<AceOnlyImageTag> --file ubi/Dockerfile.mqclient .`

`<AceOnlyImageTag>` is the tag of the image you want to add the client libs to i.e. ace-only. You can supply a customer URL for the MQ binaries by setting the argument MQ_URL

## Usage

### Accepting the License

In order to use the image, it is necessary to accept the terms of the IBM App Connect Enterprise license. This is achieved by specifying the environment variable `LICENSE` equal to `accept` when running the image. You can also view the license terms by setting this variable to `view`. Failure to set the variable will result in the termination of the container with a usage statement. You can view the license in a different language by also setting the `LANG` environment variable.

### Red Hat OpenShift SecurityContextConstraints Requirements

The predefined SecurityContextConstraint (SCC) `restricted` has been verified with the image when being run in a Red Hat OpenShift environment.

### Running the container

To run a container with ACE only with default configuration and these settings:

- ACE server name `ACESERVER`
- listener for ACE web ui on port `7600`
- listener for ACE HTTP on port `7600`
run the following command:

`docker run --name aceserver -p 7600:7600 -p 7800:7800 -p 7843:7843 --env LICENSE=accept --env ACE_SERVER_NAME=ACESERVER ace-only:latest`

Once the console shows that the integration server is listening on port 7600, you can go to the ACE UI at http://localhost:7600/. To stop the container, run `docker stop aceserver` and the container will shut down cleanly, stopping the integration server.

### Sample image

In the `sample` folder there is an example on how to build a server image with a set of configuration and BAR files based on a previously built ACE image. **[How to use the sample.](sample/README.md)**

### Environment variables supported by this image

- **LICENSE** - Set this to `accept` to agree to the App Connect Enterprise license. If you wish to see the license you can set this to `view`.
- **LANG** - Set this to the language you would like the license to be printed in.
- **LOG_FORMAT** - Set this to change the format of the logs which are printed on the container's stdout.  Set to "json" to use JSON format (JSON object per line); set to "basic" to use a simple human-readable format.  Defaults to "basic".
- **ACE_ENABLE_METRICS** - Set this to `true` to generate Prometheus metrics for your Integration Server.
- **ACE_SERVER_NAME** - Set this to the name you want your Integration Server to run with.
- **ACE_TRUSTSTORE_PASSWORD** - Set this to the password you wish to use for the trust store (if using one).
- **ACE_KEYSTORE_PASSWORD** - Set this to the password you wish to use for the key store (if using one).

- **ACE_ADMIN_SERVER_SECURITY** - Set to `true` if you intend to secure your Integration Server using SSL.
- **ACE_ADMIN_SERVER_NAME** - Set this to the DNS name of your Integration Server for SSL SAN checking.
- **ACE_ADMIN_SERVER_CA** - Set this to your Integration Server SSL CA certificates folder.
- **ACE_ADMIN_SERVER_CERT** - Set this to your Integration Server SSL certificate.
- **ACE_ADMIN_SERVER_KEY** - Set this to your Integration Server SSL key certificate.

- **FORCE_FLOW_HTTPS** - Set to 'true' and the *.key and *.crt present in `/home/aceuser/httpsNodeCerts/` are used to force all your flows to use https 

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

      ```script
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

     ```script
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ", "WRITE" and "EXECUTE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r admin
     admin1 password1
     admin2 password2
     ```

  - A text file called `operator-users.txt`. It contains a list of users to be created as `operator` users using the command `mqsiwebuseradmin`. These users will have READ and EXECUTE access on the Integration Server. The file has the following format:

     ```script
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" and "EXECUTE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r operator
     operator1 password1
     operator2 password2
     ```

  - A text file called `editor-users.txt`. It contains a list of users to be created as `editor` users using the command `mqsiwebuseradmin`. These users will have READ and WRITE access on the Integration Server. The file has the following format:

     ```script
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" and "WRITE" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r editor
     editor1 password1
     editor2 password2
     ```

  - A text file called `audit-users.txt`. It contains a list of users to be created as `audit` users using the command `mqsiwebuseradmin`. These users will have READ access on the Integration Server. The file has the following format:

     ```script
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r audit
     audit1 password1
     audit2 password2
     ```

  - A text file called `viewer-users.txt`. It contains a list of users to be created as `viewer` users using the command `mqsiwebuseradmin`. These users will have READ access on the Integration Server. The file has the following format:

     ```script
     # Lines starting with a "#" are ignored
     # Each line should specify the <user> <password>, separated by a single space
     # Each user will have "READ" access on the integration server
     # Each line will be processed by calling...
     #   mqsiwebuseradmin -w /home/aceuser/ace-server -c -u <user> -a <password> -r viewer
     viewer1 password1
     viewer2 password2
     ```

- `/home/aceuser/initial-config/agent`
  - A json file called 'switch.json' containing configuration information for the switch, this will be copied into the appropriate iibswitch directory
  - A json file called 'agentx.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
  - A json file called 'agenta.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
  - A json file called 'agentc.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
  - A json file called 'agentp.json' containing configuration information for the agent connectivity, this will be copied into the appropriate iibswitch directory
- `/home/aceuser/initial-config/extensions`
  - A zip file called `extensions.zip` will be extracted into the directory `/home/aceuser/ace-server/extensions`. This allows you to place extra files into a directory you can then reference in, for example, the server.conf.yaml
- `/home/aceuser/initial-config/ssl`
  - A pem file called 'ca.crt' will be extracted into the directory `/home/aceuser/ace-server/ssl`
  - A pem file called 'tls.key' will be extracted into the directory `/home/aceuser/ace-server/ssl`
  - A pem file called 'tls.cert' will be extracted into the directory `/home/aceuser/ace-server/ssl`
- `/home/aceuser/initial-config/bar_overrides`
  - For any parameters that need to be set via `mqsiapplybaroverride` include text files with extension `.properties` Eg:
   ```script
   sampleFlow#MQInput.queueName=NEWC
   ```
- `/home/aceuser/initial-config/workdir_overrides`
  - For any parameters that need to be set via `ibm int apply overrides --work-directory /home/aceuser/ace-server` on the integration server include text files Eg:
   ```script
   TestFlow#HTTP Input.URLSpecifier=/production
   ```


## Logging

The logs from the integration server running within the container will log to standard out. The log entries can be output in two format:

- basic: human-headable for use in development when using `docker logs` or `kubectl logs`
- json: for pushing into ELK stack for searching and visualising in Kibana

The output format is controlled by the `LOG_FORMAT` environment variable

A sample Kibana dashboard is available at sample/dashboards/ibm-ace-kibana-dashboard.json

## Monitoring

The accounting and statistics feature in IBM App Connect Enterprise provides the component level data with detailed insight into the running message flows to enabled problem determination, profiling, capacity planning, situation alert monitoring and charge-back modelling.

A Prometheus exporter runs on port 9483 if `ACE_ENABLE_METRICS` is set to `true` - the exporter listens for accounting and statistics, and resource statistics, data on a websocket from the integration server, then aggregates this data to make available to Prometheus when requested.

A sample Grafana dashboard is available at sample/dashboards/ibm-ace-grafana-dashboard.json

## License

The Dockerfile and associated scripts are licensed under the [Eclipse Public License 2.0](LICENSE). Licenses for the products installed within the images are as follows:

- IBM App Connect Enterprise for Developers is licensed under the IBM International License Agreement for Non-Warranted Programs. This license may be viewed from the image using the `LICENSE=view` environment variable as described above.

Note that the IBM App Connect Enterprise for Developers license does not permit further distribution.

## Copyright

© Copyright IBM Corporation 2015, 2018
