#!/bin/bash

homeDirectory="/home/tester"

sourceFileName="$SOURCE_FILE_NAME"
executionDirectory="$SOURCE_CODE_FOLDER"
solutionFile="$executionDirectory/$sourceFileName"
errorFile="$executionDirectory/error.txt"
outputFile="$executionDirectory/solution.out"
stdinFilenamePrefix="$STDIN_FILENAME_PREFIX"
stdoutFilenamePrefix="$STDOUT_FILENAME_PREFIX"
artefactsFilenamePrefix="$ARTEFACTS_FILENAME_PREFIX"

rm .bash_logout .bashrc .profile

g++ "$solutionFile" -o "$outputFile" 2> "$errorFile"

if [ $? -ne 0 ]; then
    cat "$errorFile"
    exit 1
fi

chmod 701 "$executionDirectory"
chmod 701 "$outputFile"

find "$executionDirectory" -type f -name "${stdinFilenamePrefix}*" | while read -r stdinFile; do
    taskId=$(basename "$stdinFile" | sed -E 's/[^0-9]*([0-9]+).*/\1/')

    stdoutFile="$executionDirectory/${stdoutFilenamePrefix}${taskId}.txt"
    touch "$stdoutFile"
    chmod 600 "$stdoutFile"

    (cd $homeDirectory && exec runuser -u tester -- "$outputFile" < "$stdinFile" > "$stdoutFile" 2>> "$errorFile")

    if [ -s "$errorFile" ]; then
        cat "$errorFile"
        exit 1
    else
        rm -f "$errorFile"
    fi

    files=($(find "$homeDirectory" -type f | sort))

    if [[ ${#files[@]} -gt 0 ]]; then
        tempConcatFile="$executionDirectory/artefacts_concat.txt"
        artefactsFile="$executionDirectory/${artefactsFilenamePrefix}${taskId}.txt"
        touch "$tempConcatFile" "$artefactsFile"
        chmod 600 "$tempConcatFile" "$artefactsFile"

        for file in "${files[@]}"; do
            cat "$file" >> "$tempConcatFile"
        done

        sha256sum "$tempConcatFile" | awk '{print $1}' > "$artefactsFile"
        rm -f "$tempConcatFile"
    fi
done

