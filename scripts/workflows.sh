#!/usr/bin/env bash

# read the workflow template
WORKFLOW_TEMPLATE=$(cat .github/workflow-template.yaml)

# create workflows dir if it doesnt exist
mkdir -p .github/workflows

# iterate each service in services directory
for SERVICE in $(ls services); do
    echo "generating workflow for services/${SERVICE}"

    # replace template service placeholder with service name
    WORKFLOW=$(echo "${WORKFLOW_TEMPLATE}" | sed "s/{{SERVICE}}/${SERVICE}/g")

    # save workflow to .github/workflows/{SERVICE}
    echo "${WORKFLOW}" > .github/workflows/${SERVICE}.yaml
done
