# ace-full

This image exists to enable the use of mqsicreatebar in build pipelines with ACE v12.
 
To build the image, run
```
docker build -t ace-full:12.0.2.0-ubuntu -f Dockerfile.ubuntu .
```
in this directory.
 
Containers must be started with the LICENSE environment variable set to "accept" for the 
product to be usable (for example, "-e LICENSE=accept" on a docker run command).

## Build image with mqsicreatebar

This container can be used as a base image for other build containers, or could be run with a 
volume mount containing the artifacts and scripts to be built. For mqsicreatebar to run without
errors, an X-Windows server must be available even though no GUI activity is needed. This image
contains a virtual X server that fulfils this need, and it must be started before running 
mqsicreatebar:
 
```
Xvfb -ac :99 &
export DISPLAY=:99
```

where the first command starts the server, and the second sets the DISPLAY variable to point 
to the virtual server. The virtual X server must be started for each container, but many 
mqsicreatebar commands can be run after the server has been started once. 
 
## Toolkit image with X-Windows

This image can also be used to run the ACE toolkit without needing it be installed on the
host system as long as the host has an X-Windows display running. In this case, no virtual
X server is needed, and the main issue is getting the authorization credentials into the
container. To do this, mount the .Xauthority file from the host into the container as follows:

```
chmod 664 ~/.Xauthority
docker run -e LICENSE=accept -e DISPLAY --network=host -v $HOME/.Xauthority:/home/aceuser/.Xauthority --rm -ti ace-full:12.0.2.0-ubuntu
```
with the "--network=host" and "-e DISPLAY" settings allowing the container the network access 
it needs to get to the host X server. If there is no .Xauthority file in the home directory, it
can be copied from /var/run/gdm on most systems, and in many cases the XAUTHORITY environment 
variable points to the correct location, so
```
cp $XAUTHORITY ~/.Xauthority
chmod 664 ~/.Xauthority
```
creates the correct file. The chmod is needed due to the docker container running as a different
user from the host user, and if the .Xauthority file is not readable then the toolkit will not be
able to connect to the host display.

Assuming permissions are set correctly, then the docker run command above should lead to the 
standard ACE profile banner, and at that point it should be possible to start the toolkit:
```
MQSI 12.0.2.0
/opt/ibm/ace-12/server

(ACE_12:)aceuser@tdolby-laptop:/$ /opt/ibm/ace-12/ace tools
Starting App Connect Enterprise Toolkit interactively
(ACE_12:)aceuser@tdolby-laptop:/$
```

The toolkit may take some time to start, but should succeed in bringing up a splash screen 
and then a prompt for a workspace location. 

### Errors from Xvfb

Attempting to run the Xvfb command from the mqsicreatebar section above while also using the
X-Windows forwarding described in this section will lead to errors:
```
(ACE_12:)aceuser@tdolby-laptop:/$ Xvfb
_XSERVTransSocketUNIXCreateListener: ...SocketCreateListener() failed
_XSERVTransMakeAllCOTSServerListeners: server already running
(EE)
Fatal server error:
(EE) Cannot establish any listening sockets - Make sure an X server isn't already running(EE) 
```
In this case, the solution is to not attempt to run Xvfb when an X-Windows display has
already been set: Xvfb is only present for "headless" running via scripts, and is not needed
if the toolkit is used.
