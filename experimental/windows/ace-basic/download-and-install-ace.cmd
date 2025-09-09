ECHO ON

cd c:\tmp

REM ****************************************************************
REM 
REM Network setup - may not always be needed!
REM 
REM **************************************************************** 
ipconfig /all
REM Not clear why we need this sometimes, but DNS resolution doesn't always work out-of-the-box
netsh interface ip set dns name="Ethernet" static 8.8.8.8

REM Use this one with --isolation process (sometimes has no effect)
REM netsh interface ip set dns name="vEthernet" static 8.8.8.8

REM **************************************************************** 
REM 
REM Download the aria2 utility to speed up the ACE download by using
REM multiple connections to the server.
REM 
REM **************************************************************** 
REM Windows curl doesn't support --dns-servers 8.8.8.8
curl --location -o aria2.zip https://github.com/aria2/aria2/releases/download/release-1.36.0/aria2-1.36.0-win-64bit-build1.zip
dir c:\tmp\aria2.zip
powershell -Command "Expand-Archive -Path c:\tmp\aria2.zip -DestinationPath c:\tmp\aria-unzip"


REM **************************************************************** 
REM 
REM Download the ACE image and unzip it.
REM 
REM **************************************************************** 
c:\tmp\aria-unzip\aria2-1.36.0-win-64bit-build1\aria2c.exe -s 10 -j 10 -x 10 %1
dir c:\tmp
powershell -Command "Expand-Archive -Path c:\tmp\13.0.4.0-ACE-WIN64-EVALUATION.zip -DestinationPath c:\tmp\ace-unzip"
dir c:\tmp\ace-unzip

REM **************************************************************** 
REM 
REM Install ACE itself with the correct options. Install logs will 
REM be placed in the %TEMP% directory if anything goes wrong.
REM 
REM Note that the .Net DLLs don't seem to install correctly from the 
REM installer itself, but can be installed explicitly afterwards.
REM 
REM Default would normally be
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'LICENSE_ACCEPTED' to value 'false'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallToolkit' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallElectronApp' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallCloudConnectors' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallWSRRnodes' to value '0'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallGAC' to value '1'
REM 
REM **************************************************************** 
cd c:\tmp\ace-unzip\

REM Note change from v12 to v13
REM .\ACESetup12.0.10.0.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\ace-12 InstallToolkit=0 InstallGAC=0 InstallElectronApp=0
.\ACESetup13.0.4.0.exe -silent -installFolder C:\ace-13 -licenseAccept yes -anonymousUsageStatistics no -installToolkit no -installWSRRnodes no -installElectronApp no 

REM Install the .Net support DLLs
call c:\ace-13\server\bin\runCommand.cmd C:\ace-13\server\bin\mqsiAssemblyInstall -i C:\ace-13\server\bin\IBM.Broker.Plugin.dll
call c:\ace-13\server\bin\runCommand.cmd C:\ace-13\server\bin\mqsiAssemblyInstall -i C:\ace-13\server\bin\IBM.Broker.Support.dll
call c:\ace-13\server\bin\runCommand.cmd C:\ace-13\server\bin\mqsiAssemblyInstall -i C:\ace-13\server\bin\microsoft.crm.sdk.proxy.dll
call c:\ace-13\server\bin\runCommand.cmd C:\ace-13\server\bin\mqsiAssemblyInstall -i C:\ace-13\server\bin\microsoft.xrm.sdk.dll
call c:\ace-13\server\bin\runCommand.cmd C:\ace-13\server\bin\mqsiAssemblyInstall -i C:\ace-13\server\bin\Microsoft.IdentityModel.dll

REM **************************************************************** 
REM 
REM Clean up to make the resulting image size smaller.
REM 
REM **************************************************************** 
cd c:\tmp
rmdir /s /q c:\tmp\ace-unzip
del /q c:\tmp\13*.zip
