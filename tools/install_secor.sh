#!/bin/bash

echo "Installing Secor..."
sudo mkdir ${SECOR_INSTALL_DIR}
sudo tar -zxvf ~/secor-0.16-SNAPSHOT-bin.tar.gz -C ${SECOR_INSTALL_DIR}
sudo mv ~/start_secor.sh ${SECOR_INSTALL_DIR}/scripts/
rm ~/secor-0.16-SNAPSHOT-bin.tar.gz
sudo sed -i 's/secor.kafka.topic_filter=.*/secor.kafka.topic_filter=auction_stream/g' ${SECOR_INSTALL_DIR}/secor.common.properties
sudo sed -i 's/secor.s3.filesystem=s3n/secor.s3.filesystem=s3a/g' ${SECOR_INSTALL_DIR}/secor.common.properties
sudo sed -i 's/secor.s3.bucket=/secor.s3.bucket=adomik-firehose-dump/g' ${SECOR_INSTALL_DIR}/secor.dev.properties 
sudo sed -i 's/zookeeper.quorum=localhost/zookeeper.quorum='"$ZOOKEEPER_HOST"'/g' ${SECOR_INSTALL_DIR}/secor.dev.properties
