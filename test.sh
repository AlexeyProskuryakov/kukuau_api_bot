#!/usr/bin/env bash
#set -e
USERNAME="alesha"
#su -l -c "go env" ${USERNAME}

#su ${USERNAME} <<'EOF'
#go env
#EOF

sudo -u ${USERNAME} -c "go env"
#su -c "go env" ${USERNAME}
