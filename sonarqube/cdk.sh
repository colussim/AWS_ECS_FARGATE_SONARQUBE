#!/bin/bash

# Lire la valeur de "Index" du fichier JSON
json_file="../config_crd.json"

if ! command -v jq &> /dev/null; then
    echo "jq is not installed. Please install it first."
    exit 1
fi

index_value=$(jq -r '.Index' "$json_file")

if [ "$1" = "deploy" ]; then
   
     cdk deploy SonarqubeStack"$index_value" --context deploySonarqube=true --context deployGetPublicIP=false
     cdk deploy GetPublicIP"$index_value deploySonarqube=false --context deployGetPublicIP=true"
elif [ "$1" = "destroy" ]; then
    cdk destroy SonarqubeStack"$index_value" --context deploySonarqube=true --context deployGetPublicIP=false --force
else
    echo "Usage: $0 [deploy|destroy]"
    exit 1
fi
