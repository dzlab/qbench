#!/bin/bash

echo "Installing pre-requisites..."
sudo yum -y install tar
sudo yum -y install wget
#sudo easy_install supervisor

echo "Installing JAVA"
wget --no-check-certificate --no-cookies --header "Cookie: oraclelicense=accept-securebackup-cookie" http://download.oracle.com/otn-pub/java/jdk/7u67-b01/jdk-7u67-linux-x64.rpm
sudo rpm -Uvh jdk-7u67-linux-x64.rpm
rm jdk-7u67-linux-x64.rpm
sudo sh -c 'echo "export JAVA_HOME=/usr/java/default" >> /etc/profile'
#sudo sh -c 'echo "export PATH=$PATH:/usr/java/default/bin" >> /etc/profile'
sudo sh -c 'echo "export CLASSPATH=.:/usr/java/default/jre/lib:/usr/java/default/lib:/usr/java/default/lib/tools.jar" >> /etc/profile'
source /etc/profile

echo "Downloading Kafka..."
BROKER_ID=$(cat /dev/urandom | tr -dc '0-9' | fold -w 3 | head -n 1)
# if BROKER_ID > 1000 then reserved.broker.max.id should be set
wget http://apache.mirrors.ovh.net/ftp.apache.org/dist/kafka/0.9.0.0/kafka_2.10-0.9.0.0.tgz
tar -zxvf kafka_2.10-0.9.0.0.tgz
rm kafka_2.10-0.9.0.0.tgz
sed -i 's/zookeeper.connect=localhost/zookeeper.connect='"$ZOOKEEPER_HOST"'/g' kafka_2.10-0.9.0.0/config/server.properties
sed -i 's/broker.id=0/broker.id='"$BROKER_ID"'/g' kafka_2.10-0.9.0.0/config/server.properties
sudo mv kafka_2.10-0.9.0.0/ /opt/kafka
sudo sh -c 'echo "# KAFKA_HOME" >> /etc/profile'
sudo sh -c 'echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile'
#sudo sh -c 'echo "export PATH=$PATH:KAFKA_HOME/bin" >> /etc/profile'
source /etc/profile
sudo sh -c 'echo "export PATH=$PATH:/usr/java/default/bin:/opt/kafka/bin" >> /etc/profile'

