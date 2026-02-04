#!/bin/bash

usage() {
	echo "usage: $0 startme|stopme"
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
	pushd > /dev/null feeder/feederManager && ./feederManager -b -f feederList-runtest.json && popd > /dev/null
	sleep 1s

	echo "Starting server"
	screen -S vissv2server -dm bash -c "pushd server/vissv2server && ./vissv2server -m &> ./logs/vissv2server-log.txt && popd"

	sleep 1s
	echo "Starting feeder(s)"
	pushd > /dev/null feeder/feederManager && ./feederManager -s -f feederList-runtest.json && popd > /dev/null

	sleep 2s
	echo "Building testClient..."
	cd client/client-1.0/testClient
	go build
	cd ../../../

	echo "Starting testClient"
	screen -S testClient bash -c "pushd client/client-1.0/testClient && go build && ./testClient -m 4 && popd"

	screen -list
}

stopme() {
	echo "Terminating feeder(s)"
	pushd > /dev/null feeder/feederManager && ./feederManager -t -f feederList-runtest.json && popd > /dev/null

	echo "Terminating vissv2server"
	screen -X -S vissv2server quit

	sleep 1s
	screen -wipe
}

if [ $# -ne 1 ]
then
	usage $0
	exit 1
fi

case "$1" in 
	startme)
		stopme
		startme ;;
	stopme)
		stopme
		;;
	*)
		usage
		exit 1
		;;
esac
