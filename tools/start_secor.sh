#!/bin/bash

# Wait until Kafka started and listens on port 9092.
while [ -z "`netstat -tln | grep 9092`" ]; do
  echo 'Waiting for Kafka to start ...'
  sleep 1
done
echo 'Kafka started.'

# Stop child process if this main process dies
trap "kill -- -$$" EXIT

cd ${SECOR_INSTALL_DIR} && java -ea -Dsecor_group=secor_backup \
	-Dlog4j.configuration=log4j.prod.properties \
	-Dconfig=secor.prod.backup.properties \
	-cp secor-0.16-SNAPSHOT.jar:lib/* \
	com.pinterest.secor.main.ConsumerMain
