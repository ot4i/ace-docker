ECHO ON

dir c:\tmp


powershell -Command "Expand-Archive -Path c:\tmp\13.0.4.0-ACE-WIN64-EVALUATION.zip -DestinationPath c:\tmp\ace-unzip"
dir c:\tmp\ace-unzip