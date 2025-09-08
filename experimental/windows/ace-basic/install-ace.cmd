ECHO ON

dir c:\tmp
dir c:\tmp\ace-unzip

cd c:\tmp\ace-unzip\
REM .\ACESetup13.0.4.0.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\ace-v13 InstallToolkit=0 InstallGAC=0
.\ACESetup13.0.4.0.exe -silent -installFolder C:\ace-13 -licenseAccept yes -anonymousUsageStatistics no -installToolkit no -installWSRRnodes no -installElectronApp no 

REM Install the .Net support DLLs
call c:\ace-v13\server\bin\runCommand.cmd C:\ace-v13\server\bin\mqsiAssemblyInstall -i C:\ace-v13\server\bin\IBM.Broker.Plugin.dll
call c:\ace-v13\server\bin\runCommand.cmd C:\ace-v13\server\bin\mqsiAssemblyInstall -i C:\ace-v13\server\bin\IBM.Broker.Support.dll
call c:\ace-v13\server\bin\runCommand.cmd C:\ace-v13\server\bin\mqsiAssemblyInstall -i C:\ace-v13\server\bin\microsoft.crm.sdk.proxy.dll
call c:\ace-v13\server\bin\runCommand.cmd C:\ace-v13\server\bin\mqsiAssemblyInstall -i C:\ace-v13\server\bin\microsoft.xrm.sdk.dll
call c:\ace-v13\server\bin\runCommand.cmd C:\ace-v13\server\bin\mqsiAssemblyInstall -i C:\ace-v13\server\bin\Microsoft.IdentityModel.dll
