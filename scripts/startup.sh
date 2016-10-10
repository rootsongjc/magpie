#!/bin/bash
#Init the hadoop configuration
#Author:jingchao.song@tendcloud.com
#Date:2016-06-12

srv=$1
#Modify ulimits
echo "* soft nofile 655350" >> /etc/security/limits.conf
echo "* hard nofile 655350" >> /etc/security/limits.conf
echo "@hadoop        hard    nproc           655350" >> /etc/security/limits.conf
echo "@hadoop        soft    nproc           655350" >> /etc/security/limits.conf
echo "@root        soft    nproc           655350" >> /etc/security/limits.conf
echo "@root        hard    nproc           655350" >> /etc/security/limits.conf
echo "ulimit -SH 655350" >> /etc/rc.local

#Export LANG
echo 'export LANG="en_US.UTF-8"'>>/etc/profile
#Edit hadoop configuration
if [ $HA = "yes" ]; then
echo "With hadoop HA"
mv $HADOOP_HOME/etc/hadoop/withha/* $HADOOP_HOME/etc/hadoop/
sed -i -E "s/NAMESERVICE/$NAMESERVICE/g" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i -E "s/NAMESERVICE/$NAMESERVICE/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/ACTIVE_NAMENODE_IP/$ACTIVE_NAMENODE_IP/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/ACTIVE_NAMENODE_ID/$ACTIVE_NAMENODE_ID/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/STANDBY_NAMENODE_IP/$STANDBY_NAMENODE_IP/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/STANDBY_NAMENODE_ID/$STANDBY_NAMENODE_ID/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/HA_ZOOKEEPER_QUORUM/$HA_ZOOKEEPER_QUORUM/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/HA_ZOOKEEPER_QUORUM/$HA_ZOOKEEPER_QUORUM/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_ZK_DIR/$YARN_ZK_DIR/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_CLUSTER_ID/$YARN_CLUSTER_ID/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_RM1_IP/$YARN_RM1_IP/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_RM2_IP/$YARN_RM2_IP/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_JOBHISTORY_IP/$YARN_JOBHISTORY_IP/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/YARN_JOBHISTORY_IP/$YARN_JOBHISTORY_IP/g" $HADOOP_HOME/etc/hadoop/mapred-site.xml
elif [ $HA = "no" ]; then
echo "Without hadoop HA"
mv $HADOOP_HOME/etc/hadoop/withoutha/* $HADOOP_HOME/etc/hadoop/
sed -i -E "s/NAMENODE_IP/$NAMENODE_IP/g" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i -E "s/NAMENODE_IP/$NAMENODE_IP/g" $HADOOP_HOME/etc/hadoop/hdfs-site.xml
sed -i -E "s/RESOURCEMANAGER_IP/$RESOURCEMANAGER_IP/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/HISTORYSERVER_IP/$HISTORYSERVER_IP/g" $HADOOP_HOME/etc/hadoop/mapred-site.xml
fi
#nodemanager resource limit
sed -i -E "s/CPU_CORE_NUM/$CPU_CORE_NUM/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml
sed -i -E "s/NODEMANAGER_MEMORY_MB/$NODEMANAGER_MEMORY_MB/g" $HADOOP_HOME/etc/hadoop/yarn-site.xml

#Start yarn resourcemanager
if [ $srv = "resourcemanager" ]; then
su - hadoop -c "yarn-daemon.sh start resourcemanager"
su - hadoop -c "mr-jobhistory-daemon.sh  start historyserver"
elif [ $srv = "nodemanager" ]; then
su - hadoop -c "yarn-daemon.sh start nodemanager"
else 
echo "No such arguments. Plsease use resourcemanager or nodemanager."
fi

#Foreground
echo "Press Ctrl+P and Ctrl+Q to background this process."
echo 'Use exec command to open a new bash instance for this instance (Eg. "docker exec -i -t CONTAINER_ID bash"). Container ID can be obtained using "docker ps" command.'
echo "Start Terminal"
bash
echo "Press Ctrl+C to stop instance."
sleep infinity
