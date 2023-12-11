ECHO ON

dir c:\tmp


powershell -Command "Expand-Archive -Path c:\tmp\12.0.10.0-ACE-WIN64-DEVELOPER.zip -DestinationPath c:\tmp\ace-unzip"
dir c:\tmp\ace-unzip