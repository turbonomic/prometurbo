#!/bin/env bash

SCRIPT_DIR="$(cd "$(dirname $0)" && pwd)"
ERR_LOG=$(mktemp --suffix _prome.errlog)
WAIT_FOR_DEPLOYMENT=60
OPERATOR_IMAGE="icr.io\/cpopen\/prometurbo-operator:${OPERATOR_IMAGE_VERSION}"
PROMETURBO_IMAGE_REPO="icr.io/cpopen/turbonomic/prometurbo"
KUBECTL=kubectl
# username and password for the local ops-manager
OPS_MANAGER_USERNAME=administrator
OPS_MANAGER_PASSWORD=administrator
# assume prometurbo operator and prometurbo are always in the same ns
TEST_NS=turbo
# container used to search for logs
CONTAINER_NAME=prometurbo

# turbonomic is the default namespace that matches the xl setup
# please change the value according to your set up
export namespace=turbonomic

function rediscover_target {
	DISPLAY_NAME=$1
    ip=$($KUBECTL get services --no-headers -n $namespace | grep nginx | awk '{print $4; exit}' | cut -d "," -f 1)
    cookie=$(curl -k -s -v "https://$ip/vmturbo/rest/login" --data "username=$OPS_MANAGER_USERNAME&password=$OPS_MANAGER_PASSWORD" 2>&1 | grep JSESSION | awk -F'=' '{print $2}' | awk -F';' '{print $1}')
	response=$(curl -k -s --cookie "JSESSIONID=$cookie"-X GET "https://$ip/vmturbo/rest/targets?q=$DISPLAY_NAME&target_category=Custom&order_by=validation_status&ascending=true&query_method=regex" -H "accept: application/json")
	target_uuid=$(echo "$response" | jq '. | .[] | "\(.uuid)"' | tr -d '"')
	# rediscover
	curl -k -s --cookie "JSESSIONID=$cookie"-X POST "https://${ip}/vmturbo/rest/targets/${target_uuid}?rediscover=true" -H "accept: application/json" -d '' > /dev/null
}

function install_operator {
	NS=$1
	[ -z "${NS}" ] && echo -e "Operator namespace not provided" | tee -a ${ERR_LOG} && exit 1
	$KUBECTL create ns ${NS}
	# $KUBECTL delete -f ${SCRIPT_DIR}/deploy/crds/*crd.yaml
	$KUBECTL apply -f ${SCRIPT_DIR}/deploy/crds/*crd.yaml
	$KUBECTL apply -f ${SCRIPT_DIR}/deploy/service_account.yaml -n ${NS}
	$KUBECTL apply -f ${SCRIPT_DIR}/deploy/role_binding.yaml
	$KUBECTL apply -f ${SCRIPT_DIR}/deploy/prometurbo-operator-cluster-role.yaml
	sed "s/image:.*/image: ${OPERATOR_IMAGE}/g" ${SCRIPT_DIR}/deploy/operator.yaml > ${SCRIPT_DIR}/deploy/updated_operator.yaml
	$KUBECTL apply -f ${SCRIPT_DIR}/deploy/updated_operator.yaml -n ${NS}
}

function uninstall_operator {
	NS=$1
	[ -z "${NS}" ] && echo -e "Operator namespace not provided" | tee -a ${ERR_LOG} && exit 1
	$KUBECTL delete -f ${SCRIPT_DIR}/deploy/updated_operator.yaml -n ${NS}
	rm ${SCRIPT_DIR}/deploy/updated_operator.yaml
	$KUBECTL delete -f ${SCRIPT_DIR}/deploy/role_binding.yaml
	$KUBECTL delete -f ${SCRIPT_DIR}/deploy/service_account.yaml -n ${NS}
	$KUBECTL delete ns ${NS}
}

