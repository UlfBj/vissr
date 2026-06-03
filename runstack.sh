#!/bin/bash

usage() {
	echo "usage: $0 startme | stopme"
}

startme() {
	echo "Building server..."
	cd server/vissv2server
	go build && mkdir -p logs
	cd ../../

	echo "Building feederManager..."
	pushd > /dev/null feeder/feederManager && go build -o feederManager && popd > /dev/null
	sleep 1s

	echo "Building feeder(s)..."
	pushd > /dev/null feeder/feederManager && ./feederManager -b -f feederList-runstack.json && popd > /dev/null
	sleep 1s

	echo "Starting server"
	screen -S vissv2server -dm bash -c "pushd server/vissv2server && ./vissv2server -m -d &> ./logs/vissv2server-log.txt && popd"

	sleep 1s
	echo "Starting feeder(s)"
	pushd > /dev/null feeder/feederManager && ./feederManager -s -f feederList-runstack.json && popd > /dev/null

	screen -list
}

stopme() {
	echo "Terminating feeder(s)"
	pushd > /dev/null feeder/feederManager && ./feederManager -t -f feederList-runstack.json && popd > /dev/null

	echo "Terminating vissv2server"
	screen -X -S vissv2server quit

	sleep 1s
	screen -wipe
}

if [ $# -ne 1 ];
then
	usage $0
	exit 1
fi

case "$1" in 
	startme)
		stopme
		startme
		;;
	stopme)
		stopme
		;;
	*)
		usage
		exit 1
		;;
esac
