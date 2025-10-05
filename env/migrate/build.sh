#!/bin/bash

set -eo pipefail

cd "$(dirname "$0")" || exit 1

docker build -t task-list-migrate .
