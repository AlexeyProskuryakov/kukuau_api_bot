#!/bin/bash
NAME="KlichatBot"
USERNAME="alesha"

GOHOME="/usr/local/go"
HOME=`pwd`
EXEC=${GOHOME}/bin/go

#building
${EXEC} build -o ${HOME}/build/start_demo_bot ${HOME}/src/start_demo_bot.go

#forming config

build(){
    GOPATH=${HOME}
    ${EXEC} get github.com/looplab/fsm
    ${EXEC} get github.com/tealeg/xlsx
    ${EXEC} get gopkg.in/mgo.v2
    ${EXEC} get github.com/go-martini/martini
    ${EXEC} get github.com/martini-contrib/auth
    ${EXEC} get github.com/martini-contrib/render
    ${EXEC} get gopkg.in/olivere/elastic.v2


    #building
    rm -rf ${HOME}/build
    mkdir ${HOME}/build
    ${EXEC} build -o ${HOME}/build/start_bot ${HOME}/src/start_bot.go
    cp ${HOME}/config.json ${HOME}/build
    cp -r ${HOME}/templates ${HOME}/build
    cp -r ${HOME}/static ${HOME}/build
}

install() {
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

supervisorctl reread
supervisorctl update

supervisorctl restart ${NAME}

 }

case "$1" in
    build)
        build
            ;;
    install)
        install
            ;;
    *)
            echo "usage: {build|install (with sudo please...)}" >&2
            exit 3
            ;;
esac

