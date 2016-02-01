#!/bin/sh

# setup Kafka on AWS

# Get the ID of an instance by name
function get_instance_id () {
  local instance_name=$1
  local instance_id=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${instance_name} | jq '.Reservations[0].Instances[0].InstanceId' | sed -e 's/^"//'  -e 's/"$//')
  echo "${instance_id}"
}

AWS_VPC_ID="vpc-be157ddb"
subnet_id="subnet-a9a28ade"
group_name="kafka-cluster"
group_id=$(aws ec2 describe-security-groups --filters Name=group-name,Values=${group_name} | jq '.SecurityGroups[0].GroupId' | sed -e 's/^"//'  -e 's/"$//')
kafka_instance_name="kafka-node"
kafka_instance_id=$(get_instance_id ${kafka_instance_name})
zk_instance_name="zookeeper"
instance_type="m3.2xlarge"

# Add an InBound rule to a security group
function add_sg_rule() {
  local port=$1
  local cidr=$2
  aws ec2 authorize-security-group-ingress --group-id ${group_id} --protocol tcp --port $port --cidr $cidr 
}

# launch an instance and return its meta-data
function launch_instances() {
  # get parameters
  local instance_type=$1
  local instance_name=$2
  # lanunch the instance
  meta_data=$(aws ec2 run-instances --image-id ami-33734044 --count 1 --instance-type ${instance_type} --key-name data-pipeline-dev --security-group-ids ${group_id} --subnet-id ${subnet_id} --associate-public-ip-address)
  instance_id=$(echo ${meta_data} | jq '.["Instances"][0].InstanceId' | sed -e 's/^"//'  -e 's/"$//')
  aws ec2 create-tags --resources ${instance_id} --tags Key=Name,Value=${instance_name}
  public_dns_name=$(echo ${meta_data} | jq '.Instances[0].PublicDnsName')
}

start() {
  # setup security group
  echo "Creating security group '${group_name}'"
  group_id=$(aws ec2 create-security-group --group-name ${group_name} --vpc-id ${AWS_VPC_ID} --description "A Security Group for Kafka cluster" \
    | jq '.GroupId' \
    | sed -e 's/^"//'  -e 's/"$//'
  )
  add_sg_rule 22   0.0.0.0/0
  add_sg_rule 9092 0.0.0.0/0
  add_sg_rule 2181 0.0.0.0/0
  # launch the instance
  echo "Launching instance within security group '${group_id}'"
  launch_instances "m3.large" ${zk_instance_name}
  launch_instances "m3.2xlarge" ${kafka_instance_name}
  #echo "Instance ${instance_id} available on ${public_dns_name}"
}

teardown() {
  echo "Terminating instance ${instance_id}"
  aws ec2 terminate-instances --instance-ids ${instance_id}
  echo "Deleting security group ${security_group}"
  aws ec2 delete-security-group --group-id ${group_id} 
}

case $1 in
  up)
    start
    ;;
  down)
    teardown
    ;;
  *)
    echo "Unknown command $1, valid ones are 'up', 'down'"
    exit 1
    ;;
esac
