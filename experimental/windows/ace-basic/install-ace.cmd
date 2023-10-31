ECHO ON

dir c:\tmp
dir c:\tmp\ace-unzip

cd c:\tmp\ace-unzip\
.\ACESetup12.0.10.0.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\ace-v12 InstallToolkit=0 InstallGAC=0
REM Install the .Net support DLLs
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\IBM.Broker.Plugin.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\IBM.Broker.Support.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\microsoft.crm.sdk.proxy.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\microsoft.xrm.sdk.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\Microsoft.IdentityModel.dll
