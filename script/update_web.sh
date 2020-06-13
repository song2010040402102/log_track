#!/bin/bash

if [ "$1" != "1" ]; then
 	cd ..		
	tar -zcvf template.tgz template	

	ip=118.25.137.32
	if [ "$1" != "" ]; then
		ip=$1
	fi
	scp template.tgz script/update_web.sh root@$ip:/opt
	ssh root@$ip "cd /opt; ./update_web.sh 1; rm -f update_web.sh"
	rm -f template.tgz

	echo "update done!"
else
	cd log_track
	rm -rf template
	tar -zxvf ../template.tgz	
	rm -f template.tgz
	cd ..
fi