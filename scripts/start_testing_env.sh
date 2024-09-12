#!/bin/env bash
#
# File: start_testing_env.sh
# Description: pulls netbox-docker and spins up a netbox environment to test against
export REPO_ROOT=$(git rev-parse --show-toplevel)

# if given, use version from parameters
if [ "$#" -eq 1 ]; then
	export VERSION="v"$1
fi

# pull netbox-docker
git clone https://github.com/netbox-community/netbox-docker.git .testing || true

# replace original compose manifest
cp ${REPO_ROOT}/testdata/docker-compose.override.yml .testing/docker-compose.override.yml

if [ ! -f ${REPO_ROOT}/testdata/sql/netbox_${VERSION}.sql ]; then
	# unless a version specific sql file exists, use the default
	export DB_DATA="${REPO_ROOT}/testdata/sql/netbox.sql"
else
	export DB_DATA="${REPO_ROOT}/testdata/sql/netbox_${VERSION}.sql"
fi

$(cd ${REPO_ROOT}/.testing && docker-compose up -d)
