#!/bin/bash

# Helm operations script driven by per‑app configuration.
# The expected interface is:
#   helm.sh <command> <app-name>
# Supported commands: deploy, destroy, lint, validate
# All other variables (release name, values file, namespace) are read
# from apps/<app-name>/app.env so that makefiles can stay generic.

set -eu

COMMAND="${1:?Error: COMMAND not provided (deploy|destroy|lint|validate)}"
APP_NAME="${2:?Error: APP_NAME not provided}"

# load app configuration if it exists
if [ -f "apps/${APP_NAME}/app.env" ]; then
    # shellcheck disable=SC1090
    set -a; . "apps/${APP_NAME}/app.env"; set +a
fi

# expected variables from app.env
HELM_RELEASE="${HELM_RELEASE_NAME:?app.env must define HELM_RELEASE_NAME}"
HELM_VALUES="${HELM_VALUES:?app.env must define HELM_VALUES}"
NAMESPACE="${NAMESPACE:-${APP_NAME}}"   # default to app name

HELM_CHART_PATH="apps/${APP_NAME}/deploy/helm"

case "${COMMAND}" in
    deploy)
        echo "Deploying ${APP_NAME} to ${NAMESPACE} namespace..."
        helm lint "${HELM_CHART_PATH}" > /dev/null || exit 1
        helm upgrade --install "${HELM_RELEASE}" "${HELM_CHART_PATH}" \
            -f "${HELM_CHART_PATH}/${HELM_VALUES}" \
            --namespace "${NAMESPACE}" --create-namespace
        echo "";
        echo "✓ Deployment complete! Access the application:"
        echo "  - HTTP: http://${APP_NAME}.k3d.local"
        echo "  - Metrics: http://${APP_NAME}.k3d.local/metrics"
        echo "";
        echo "View deployment status:";
        echo "  kubectl get deployments -n ${NAMESPACE}";
        echo "  kubectl get pods -n ${NAMESPACE}";
        echo "  kubectl get svc -n ${NAMESPACE}";
        ;;

    destroy)
        echo "Uninstalling ${HELM_RELEASE} from ${NAMESPACE} namespace..."
        helm uninstall "${HELM_RELEASE}" -n "${NAMESPACE}" || true
        kubectl delete namespace "${NAMESPACE}" --ignore-not-found
        echo "✓ Uninstall complete"
        ;;

    lint)
        echo "Linting Helm chart..."
        helm lint "${HELM_CHART_PATH}" -f "${HELM_CHART_PATH}/${HELM_VALUES}"
        echo "✓ Lint complete"
        ;;

    validate)
        echo "Validating Helm chart (dry-run)..."
        echo "Using values file: ${HELM_CHART_PATH}/${HELM_VALUES}"
        helm template "${HELM_RELEASE}" "${HELM_CHART_PATH}" \
            -f "${HELM_CHART_PATH}/${HELM_VALUES}" \
            --namespace "${NAMESPACE}"
        ;;

    *)
        echo "Error: Unknown command '${COMMAND}'"
        echo "Valid commands: deploy, destroy, lint, validate"
        exit 1
        ;;
esac
