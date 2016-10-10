#Yarn On Docker

Build and run an yarn cluster on docker, pass the config item to hadoop configuration files through docker ENV.

###How the image was build and run?

- Step1

  Prepare the direcotries for yarn and chown owner to hadoop.

  Download hadoop-2.6.0-cdh5.5.2.tar.gz and install it in docker image. Unzip it and remove the default hadoop configuration files.

- Step2

  Put the codec lib.so files into the hadoop native directory.

  Put custom hadoop configuration files to hadoop conf directory.

- Step3

  Set the ENV and Entrypoint. 

- Stop4

  Run a container with sepecific ENV passed to it.



###Build image

Edit Dockerfile and change the base image to your own JDK7 image.

```
./build.sh
docker build -t hadoop-yarn:v0.1 .
```

###Run image

For example

With hadoop ha

```
docker run -d -e NAMESERVICE=addmp -e ACTIVE_NAMENODE_ID=namenode29 -e STANDBY_NAMENODE_ID=namenode63 -e HA_ZOOKEEPER_QUORUM=172.16.20.50:2181,172.16.20.51:2181,172.16.20.52:2181 -e YARN_ZK_DIR=rmstore -e YARN_CLUSTER_ID= yarnRM -e YARN_RM1_IP=172.16.20.52 -e YARN_RM2_IP=172.16.20.51 -e YARN_JOBHISTORY_IP=172.16.20.52 -e ACTIVE_NAMENODE_IP=172.16.20.50 -e STANDBY_NAMENODE_IP=172.16.20.51  -e HA=yes hadoop-yarn:v0.1 resourcemanager

docker run -d -e NAMESERVICE=addmp -e ACTIVE_NAMENODE_ID=namenode29 -e STANDBY_NAMENODE_ID=namenode63 -e HA_ZOOKEEPER_QUORUM=172.16.20.50:2181,172.16.20.51:2181,172.16.    20.52:2181 -e YARN_ZK_DIR=rmstore -e YARN_CLUSTER_ID= yarnRM -e YARN_RM1_IP=172.16.20.52 -e YARN_RM2_IP=172.16.20.51 -e YARN_JOBHISTORY_IP=172.16.20.52 -e ACTIVE_NAMENO    DE_IP=172.16.20.50 -e STANDBY_NAMENODE_IP=172.16.20.51  -e HA=yes hadoop-yarn:v0.1 nodemanager
```

Without hadoop ha

```
docker run -d -e NANENODE_IP=172.16.31.63 -e RESOURCEMANAGER_IP=172.16.31.63 -e YARN_JOBHISTORY_IP=172.16.31.63 -e HA=no hadoop-yarn:v0.1 resourcemanager

docker run -d -e NANENODE_IP=172.16.31.63 -e RESOURCEMANAGER_IP=172.16.31.63 -e YARN_JOBHISTORY_IP=172.16.31.63 -e HA=no hadoop-yarn:v0.1 nodemanager
```

###ENV included with hadoop HA 

- HA (default yes)

- NAMESERVICE

- ACTIVE_NAMENODE_IP

- STANDBY_NAMENODE_IP

- ACTIVE_NAMENODE_ID

- STANDBY_NAMENODE_ID

- HA_ZOOKEEPER_QUORUM

- YARN_ZK_DIR

- YARN_CLUSTER_ID

- YARN_RM1_IP

- YARN_RM2_IP

- YARN_JOBHISTORY_IP

###ENV included without hadoop HA

- NAMENDOE_IP

- RESOURCEMANAGER_IP

- HISTORYSERVER_IP

###NodeManager resource limit

- CPU_CORE_NUM

- NODEMANAGER_MEMORY_MB

##Management Tool Magpie - A Yarn on Docker Operation Tool

##Precondition
- No-password login to all the active resource managers.
- Docker container's name must contain the cluster name.

###Configuration
conf/conf.ini
The flowing item need to be configured.
- Python
- Cluster names
- Resource managers' ip address.
- Shipyard
- Swarm

###Run
./magpie.py -h

###Reference

Docker remote API: https://docs.docker.com/engine/reference/api/docker_remote_api_v1.23/

Shipyard API: http://shipyard-project.com/docs/api/

YARN RESTful API: https://hadoop.apache.org/docs/r2.6.0/hadoop-yarn/hadoop-yarn-site/ResourceManagerRest.html

Swarm API: https://docs.docker.com/swarm/swarm-api/

Docker network plugin: https://github.com/TalkingData/Shrike

###About

Author: rootsongjc@gmail.com

*FYI: If you want to create a yarn cluster with multiple nodemanagers, you need a docker plugins to make the docker container on different hosts can be accessed with each others.*
You need a docker ipam plugin to make the continers located on different hosts can be accessed by each others. 
Try this:https://github.com/rootsongjc/docker-ipam-plugin
You also need a plugin to listen on docker nodes and register container's IP-hostname into a DNS server so that docker containers can recognise each other by the hostname which is the same with the container ID.
