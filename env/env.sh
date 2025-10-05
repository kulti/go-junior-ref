#!/bin/bash

set -euo pipefail

function main() {
    ENV_DIR="$(cd "$(dirname "$0")" && pwd)"
    cd "$ENV_DIR"

    readonly PROJECT_NAME=${DOCKER_COMPOSE_PROJECT_NAME:-dev}

    export IMAGE_TAG=${IMAGE_TAG:-latest}
    export IMAGE_TASK_LIST=${IMAGE_TASK_LIST:-task-list:$IMAGE_TAG}
    export IMAGE_TASK_LIST_MIGRATE=${IMAGE_TASK_LIST_MIGRATE:-task-list-migrate:$IMAGE_TAG}

    local cmd=$1
    shift || true

    case "${cmd}" in
    run | down | ps | exec | start | stop | restart | logs)
        docker compose -p "${PROJECT_NAME}" "${cmd}" "$@"
        ;;
    up)
        docker compose -p "${PROJECT_NAME}" up -d --remove-orphans --wait "$@"
        ;;
    recreate)
        docker compose -p "${PROJECT_NAME}" up --remove-orphans -d --no-deps --wait "$@"
        ;;
    clean)
        docker compose -p "${PROJECT_NAME}" down -v --remove-orphans
        ;;
    psql)
        docker compose -p "${PROJECT_NAME}" exec -u postgres postgres psql "$@"
        ;;
    *)
        if [[ -n "${cmd}" ]]; then
            echo "Error: Unsupported command: ${cmd}"
            echo
        fi
        exit 1
        ;;
    esac
}

main "$@"
