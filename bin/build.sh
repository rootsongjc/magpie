#!/bin/bash
#Download hadoop packages

wget http://archive.cloudera.com/cdh5/cdh/5/hadoop-2.6.0-cdh5.5.2.tar.gz ../ && \
cd .. && \
tar xvf hadoop-2.6.0-cdh5.5.2.tar.gz && \
rm -f hadoop-2.6.0-cdh5.5.2.tar.gz && \
rm -rf hadoop-2.6.0-cdh5.5.2/etc/hadoop 
