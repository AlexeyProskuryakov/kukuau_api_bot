### BEGIN INIT INFO
# Provides:          sayme_quotations
# Required-Start:    $local_fs $remote_fs $network $syslog
# Required-Stop:     $local_fs $remote_fs $network $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: starts the Sayme demo bot client
# Description:       starts Sayme demo bot client using start-stop-daemon
### END INIT INFO

GOHOME="/usr/local/go"
HOME="/home/alesha/go"
EXEC=${GOHOME}/bin/go
PID=${HOME}/sdb.pid
LOG=${HOME}/logs/sdb.logs


start()
{   
    export GOPATH=${HOME}
    ${EXEC} get github.com/looplab/fsm
    ${EXEC} build ${HOME}/src/start_demo_bot.go
    ${HOME}/start_demo_bot >  ${LOG} 2>&1 &
    pidof start_demo_bot > ${PID} 
}

stop()
{
    pid=$(<${PID})
    kill ${pid}
    rm ${PID}
}

case "$1" in
    start)
        start
            ;;
    stop)
        stop 
            ;;
    restart)
        if [ -f "$PID" ]; then
            stop
            start
        else
            echo "service not running, will do nothing"
            exit 1
        fi
            ;;
    *)
            echo "usage: daemon {start|stop|restart}" >&2
            exit 3
            ;;
esac
