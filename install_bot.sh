#!/bin/bash
NAME="KlichatBot"
USERNAME="alesha"

GOHOME="/usr/local/go"
HOME=`pwd`
EXEC=${GOHOME}/bin/go
#building
${EXEC} build -o ${HOME}/build/start_demo_bot ${HOME}/src/start_demo_bot.go

#forming config
echo "
[program:${NAME}]
command=${HOME}/build/start_demo_bot
user=${USERNAME}
autostart=true
autorestart=true
stopwaitsecs=5
startsecs=5
directory=${HOME}/build/
stdout_logfile=${HOME}/logs/out.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
stderr_logfile=${HOME}/logs/out.log
stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5
" | tee /etc/supervisor/conf.d/${NAME}.conf

supervisorctl reread
supervisorctl update
