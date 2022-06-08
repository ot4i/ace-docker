FROM mcr.microsoft.com/windows/servercore/insider:10.0.20348.1

WORKDIR c:\\tmp
COPY ACESetup12.0.1.0.exe c:\\tmp\\ACESetup.exe
# Install the product, but do not install the .Net support DLLs; installer issues prevent this from working
# in a Docker container
RUN c:\\tmp\\ACESetup.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\\ace-v12 InstallToolkit=0 InstallGAC=0
# Install the .Net support DLLs
RUN c:\\ace-v12\\server\\\bin\\\runCommand.cmd C:\\ace-v12\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v12\\server\\bin\\IBM.Broker.Plugin.dll
RUN c:\\ace-v12\\server\\\bin\\\runCommand.cmd C:\\ace-v12\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v12\\server\\bin\\IBM.Broker.Support.dll
RUN c:\\ace-v12\\server\\\bin\\\runCommand.cmd C:\\ace-v12\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v12\\server\\bin\\microsoft.crm.sdk.proxy.dll
RUN c:\\ace-v12\\server\\\bin\\\runCommand.cmd C:\\ace-v12\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v12\\server\\bin\\microsoft.xrm.sdk.dll
RUN c:\\ace-v12\\server\\\bin\\\runCommand.cmd C:\\ace-v12\\server\\bin\\mqsiAssemblyInstall -i C:\\ace-v12\\server\\bin\\Microsoft.IdentityModel.dll

# Could also be set to ContainerUser
USER ContainerAdministrator
