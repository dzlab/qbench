#!/bin/bash

echo "Installing pre-requisites..."
sudo yum -y install tar
sudo yum -y install wget
#sudo easy_install supervisor

echo "Installing JAVA"
wget --no-check-certificate --no-cookies --header "Cookie: oraclelicense=accept-securebackup-cookie" http://download.oracle.com/otn-pub/java/jdk/7u67-b01/jdk-7u67-linux-x64.rpm
sudo rpm -Uvh jdk-7u67-linux-x64.rpm
sudo sh -c 'echo "export JAVA_HOME=/usr/java/default" >> /etc/profile'
sudo sh -c 'echo "export PATH=$PATH:$JAVA_HOME/bin" >> /etc/profile'
sudo sh -c 'echo "export CLASSPATH=.:$JAVA_HOME/jre/lib:$JAVA_HOME/lib:$JAVA_HOME/lib/tools.jar" >> /etc/profile'
source /etc/profile

echo "Downloading Kafka..."
wget http://apache.mirrors.ovh.net/ftp.apache.org/dist/kafka/0.9.0.0/kafka_2.10-0.9.0.0.tgz
tar -zxvf kafka_2.10-0.9.0.0.tgz
rm kafka_2.10-0.9.0.0.tgz
sudo mv kafka_2.10-0.9.0.0/ /opt/kafka
sudo sh -c 'echo "# KAFKA_HOME" >> /etc/profile'
sudo sh -c 'echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile'
sudo sh -c 'echo "export PATH=$PATH:$KAFKA_HOME/bin" >> /etc/profile'
source /etc/profile

#echo "Starting Kafka server..."
#sudo
#sudo /opt/kafka/bin/zookeeper-server-start.sh /opt/kafka/config/zookeeper.properties &
#sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties &
