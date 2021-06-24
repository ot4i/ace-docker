#!/bin/bash

#Check if any iFix's are to be applied.

if [ -z $1 ]; then
  echo "No iFix's being used skipping this phase."
  exit 0 #No iFix specified
else #Create array of iFix's to loop through by removing commas from input.
  IFIX_LIST_ARRAY=$(echo $1 | tr ',' ' ')
  echo "iFix's being applied: ${IFIX_LIST_ARRAY}"
fi

for ifixlink in $IFIX_LIST_ARRAY  
do
  #Make temporary fix directory
  mkdir ./fix 
  cd fix 
  
  #Download and unzip iFix tar file.
  curl -Ls $ifixlink | tar -xz
  
  #Execute install command
  ifixname="${ifixlink##*/}"
  ifixname="${ifixname%.tar*}" 
  
  ./mqsifixinst.sh /opt/ibm/ace-12 install $ifixname
  
  #Delete directory
  cd ..
  rm -rf ./fix
  
  rm -rf /opt/ibm/ace-12/fix-backups.12.0.1.0
  rm /opt/ibm/ace-12/mqsifixinst.log
  rm /opt/ibm/ace-12/mqsifixinst.sh
  
done
