#Docker Project Standard#

All our project follow the same organisation and mechanism to build and deploy images.



##Project Tree##

| Name                 | Comment                                  |
| -------------------- | ---------------------------------------- |
| Dockerfile           | Docker image cook recipes                |
| Makefile             | Common target to speed up repetitive tasks |
| LICENSE              | License content                          |
| README.md            | A full description of current project (purpose, feature, change log) |
| README-shotr.md      | A short description of current project.  |
| .dockerignore        | A list of excluded resource (not visible by docker deamon to build) |
| .gitignore           | A list of excluded resource for git      |
| etc/services-config/ | A supervisor program's configiration file, .conf or .ini |
| usr/sbin/            | A supervisor program's bootstrap script. |



##Project file tree example##

```
.
|-- Dockerfile
|-- etc
|   `-- services-config
|       `-- hadoop-bootstrap.conf
|-- LICENSE
|-- Makefile
|-- README.md
|-- README-short.md
`-- usr
    `-- sbin
            `-- hadoop-bootstrap
```

**hadoop-bootstrap.conf**

```ini
[program:hadoop-bootstrap]
priority = 5
command = /usr/sbin/hadoop-bootstrap
autostart = true
user = hadoop
startsecs = 0
startretries = 0
autorestart = false
redirect_stderr = true
stdout_logfile = /var/log/hadoop-bootstrap.log
stdout_events_enabled = true
hadoop-bootstap
#!/bin/bash
#Init the hadoop configuration
#Author:jimmysong
#Date:2016-11-09
HADOOP_PREFIX=/usr/local/hadoop 
#Start service
$HADOOP_PREFIX/etc/hadoop/hadoop-env.sh 
$HADOOP_PREFIX/sbin/hadoop-daemon.sh start namenode
$HADOOP_PREFIX/sbin/hadoop-daemon.sh start datanode
$HADOOP_PREFIX/bin/hdfs dfs -mkdir -p /user/hadoop
$HADOOP_PREFIX/bin/hdfs dfs -chown hadoop:hadoop /user/hadoop
$HADOOP_PREFIX/sbin/yarn-daemon.sh start resourcemanager
$HADOOP_PREFIX/sbin/yarn-daemon.sh start nodemanager
```



##Make##

All our Makefile, offer the same target. Install "make" utility, and execute: make build to build current project.

In Makefile, you could retrieve this variables:

- NAME: declare a full image name (aka airdock/base, airdoc/oracle-jdk, ...)
- VERSION: declare image version

and tasks:         
- all: alias to 'build'
- clean: remove all container which depends on this image, and remove image previously builded
- build: clean and build the current version
- tag_latest: tag current version with ":latest"
- release: build and execute tag_latest, push image onto registry, and tag git repository
- debug: launch default command with builded image in interactive mode
- run: run image as daemon with common port exposition (if relevant) and print IP address`
