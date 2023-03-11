ECHO ON

echo %1
cd c:\tmp
ipconfig /all

REM Not clear why we need this sometimes, but DNS resolution doesn't always work out-of-the-box
netsh interface ip set dns name="Ethernet" static 8.8.8.8

curl --location -o aria2.zip https://github.com/aria2/aria2/releases/download/release-1.36.0/aria2-1.36.0-win-64bit-build1.zip
dir c:\tmp\aria2.zip
powershell -Command "Expand-Archive -Path c:\tmp\aria2.zip -DestinationPath c:\tmp\aria-unzip"
c:\tmp\aria-unzip\aria2-1.36.0-win-64bit-build1\aria2c.exe -s 10 -j 10 -x 10 %1

dir c:\tmp

powershell -Command "Expand-Archive -Path c:\tmp\12.0.7.0-ACE-WIN64-DEVELOPER.zip -DestinationPath c:\tmp\ace-unzip"
dir c:\tmp\ace-unzip

REM Default would normally be
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'LICENSE_ACCEPTED' to value 'false'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallToolkit' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallElectronApp' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallCloudConnectors' to value '1'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallWSRRnodes' to value '0'
REM [0660:0760][2023-03-10T17:21:57]i000: Initializing string variable 'InstallGAC' to value '1'

cd c:\tmp\ace-unzip\
.\ACESetup12.0.7.0.exe /quiet LICENSE_ACCEPTED=true InstallFolder=C:\ace-v12 InstallToolkit=0 InstallGAC=0
REM Install the .Net support DLLs
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\IBM.Broker.Plugin.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\IBM.Broker.Support.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\microsoft.crm.sdk.proxy.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\microsoft.xrm.sdk.dll
call c:\ace-v12\server\bin\runCommand.cmd C:\ace-v12\server\bin\mqsiAssemblyInstall -i C:\ace-v12\server\bin\Microsoft.IdentityModel.dll

cd c:\tmp
rmdir /s /q c:\tmp\ace-unzip
del /q c:\tmp\12*.zip
