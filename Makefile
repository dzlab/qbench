deps:
	go get 

build:
	go build 

dockerize:
	export PWD=`pwd`
	docker run --rm -v "$(PWD):/src" -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
	docker build -t qbench .
	docker tag -f qbench dzlabs/qbench
	docker push dzlabs/qbench:latest
# to run: docker run -it --rm --name qbech-running qbench qbench -t 1000 -p 8080 -s 1000 http -u http://hb-endpoint-elb-1722231236.eu-west-1.elb.amazonaws.com -m GET

ecs-configure:
	ecs-cli configure --cluster qbench-cluster --region eu-west-1
	
ecs-run:	
	ecs-cli up --keypair data-pipeline-dev --capability-iam --size 2 --instance-type t2.medium
	ecs-cli compose --file docker-compose.yml up

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
	ansible-playbook tools/kafka.yml -v -i tools/hosts
# create an elb with nginx 
nginx-cluster:
	ansible-playbook tools/nginx.yml -v -i tools/hosts

beanstalk-create:
	yes firehose-dev | eb create
beanstalk-terminate:
	yes firehose-dev | eb terminate
