deps:
	go get 

build:
	go build 

infra:  
	ansible-playbook tools/site.yml -v --private-key=~/.ssh/data-pipeline-dev.pem -i tools/hosts
