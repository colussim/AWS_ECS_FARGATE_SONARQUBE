#!/bin/bash

# Path to your JSON file
json_file="../config_crd.json"

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "jq is not installed. Please install it first."
    exit 1
fi

# Read the value of "Index" from the JSON file
index_value=$(jq -r '.Index' "$json_file")


if [ "$1" = "deploy" ]; then
    cdk deploy SonarqubeStack$index_value
    cdk deploy GetPublicIP$index_value 
elif [ "$1" = "destroy" ]; then
    cdk destroy SonarqubeStack$index_value --force
else
    echo "Usage: $0 [deploy|destroy]"
    exit 1
fi

