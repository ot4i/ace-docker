# ace-alpine
ACE in an Alpine container

export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NON_IBM_JAVA=1
export MQSI_NO_CACHE_SUPPORT=1
--admin-rest-api -1

export LD_LIBRARY_PATH=/lib:/usr/lib/jvm/default-jvm/jre/lib/amd64/server:/usr/lib/jvm/default-jvm/jre/lib/amd64:$LD_LIBRARY_PATH


sed -i 's/LICENSE_ACCEPTED=0/LICENSE_ACCEPTED=1/g' mqsiprofile
