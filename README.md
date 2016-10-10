#Magpie - A Yarn on Docker Operation Tool

##Precondition
- No-password login to all the active resource managers.
- Docker container's name must contain the cluster name.

##Configuration
conf/conf.ini
The flowing item need to be configured.
- Python
- Cluster names 
- Resource managers' ip address.
- Shipyard
- Swarm

##Run
./magpie.py -h

## Reference

Docker remote API: https://docs.docker.com/engine/reference/api/docker_remote_api_v1.23/

Shipyard API: http://shipyard-project.com/docs/api/

YARN RESTful API: https://hadoop.apache.org/docs/r2.6.0/hadoop-yarn/hadoop-yarn-site/ResourceManagerRest.html

Swarm API: https://docs.docker.com/swarm/swarm-api/

Docker network plugin: https://github.com/TalkingData/Shrike
