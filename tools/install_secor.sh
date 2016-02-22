#!/bin/bash

echo "Installing Secor..."
sudo mkdir ${SECOR_INSTALL_DIR}
sudo tar -zxvf ~/secor-0.16-SNAPSHOT-bin.tar.gz -C ${SECOR_INSTALL_DIR}
sudo mv ~/start_secor.sh ${SECOR_INSTALL_DIR}/scripts/
rm ~/secor-0.16-SNAPSHOT-bin.tar.gz
# configure kafka topic
sudo sed -i 's/secor.kafka.topic_filter=.*/secor.kafka.topic_filter=auction_stream/g' ${SECOR_INSTALL_DIR}/secor.common.properties
# configure S3 connection
sudo sed -i 's/aws.access.key=.*/aws.access.key='"${AWS_ACCESS_KEY_ID}"'/g' ${SECOR_INSTALL_DIR}/secor.common.properties
sudo sed -i 's/aws.secret.key=.*/aws.secret.key='"${AWS_SECRET_ACCESS_KEY}"'/g' ${SECOR_INSTALL_DIR}/secor.common.properties
#sudo sed -i 's/secor.s3.filesystem=s3n/secor.s3.filesystem=s3a/g' ${SECOR_INSTALL_DIR}/secor.common.properties
sudo sed -i 's/secor.s3.bucket=.*/secor.s3.bucket=adomik-firehose-dump/g' ${SECOR_INSTALL_DIR}/secor.prod.properties 
sudo sed -i 's/secor.max.file.size.bytes=.*/secor.max.file.size.bytes=50000000/g' ${SECOR_INSTALL_DIR}/secor.prod.properties 
# configure zookeeper connection
sudo sed -i 's/zookeeper.quorum=.*/zookeeper.quorum='"$ZOOKEEPER_HOST"':2181/g' ${SECOR_INSTALL_DIR}/secor.prod.properties
# configure kafka broker connection
sudo sed -i 's/kafka.seed.broker.host=.*/kafka.seed.broker.host=localhost/g' ${SECOR_INSTALL_DIR}/secor.prod.properties
sudo sed -i 's/kafka.seed.broker.port=.*/kafka.seed.broker.port=9092/g' ${SECOR_INSTALL_DIR}/secor.prod.properties
# configure secor logging 
sudo sed -i 's/log4j.rootLogger=.*/log4j.rootLogger=DEBUG, CONSOLE/g' ${SECOR_INSTALL_DIR}/log4j.prod.properties
sudo sed -i 's/log4j.appender.CONSOLE.Threshold=.*/log4j.appender.CONSOLE.Threshold=INFO/g' ${SECOR_INSTALL_DIR}/log4j.prod.properties
sudo sed -i 's:secor.local.path=.*:secor.local.path=/tmp/secor_data/message_logs/backup:g' ${SECOR_INSTALL_DIR}/secor.prod.backup.properties
sudo sed -i 's:log4j.appender.ROLLINGFILE.File=/mnt/:log4j.appender.ROLLINGFILE.File=/tmp/:g' ${SECOR_INSTALL_DIR}/log4j.prod.properties
# move config files
#sudo mkdir ${SECOR_INSTALL_DIR}/config
#sudo mv ${SECOR_INSTALL_DIR}/*.properties ${SECOR_INSTALL_DIR}/config
