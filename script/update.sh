#!/bin/bash

if [ "$1" != "1" ]; then
 	cd ..
	mkdir temp
	cp bin/log_track temp
	mkdir temp/script
	cp script/run.sh temp/script
	cp script/stop.sh temp/script
	cp -r template temp
	tar -zcvf temp.tgz temp
	rm -rf temp

	ip=118.25.137.32
	if [ "$1" != "" ]; then
		ip=$1
	fi
	scp temp.tgz script/update.sh root@$ip:/opt
	ssh root@$ip "cd /opt; ./update.sh 1; rm -f update.sh"
	rm -f temp.tgz

	echo "update done!"
else
	tar -zxvf temp.tgz
	rm -f temp.tgz
	cd temp/script
	./stop.sh
	cd ..
	mv ../log_track/bin ./
	mv log_track ./bin
	cd ..
	rm -rf log_track
	mv temp log_track
	cd log_track/script
	./run.sh
fi