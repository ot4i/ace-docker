ECHO ON

echo %1
cd c:\tmp
ipconfig /all

netsh interface ip set dns name="Ethernet" static 8.8.8.8
curl --location -o aria2.zip https://github.com/aria2/aria2/releases/download/release-1.36.0/aria2-1.36.0-win-64bit-build1.zip
dir c:\tmp\aria2.zip
powershell -Command "Expand-Archive -Path c:\tmp\aria2.zip -DestinationPath c:\tmp\aria-unzip"
c:\tmp\aria-unzip\aria2-1.36.0-win-64bit-build1\aria2c.exe -s 10 -j 10 -x 10 %1
