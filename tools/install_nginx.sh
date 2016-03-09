#!/bin/bash
echo "Configuring machine to run nginx"

# increase open files limit
echo "fs.file-max = 1073741824" >> /etc/sysctl.conf
echo "nginx       soft    nofile   40960" >> /etc/security/limits.conf
echo "nginx       hard    nofile   81920" >> /etc/security/limits.conf

# increase nginx limits
sed -i 's/worker_connections *[0-9]*/worker_connections  8192/g' /etc/nginx/nginx.conf
echo "worker_rlimit_nofile 65536;" >> /etc/nginx/nginx.conf

# turn off logging
sed -i 's/access_log .*/access_log off;/g' /etc/nginx/nginx.conf
sed -i 's/error_log .*/error_log off;/g' /etc/nginx/nginx.conf

# reload system config
sysctl -p

# edit the index.html file
rm /usr/share/nginx/html/index.html
echo "0" >> /usr/share/nginx/html/index.html

