#!/bin/bash
set -e

USER=$1
UPLOAD_BOTPASSWORD=$2
STAGING_DIR="/home/$USER/tmp/tmp-deploy-$$"


if [ -z "$1" ]; then
    echo "🐶 BARK! You forgot to provide your Toolforge username."
    echo "Usage: $0 <TOOLFORGE_USERNAME> [--upload-botpassword]"
    exit 1
fi

# Build your binaries
make

# Prepare remote staging dir (clean first)
ssh "$USER@login.toolforge.org" "rm -rf $STAGING_DIR && mkdir -p $STAGING_DIR"

# Upload binaries, configs, and the deploy script
scp ./frs/frs "$USER@login.toolforge.org:$STAGING_DIR/frs"
scp ./pruner/pruner "$USER@login.toolforge.org:$STAGING_DIR/pruner"
scp ./frs/config-frs.yml "$USER@login.toolforge.org:$STAGING_DIR/config-frs.yml"
scp ./config.yml "$USER@login.toolforge.org:$STAGING_DIR/config-global.yml"
scp ./pruner/config-pruner.yml "$USER@login.toolforge.org:$STAGING_DIR/config-pruner.yml"
scp ./deploy-yapper.sh "$USER@login.toolforge.org:$STAGING_DIR/deploy-yapper.sh"
scp ./jobs.yaml "$USER@login.toolforge.org:$STAGING_DIR/jobs.yaml"
# Upload botpassword if flag is passed and file exists
if [ "$UPLOAD_BOTPASSWORD" == "--upload-botpassword" ] && [ -f botpassword ]; then
    scp botpassword "$USER@login.toolforge.org:$STAGING_DIR/botpassword"
fi

# Run the deploy script as the tool user
ssh "$USER@login.toolforge.org" "become yapping-sodium bash $STAGING_DIR/deploy-yapper.sh $STAGING_DIR"
ssh "$USER@login.toolforge.org" "become yapping-sodium bash dologmsg '$USER built and uploaded a new version'"

rm -rf "/home/$USER/tmp"

echo "You should now be ready to run toolforge jobs load /data/project/yapping-sodium/prod/jobs.yaml if you want to reload the tasks"
