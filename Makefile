deps:
	go get 

build:
	go build 

cross-platform:
	GOOS=linux GOARCH=amd64 go build
#	for GOOS in linux; do
#		for GOARCH in 386 amd64; do
#			echo "Building ${GOOS}-${GOARCH}"
#			go build -o qbench-${GOOS}-${GOARCH}
#		done
#	done

# create a kafka cluster along with zookeeper instances
kafka-cluster:  
	ansible-playbook tools/kafka.yml -v --private-key=~/.ssh/data-pipeline-dev.pem -i tools/hosts

beanstalk-create:
	yes firehose-dev | eb create
beanstalk-terminate:
	yes firehose-dev | eb terminate
