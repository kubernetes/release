#!/bin/sh
set -Eeuo pipefail

#export directory to scan for yaml linting issues to env var DIR_TO_LINT
YAML_FILES_TO_LINT=$(find $DIR_TO_LINT -type f \( -name "*.yaml" -o -name "*.yml" \))
if [ -n "$YAML_FILES_TO_LINT" ]; then
    echo "Validating yaml files: \n${YAML_FILES_TO_LINT}"
    OUTPUT=$(go run ./validate_yaml.go $YAML_FILES_TO_LINT)
   if [[ "$OUTPUT" == *"YAML Validation Failed"* ]]; then
        echo "Yaml linting issues:"
        echo "$OUTPUT"
   fi
else
    echo "No YAML linting issue for $YAML_FILES_TO_LINT"
fi