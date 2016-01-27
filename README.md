# qbench
This is a simple tool for benchmarking a queue, it tries to send arbitrary bytes and measure the time it takes.
Currently only AWS Firehose is supported, the tool can be run locally but also on AWS Elastic Beanstalk.

Running:
First, install Elatic Beanstalk CLI before running this tool. Then, issue:
```
$ eb init
$ yes qbench | eb create
```
Visualisation:
```
cat duration.conf | gnuplot
```
