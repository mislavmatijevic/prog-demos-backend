#!/bin/bash

executionDirectory="/playground"

sourceFileName="$SOURCE_FILE_NAME"
mainDirectory="$SOURCE_CODE_FOLDER"
solutionFile="$mainDirectory/$SOURCE_FILE_NAME"
errorFile="$mainDirectory/$ERROR_FILENAME"
outputFile="$mainDirectory/solution.out"
stdinFilenamePrefix="$STDIN_FILENAME_PREFIX"
stdoutFilenamePrefix="$STDOUT_FILENAME_PREFIX"
artefactsFilenamePrefix="$ARTEFACTS_FILENAME_PREFIX"

clang++ -O0 "$solutionFile" -o "$outputFile" 2> "$errorFile"

ulimit -u 10 -f 10 -v 10000

if [ $? -ne 0 ]; then
    cat "$errorFile"
    exit 0
fi

chmod 701 "$mainDirectory"
chmod 701 "$outputFile"

find "$mainDirectory" -type f -name "${stdinFilenamePrefix}*" | while read -r stdinFile; do
    taskId=$(basename "$stdinFile" | sed -E 's/[^0-9]*([0-9]+).*/\1/')

    stdoutFile="$mainDirectory/${stdoutFilenamePrefix}${taskId}.txt"
    touch "$stdoutFile"
    chmod 600 "$stdoutFile"

    (cd $executionDirectory && exec runuser -u tester -- "$outputFile" < "$stdinFile" > "$stdoutFile" 2>> "$errorFile")

    if [ -s "$errorFile" ]; then
        cat "$errorFile"
        exit 0
    else
        rm -f "$errorFile"
    fi

    files=($(find "$executionDirectory" -type f | sort))

    if [[ ${#files[@]} -gt 0 ]]; then
        tempConcatFile="$mainDirectory/artefacts_concat.txt"
        artefactsFile="$mainDirectory/${artefactsFilenamePrefix}${taskId}.txt"
        touch "$tempConcatFile" "$artefactsFile"
        chmod 600 "$tempConcatFile" "$artefactsFile"

        for file in "${files[@]}"; do
            cat "$file" >> "$tempConcatFile"
        done

        sha256sum "$tempConcatFile" | awk '{print $1}' > "$artefactsFile"
        rm -f "$tempConcatFile"
    fi
done
