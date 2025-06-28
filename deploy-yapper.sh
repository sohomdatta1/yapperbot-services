#!/bin/bash
set -e

STAGING_DIR="$1"
TOOL_PROD="/data/project/yapping-sodium/prod"

mkdir -p "$TOOL_PROD/frs"
mkdir -p "$TOOL_PROD/pruner"

cp -r "$STAGING_DIR/frs" "$TOOL_PROD/frs/frs"
cp -r "$STAGING_DIR/pruner" "$TOOL_PROD/pruner/pruner"
cp "$STAGING_DIR/config-frs.yml" "$TOOL_PROD/frs/config-frs.yml"
cp "$STAGING_DIR/config-global.yml" "$TOOL_PROD/config-global.yml"
DB_USER=$(awk -F' *= *' '/user *=/ {print $2}' /data/project/yapping-sodium/replica.my.cnf)
DB_PASS=$(awk -F' *= *' '/password *=/ {print $2}' /data/project/yapping-sodium/replica.my.cnf)
sed -E "s#dsn: [^@]+@#dsn: ${DB_USER}:${DB_PASS}@#" "$STAGING_DIR/config-pruner.yml" > "$TOOL_PROD/pruner/config-pruner.yml"
chmod 600 "$TOOL_PROD/pruner/config-pruner.yml"
cp "$STAGING_DIR/jobs.yaml" "$TOOL_PROD/jobs.yaml"

chmod 600 "$TOOL_PROD/pruner/config-pruner.yml"

if [ -f "$STAGING_DIR/botpassword" ]; then
    cp "$STAGING_DIR/botpassword" "$TOOL_PROD/botpassword"
    chmod 600 "$TOOL_PROD/botpassword"
fi