#!/bin/bash
NAME="KlichatBot"
USERNAME="alesha"

GOHOME="/usr/local/go"
HOME=`pwd`
EXEC=${GOHOME}/bin/go

#ensuring libs
${EXEC} get github.com/looplab/fsm
${EXEC} get gopkg.in/mgo.v2
${EXEC} get github.com/go-martini/martini
${EXEC} get github.com/martini-contrib/auth
${EXEC} get github.com/martini-contrib/render
${EXEC} get gopkg.in/olivere/elastic.v2

#building
${EXEC} build -o ${HOME}/build/start_bot ${HOME}/src/start_bot.go
cp ${HOME}/config.json ${HOME}/build
cp -r ${HOME}/templates ${HOME}/build

#forming config
echo "
[program:${NAME}]
command=${HOME}/build/start_bot
directory=${HOME}/build/
user=${USERNAME}
autostart=true
autorestart=true
stopwaitsecs=5
startsecs=5
stdout_logfile=${HOME}/logs/out.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
stderr_logfile=${HOME}/logs/out.log
stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5
" | tee /etc/supervisor/conf.d/${NAME}.conf

#restarting supervisor

supervisorctl reread
supervisorctl update

