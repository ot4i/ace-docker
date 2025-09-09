# Combined file for easier scripting
. /opt/ibm/ace-13/server/bin/mqsiprofile

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-13/common/jdk/bin:$PATH

# Not really java-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
