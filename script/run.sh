#!/bin/bash

if [ ! -f "../bin/log_track" ]; then
	./make.sh
fi

cd ../bin
if [ ! -d "data" ]; then
	mkdir data
fi

if [ ! -d "data/log" ]; then
	mkdir data/log
fi

if [ ! -d "download" ]; then
	mkdir download
else
	rm -f download/temp_*
fi

if [ ! -d "upload" ]; then
  mkdir upload
else
	rm -f upload/temp_*
fi

if [ ! -d "tmp" ]; then
	mkdir tmp
fi

while [ "$(ps -e > /tmp/psi && grep log_track /tmp/psi)" != "" ]
do
	sleep 1
done

env GOTRACEBACK=crash nohup ./log_track stat > stat.log 2>&1 &
env GOTRACEBACK=crash nohup ./log_track plat > plat.log 2>&1 &