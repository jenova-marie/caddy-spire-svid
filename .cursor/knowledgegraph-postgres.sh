#!/bin/bash
export KNOWLEDGEGRAPH_STORAGE_TYPE="postgresql"
export KNOWLEDGEGRAPH_CONNECTION_STRING="postgresql://${PGUSER:-postgres}:${PGPASSWORD}@m${PGHOST}/knowledgegraph"
exec npx -y knowledgegraph-mcp "$@"