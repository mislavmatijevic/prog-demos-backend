#!/bin/bash

tempTasksPath="/var/temp_tasks"

cd /home/tester

g++ $tempTasksPath/$SOURCE_CODE_FOLDER/$SOURCE_FILE_NAME -o ./code.out

inputFileName=$tempTasksPath/$SOURCE_CODE_FOLDER/input.txt
outputFileName=$tempTasksPath/$SOURCE_CODE_FOLDER/output.txt
errorFileName=$tempTasksPath/$SOURCE_CODE_FOLDER/error.txt

touch $outputFileName
touch $errorFileName

if [ $? -eq 0 ]; then
  runuser tester -c ./code.out < $inputFileName > $outputFileName
else
  echo "Compilation failed" > $errorFileName
fi
