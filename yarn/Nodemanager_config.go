package yarn

type Nodemanager_config struct {
	//The config items with all capital alphabet are docker image ENV
	//The following items have default config on docker image
	HA                  string //default yes
	NAMESERVICE         string
	ACTIVE_NAMENODE_IP  string
	STANDBY_NAMENODE_IP string
	ACTIVE_NAMENODE_ID  string
	STANDBY_NAMENODE_ID string
	HA_ZOOKEEPER_QUORUM string

	//You must set the following items with your own address
	YARN_ZK_DIR        string
	YARN_CLUSTER_ID    string
	YARN_RM1_IP        string
	YARN_RM2_IP        string
	YARN_JOBHISTORY_IP string

	//Without HA config
	NAMENODE_IP        string
	RESOURCEMANAGER_IP string

	CPU_CORE_NUM          string //default 4
	NODEMANAGER_MEMORY_MB string //default 10240
	Network_mode          string //default mynet
	Limit_cpus            int64  //default 5
	Limit_memory_mb       int64  //default 12288

	Image    string //nodemanager docker image
	Container_name string //docker container name
}