function create_cr {
	CR_SUFFIX=$1
	CLUSTER_ROLE=${2-"cluster-admin"}

	# if we have .turbocreds pull the credentials from it instead
	FILE="~/.turbocreds"
	if [ -f "${FILE}" ]; then
		if [[ $(grep -c "opsman_username" ${FILE}) -eq 1 ]]; then
			OPS_MANAGER_USERNAME=`grep "opsman_username" ${FILE} | cut -d'=' -f2`
			OPS_MANAGER_PASSWORD=`grep "opsman_password" ${FILE} | cut -d'=' -f2`
		fi
	fi

	command -v xl_version &> /dev/null
	[ $? -gt 0 ] && echo -e "Failed to invoke xl_version from environemnt" | tee -a ${ERR_LOG} && exit 1
	XL_VERSION_DETAIL=$(xl_version)

	# generated testing cr file based on the given input
	HOST_IP=$(echo -e "${XL_VERSION_DETAIL}" | grep "Server:" | awk '{print $2}')
	[ -z "${HOST_IP}" ] && echo -e "Failed to get exposed XL IP" | tee -a ${ERR_LOG} && exit 1

	XL_VERSION=$(echo -e "${XL_VERSION_DETAIL}" | grep "Version:" | awk '{print $2}')
	[ -z "${XL_VERSION}" ] && echo -e "Failed to get exposed XL version" | tee -a ${ERR_LOG} && exit 1

	CR_FILEPATH=$(mktemp --suffix _prometurbo_cr_${CR_SUFFIX}.yaml)
	echo ${CR_FILEPATH}

	cat > ${CR_FILEPATH} <<- EOT
	apiVersion: charts.helm.k8s.io/v1
	kind: Prometurbo
	metadata:
	  name: prometurbo-${CR_SUFFIX}
	spec:
	  serverMeta:
	    version: ${XL_VERSION}
	    turboServer: https://${HOST_IP}
	  restAPIConfig:
	    turbonomicCredentialsSecretName: "turbonomic-credentials"
	    opsManagerUserName: ${OPS_MANAGER_USERNAME}
	    opsManagerPassword: ${OPS_MANAGER_PASSWORD}
	  targetConfig:
	    targetName: ${CR_SUFFIX}
	  roleName: ${CLUSTER_ROLE}
	  image:
	    prometurboRepository: ${PROMETURBO_IMAGE_REPO}
	    prometurboTag: ${PROMETURBO_IMAGE_VERSION}
	    turbodifTag: ${PROMETURBO_IMAGE_VERSION}
	    pullPolicy: Always
	EOT
}

