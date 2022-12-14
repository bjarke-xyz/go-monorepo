#!/usr/bin/env bash

function parse_yaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
   }'
}

# read the workflow template
WORKFLOW_TEMPLATE=$(cat .github/workflow-template.yaml)

# create workflows dir if it doesnt exist
mkdir -p .github/workflows

# iterate each service in services directory
for SERVICE in $(ls services); do
    echo "generating workflow for services/${SERVICE}"

    eval $(parse_yaml "services/${SERVICE}/workflow.yaml" "VAL_")

    # replace template service placeholder with service name
    WORKFLOW=$(echo "${WORKFLOW_TEMPLATE}" | sed -e "s/{{NAMESPACE}}/$VAL_NAMESPACE/g" -e "s/{{WORKLOADNAME}}/$VAL_WORKLOADNAME/g" -e "s/{{WORKLOADKIND}}/$VAL_WORKLOADKIND/g" -e "s/{{SERVICE}}/${SERVICE}/g" -e "s/{{TEMPLATEWARNING}}/DO NOT EDIT THIS FILE, IT IS GENERATED/g")

    # save workflow to .github/workflows/{SERVICE}
    echo "${WORKFLOW}" > .github/workflows/${SERVICE}.yaml
done


