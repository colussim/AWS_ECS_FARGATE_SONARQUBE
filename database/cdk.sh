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

# Get RDS EndPoint
rds_endpoint=$(aws cloudformation describe-stacks --stack-name DatabaseStack02 --query 'Stacks[0].Outputs[?OutputKey==`RdsEndpoint`].OutputValue' --output text)

if [ "$1" = "deploy" ]; then
    cdk deploy DatabaseStack$index_value
    cdk deploy Secretup$index_value --context Endpoint=$rds_endpoint
elif [ "$1" = "destroy" ]; then
    cdk destroy --force
else
    echo "Usage: $0 [deploy|destroy]"
    exit 1
fi