function install_cr {
	# install cr in a given namespace and run checks
	NS=$1
	CR_FILE=$2
	TARGET_NAME=$3
	TEST_DESC=${4-"Install Prometurbo CR in ${NS}"}

	[ -z "${NS}" ] && echo "Namespace not provided" | tee -a ${ERR_LOG} && return
	[ ! -f "${CR_FILE}" ] && echo "CR file ${CR_FILE} not provided" | tee -a ${ERR_LOG} && return

	echo -e "> Start testing for ${TEST_DESC}"
	$KUBECTL apply -f ${CR_FILE} -n ${NS} && echo -e "Wait for ${WAIT_FOR_DEPLOYMENT}s to finish container creation"
	sleep ${WAIT_FOR_DEPLOYMENT}

	# rediscover target in case initial discovery doesn't show up
	rediscover_target $TARGET_NAME && echo -e "Wait for 10s to finish rediscovering"
	sleep 10

	# check if the prometurbo deployment get generated
	FAILED=0
	# assumes there is only one deployment in the ns aside from operator deployment
	CR_DEPLOY=$($KUBECTL get deploy -n ${NS} -o NAME | grep -v operator)
	if [ -z "${CR_DEPLOY}" ]; then
		echo -e "> ${TEST_DESC}........FAILED" | tee -a ${ERR_LOG}
		return
	elif [ -z "$($KUBECTL rollout status ${CR_DEPLOY} -n ${NS} | grep successfully)" ]; then
		FAILED=$((${FAILED}+1))
		echo -e ">  Deployment ${CR_DEPLOY} cr check........FAILED" | tee -a ${ERR_LOG}
	else
		echo -e ">  Deployment ${CR_DEPLOY} cr check........PASSED"
	fi

	# check if the cluster log is healthy
	if [[ -z "$($KUBECTL logs ${CR_DEPLOY} -c ${CONTAINER_NAME} -n ${NS} | grep "Discovered .* entities")" ]]; then
		FAILED=$((${FAILED}+1))
		echo -e ">  Deployment ${CR_DEPLOY} log check........FAILED" | tee -a ${ERR_LOG}
	else
		echo -e ">  Deployment ${CR_DEPLOY} log check........PASSED"
	fi

	CR_DEPLOY_NAME=$($KUBECTL get ${CR_DEPLOY} -n ${NS} -o jsonpath={.metadata.name})

	# check if the correct cluster role binding get generated
	TARGET_ROLEBINDING="prometurbo-binding-${CR_DEPLOY_NAME}-${NS}"
	$KUBECTL get ClusterRoleBinding "${TARGET_ROLEBINDING}" >/dev/null
	if [ $? -gt 0 ]; then
		FAILED=$((${FAILED}+1))
		echo -e ">  ClusterRoleBinding ${TARGET_ROLEBINDING} check........FAILED" | tee -a ${ERR_LOG}
	else
		echo -e ">  ClusterRoleBinding ${TARGET_ROLEBINDING} check........PASSED"
	fi

	TARGET_ROLENAME=$(grep roleName ${CR_FILE} | awk '{print $2}')
	if [ -z "${TARGET_ROLENAME}" ]; then
		TARGET_ROLENAME="cluster-admin"
	elif [ "${TARGET_ROLENAME}" != "cluster-admin" ]; then
		TARGET_ROLENAME="${TARGET_ROLENAME}-${CR_DEPLOY_NAME}-${NS}"
	fi
	$KUBECTL get ClusterRole "${TARGET_ROLENAME}" >/dev/null
	if [ $? -gt 0 ]; then
		FAILED=$((${FAILED}+1))
		echo -e ">  ClusterRole ${TARGET_ROLENAME} check........FAILED" | tee -a ${ERR_LOG}
	else
		echo -e ">  ClusterRole ${TARGET_ROLENAME} check........PASSED"
	fi

	# summarize the deployment test
	if [ ${FAILED} -gt 0 ]; then
		echo -e "> ${TEST_DESC}........FAILED" | tee -a ${ERR_LOG}
	else
		echo -e "> ${TEST_DESC}........PASSED"
	fi
}

function uninstall_cr {
	# unstall cr in a given namespace
	NS=$1
	CR_FILE=$2

	if [ -f "${CR_FILE}" ] && [ -n "${NS}" ]; then
		$KUBECTL delete -f ${CR_FILE} -n ${NS}
	fi

	[ -f "${CR_FILE}" ] && rm ${CR_FILE}
}

function update_cr_for_logging {
	TEST_NS=$1
	CR_FILE1=$2
	LOGGING_LEVEL=$3
	echo -e "Updating CR to add logging level..."
	# existing method: adding spec.args.logginglevel causes pod to restart
	# sed -i -e "$ a\  args:" $CR_FILE1
	# sed -i -e "$ a\    logginglevel: 4" $CR_FILE1
	# new method: adding spec.logging.level should not cause a restart
	sed -i -e '$ a\  logging:' $CR_FILE1
	sed -i -e "$ a\    level: $LOGGING_LEVEL" $CR_FILE1	
	$KUBECTL apply -f ${CR_FILE} -n ${NS}
}

function wait_for_logging_update {
	NAMESPACE=$1
	POD_NAME=$2
	MSG=""
	declare -i COUNT=0

	while [ -z "$MSG" ]
	do 
		echo "Wait for 10s for log level changes to be reflected..."
		sleep 10
		MSG=$($KUBECTL -n ${NAMESPACE} logs ${POD_NAME} -c ${CONTAINER_NAME} | grep "Logging level is changed from")
		if [ $COUNT -eq 15 ]; then
			echo "Timed out waiting for log level changes..." | tee -a ${ERR_LOG} && return
		fi
		COUNT+=1
	done
	echo $MSG
}

