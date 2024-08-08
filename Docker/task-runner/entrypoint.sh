#!/bin/bash

tempTasksPath="/var/temp_tasks"
tempFolderPath="$tempTasksPath/$SOURCE_CODE_FOLDER"


stdinFileName=$tempFolderPath/stdin.txt
stdoutFileName=$tempFolderPath/stdout.txt
artefactsFileName=$tempFolderPath/artefacts.txt
errorFileName=$tempFolderPath/error.txt

touch $stdoutFileName
touch $artefactsFileName
touch $errorFileName

cd /home/tester
g++ $tempFolderPath/$SOURCE_FILE_NAME -o ./code.out

if [ $? -eq 0 ]; then
  runuser tester -c ./code.out < $stdinFileName > $stdoutFileName
  find . -maxdepth 1 -type f -name 'output*' -exec echo {} > $artefactsFileName \; -exec mv {} "$tempFolderPath" \;
else
  echo "Compilation failed" > $errorFileName
fi
