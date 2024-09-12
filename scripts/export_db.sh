#!/bin/env bash
#
# File: export_db.sh
# Description: exports the entire integration test database for versioning in git

if [ '$(which docker)' = '' ]; then
	echo "docker not found or not installed"
	exit 1
fi

if [ '$(docker ps | grep testing-postgres-1)' = '' ]; then
	echo "failed to find netbox's postgres container"
	exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel)

# if given, use version from parameters
if [ "$#" -eq 1 ]; then
	export VERSION="_v"$1
else
  #export VERSION="v3.7.8"
  export VERSION=""
fi

docker exec testing-postgres-1 pg_dump -U netbox -w -d netbox > ${REPO_ROOT}/testdata/sql/netbox${VERSION}.sql

if [ "$?" -ne '0' ]; then
	echo "failed to export database"
	exit 1
fi

echo "finished exporting database to ${REPO_ROOT}/testdata/sql/netbox.sql"
