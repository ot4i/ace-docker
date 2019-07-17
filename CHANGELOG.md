# Change log

## 11.0.0.5 (2019-07-x)

**Breaking changes**:
* When using MQ, the UID of the mqm user is now 888.  You need to run the container with an entrypoint of `runmqserver -i` under the root user to update any existing files.
* MQSC files supplied will be verified before being run. Files containing invalid MQSC will cause the container to fail to start

**Other changes**:
* Security fixes
* Web console added to production image
* Container built on RedHat host (UBI)
* Includes MQ 9.1.2
* Supports configuring agent files
* Supports installing additional config files using extensions.zip

## 11.0.0.3 (2019-02-04)

**Breaking changes**:
NONE

**Other changes**:
* Provides samples for building image with MQ Client
* Code to generate RHEL based images
* Fix for overriding the hostname and port for RestAPI in the UI / Swagger Docs.

## 11.0.0.2 (2018-11-20)

**Breaking changes**:
NONE

**Other changes**:
* Updated to support 11.0.0.2 runtime
* Updated to support ICP platform

## 11.0.0.0 (2019-10-08)

* Initial version