function test_dynamic_logging {
	echo -e "> Running dynamic logging tests"
	NEW_LOGGING_LEVEL=4
	NEW_LOGLEVEL_MSG="Begin to handle path" # example level 4 msg
	install_operator ${TEST_NS}
	CR_LABEL=testcr1
	CR_FILE1=$(create_cr ${CR_LABEL})
	TARGET_NAME=Prometheus
	install_cr ${TEST_NS} ${CR_FILE1} ${TARGET_NAME} "Deploy prometurbo for dynamic logging"
	POD_NAME=$($KUBECTL get pod -n ${NS} | grep ${CR_LABEL} | awk '{print $1}')
	NUM_RESTART_BEFORE=$($KUBECTL get pod -n ${NS} ${POD_NAME} | awk 'FNR == 2 {print $4}')
	HEAD_LOGS_BEFORE=$($KUBECTL -n ${NS} logs ${POD_NAME} -c ${CONTAINER_NAME} | head -n 2)

	# DEBUGGING ONLY
	echo "-------------Pod log before-----------------------------------------"
	$KUBECTL -n ${NS} logs ${POD_NAME} -c ${CONTAINER_NAME}
	echo "-------------End Pod log before-----------------------------------------"
	update_cr_for_logging ${TEST_NS} ${CR_FILE1} ${NEW_LOGGING_LEVEL}

	# wait until log level changes msg shows up
	wait_for_logging_update ${NS} ${POD_NAME}

	# rediscover prometurbo to generate new logs
	rediscover_target $TARGET_NAME	&& echo -e "Wait for 10s to finish rediscovering"
	sleep 10

	# check to make sure new logging level is updated in the configmap
	# expecting something like this in the configmap "{\n \"logging\": {\n \"level\": 5\n }\n}"
	LOGLEVEL_CM=$($KUBECTL get cm prometurbo-config-prometurbo-${CR_LABEL} -n ${TEST_NS} -o json | jq '.data."turbo-autoreload.config"')
	SEARCH_STR='level\\": '$NEW_LOGGING_LEVEL
	[[ -z $(echo $LOGLEVEL_CM | grep "$SEARCH_STR") ]] && echo "Error: incorrect turbo-autoreload.config $LOGLEVEL_CM" | tee -a ${ERR_LOG}	

	# check to make sure pod still exists
	[[ $($KUBECTL get pod -n ${NS} ${POD_NAME} | grep "not found") ]] && echo "Error: pod not found after update" | tee -a ${ERR_LOG}	
	
	NUM_RESTART_AFTER=$($KUBECTL get pod -n ${NS} ${POD_NAME} | awk 'FNR == 2 {print $4}')
	HEAD_LOGS_AFTER=$($KUBECTL -n ${NS} logs ${POD_NAME} -c ${CONTAINER_NAME} | head -n 2)
	# check the pod restart number before and after cr update
	[[ $NUM_RESTART_BEFORE != $NUM_RESTART_AFTER ]] && echo "Pod restarted after updating logging level" | tee -a ${ERR_LOG}	
	# check if previous logs are preserved
	[[ $HEAD_LOGS_BEFORE != $HEAD_LOGS_AFTER ]] && echo "Logs before CR update was erased\n$HEAD_LOGS_BEFORE\n$HEAD_LOGS_AFTER" | tee -a ${ERR_LOG}
	# check if new logs contains message that should only show up in the new logging level
	[[ -z $($KUBECTL -n ${NS} logs ${POD_NAME} -c ${CONTAINER_NAME} | grep "$NEW_LOGLEVEL_MSG") ]] && echo "Error: expected new log not found" | tee -a ${ERR_LOG}	

	# DEBUGGING ONLY
	echo "-------------Pod log after-----------------------------------------"
	$KUBECTL -n ${NS} logs ${POD_NAME} -c ${CONTAINER_NAME}
	echo "-------------End Pod log after-----------------------------------------"

	echo -e "Cleanup..."
	uninstall_cr ${TEST_NS} ${CR_FILE1}
	uninstall_operator ${TEST_NS}

	TEST_RESULT=$(cat ${ERR_LOG})
	if [ -n "${TEST_RESULT}" ]; then
		echo -e "Dynamic logging test failed see logs at: ${ERR_LOG}" && exit 1
	else
		echo "Test passed!"
	fi
}

function main {	
	# test_prometurbo_setup
	test_dynamic_logging
	rm ${ERR_LOG}
}

main
