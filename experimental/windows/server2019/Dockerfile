FROM mcr.microsoft.com/windows/servercore:1809

WORKDIR c:\\tmp
COPY ACESetup11.0.16742.4.exe c:\\tmp\\ACESetup.exe
# Install the product, but do not install the .Net support DLLs; installer issues prevent this from working
# in a Docker container
RUN c:\\tmp\\ACESetup.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\\ace-v11 InstallToolkit=0 InstallGAC=0
# Install the .Net support DLLs
RUN c:\\ace-v11\\server\\\bin\\\runCommand.cmd C:\\ace-v11\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v11\\server\\bin\\IBM.Broker.Plugin.dll
RUN c:\\ace-v11\\server\\\bin\\\runCommand.cmd C:\\ace-v11\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v11\\server\\bin\\IBM.Broker.Support.dll
RUN c:\\ace-v11\\server\\\bin\\\runCommand.cmd C:\\ace-v11\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v11\\server\\bin\\microsoft.crm.sdk.proxy.dll
RUN c:\\ace-v11\\server\\\bin\\\runCommand.cmd C:\\ace-v11\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v11\\server\\bin\\microsoft.xrm.sdk.dll
RUN c:\\ace-v11\\server\\\bin\\\runCommand.cmd C:\\ace-v11\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v11\\server\\bin\\Microsoft.IdentityModel.dll
