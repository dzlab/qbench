# To build: docker build -t qbench --build-arg TOTAL=1000,URL=http://www.example.com/ .
# To   run: docker run -it --rm --name qbench-running qbench
FROM busybox
MAINTAINER dzlabs <dzlabs@outlook.com>
#FROM golang

#ENV WORKDIR /go/src/github.com/dzlab/qbench
#ENV PATH $PATH:$WORKDIR

#COPY . $WORKDIR
#RUN cd $WORKDIR && go get -d -v
#RUN cd $WORKDIR && go install -v

#CMD $WORKDIR/qbench -t ${TOTAL} -p 8080 -s 1000 http -u ${URL} -m GET

VOLUME ["/tmp"]
COPY qbench /
#ENTRYPOINT ["/qbench"]
