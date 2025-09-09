# Windows containers

Simple docker images for ACE v13 on Windows, using Windows containers. These containers
will run Windows Server Core, and are not running Linux containers on Windows.

Dockefiles in the following directories are used for various purposes:

- ace-basic can run the product server, including all files except the toolkit.
- ace-sample contains a sample BAR file and Dockerfiles for building runnable images to serve HTTP clients.

## Docker Engine versus Docker Desktop

Docker Desktop for Windows may require licensing, and Docker Engine can be used instead.
Follow the instructions at 
https://docs.docker.com/engine/install/binaries/#install-server-and-client-binaries-on-windows
to install and start the Docker service.

## Notes on Windows base images

Not all Windows container base images can be used with all Windows versions, and the
sizes are different between the different images:
```
C:\>docker images
REPOSITORY                             TAG        IMAGE ID       CREATED       SIZE
mcr.microsoft.com/windows/servercore   ltsc2025   53e5675cc5d0   3 weeks ago   7.69GB
mcr.microsoft.com/windows/servercore   ltsc2019   7cad9b9d6557   4 weeks ago   5GB
mcr.microsoft.com/windows/servercore   ltsc2022   ddc7ece39744   4 weeks ago   5.2GB
```
At the time of writing, Windows 11 (10.0.26100) using Docker Engine 28.4.0 will run 
ltsc2022 and ltsc2025 but not ltsc2019. This combination takes a long time to build
if Windows Defender scans every file, so running 
```
Add-MpPreference -ExclusionPath C:\ProgramData\docker\windowsfilter
```
from an Administrator PowerShell should improve build times. 

The `dockerd` process will still use `C:\Windows\SystemTemp` for temporary file storage
unless the `SystemTemp` environment variable is set (which is hard to do for a normal
service without affecting the whole machine but can be done by running `dockerd -D` from
an Adminstrator shell), but moving this to an excluded directory does not seem to change
the build times by very much; it seems the layer copy triggers Defender scanning even
when copying to and from an excluded location (possibly due to NTFS file log updates).
