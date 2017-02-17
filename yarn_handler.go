package yarn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-simplejson"
	"github.com/rootsongjc/magpie/docker"
	"github.com/rootsongjc/magpie/utils"
	"github.com/samalba/dockerclient"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type YarnStatus struct {
	appsPending           string
	reservedVirtualCores  string
	availableVirtualCores string
	allocatedVirtualCores string
	totalVirtualCores     string
	lostNodes             int64
	activeNodes           int64
	appsRunning           string
	appsFailed            string
	appsKilled            string
	availableMB           string
	allocatedMB           string
	containersPending     string
	totalMB               string
	totalNodes            int64
	rebootedNodes         string
	appsSubmitted         string
	appsCompleted         string
	containersAllocated   string
	reservedMB            string
	containersReserved    string
	unhealthyNodes        int64
	decommissionedNodes   int64
}

//Get yarn cluster node status
func Get_yarn_status(cluster_names []string) {
	fmt.Println("======================YARN CLSUTER STATUS===========================")
	fmt.Println("CLUSTER\tTOTAL\tACTIVE\tDECOM\tLOST\tUNHEALTHY\tUsed")
	var total_nodes, total_active, total_decom, total_lost, total_unhealthy int64
	for i := range cluster_names {
		name := cluster_names[i]
		url := "http://" + utils.Clustername2ip(name) + ":8088/ws/v1/cluster/metrics"
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Yarn cluster ", name, " not found.")
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		js, err := simplejson.NewJson(body)
		if err != nil {
			panic(err.Error())
		}
		nodes, _ := js.Get("clusterMetrics").Map()
		activeNodes, _ := nodes["activeNodes"].(json.Number).Int64()
		totalNodes, _ := nodes["totalNodes"].(json.Number).Int64()
		decommissionedNodes, _ := nodes["decommissionedNodes"].(json.Number).Int64()
		lostNodes, _ := nodes["lostNodes"].(json.Number).Int64()
		unhealthyNodes, _ := nodes["unhealthyNodes"].(json.Number).Int64()
		usage := get_yarn_resource_usage(name)
		total_active += activeNodes
		total_decom += decommissionedNodes
		total_nodes += totalNodes
		total_lost += lostNodes
		total_unhealthy += unhealthyNodes
		fmt.Println(name, "\t", totalNodes, "\t", activeNodes, "\t", decommissionedNodes,
			"\t", lostNodes, "\t", unhealthyNodes, "\t", usage)
	}
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("TOTAL", "\t", total_nodes, "\t", total_active, "\t", total_decom, "\t", total_lost, "\t", total_unhealthy)
}

//Get yarn cluster resource usage percent
func get_yarn_resource_usage(clustername string) float64 {
	url := "http://" + utils.Clustername2ip(clustername) + ":8088/ws/v1/cluster/scheduler"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Yarn cluster ", clustername, " not found.")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	js, err := simplejson.NewJson(body)
	if err != nil {
		panic(err.Error())
	}
	used, _ := js.Get("scheduler").Get("schedulerInfo").Get("usedCapacity").Float64()
	return used
}

//Show the yarn nodemanagers distribution.
func Yarn_distribution(clustername string) {
	fmt.Println("====================RUNNING DOCKERS DISTRIBUTION====================")
	fmt.Println("HOSTNAME\tNUM")
	containers, err := docker.Get_running_docker_containers()
	if err != nil {
		panic(err)
	} else {
		distribution := make(map[string]int, len(containers))
		for i := range containers {
			c := docker.Get_container_name(containers[i].Names[0])
			if strings.HasPrefix(c, clustername) {
				h := docker.Get_nodemanager_host(containers[i].Names[0])
				distribution[h] += 1
			}
		}
		// sorted as the hostname
		sorted_keys := make([]string, 0)
		for k, _ := range distribution {
			sorted_keys = append(sorted_keys, k)
		}

		sort.Strings(sorted_keys)

		for _, k := range sorted_keys {
			fmt.Println(k, "\t", distribution[k])
		}
	}

}

//Inspect the yarn cluster container view
func Yarn_view(clustername string) {
	containers := docker.Get_all_yarn_containers()
	flag := false
	fmt.Println("ID\tCLUSTER\tNAME\tSTATUS\tSTATE\tIP\tHOST")
	fmt.Println("=======================================================================================================================================")
	for i := range containers {
		c := containers[i]
		if c.Clustername == clustername {
			flag = true
			fmt.Println(c.ID, "\t", c.Clustername, "\t", c.Name, "\t", c.Status, "\t", c.State, "\t", c.Ip, "\t", c.Host)
		}
	}
	if flag == false {
		fmt.Println("The cluster does not exited.")
	}

}

//Decommising the nodemanagers on each resourcemanagers.
//No matter which yarn clusters the nodemanager belonging to,
//you can put the in the same nodefile togethter.
//Magpie will recognize the yarn cluster automatically.
func Decommis_nodemanagers(nodemanagers []string) {
	nms := make([]docker.Yarn_docker_container, 0)
	containers := docker.Get_all_yarn_containers()
	for _, n := range nodemanagers {
		found := false
		for _, c := range containers {
			cid := c.ID
			if cid == n {
				found = true
				nms = append(nms, c)
			}
		}
		if found == false {
			fmt.Println("The nodemanager can not be found:", n)
		}
	}

	yarns := make(map[string]string, len(nodemanagers))
	for _, n := range nms {
		if len(yarns) == 0 {
			yarns[n.Clustername] = n.ID
		} else {
			yarns[n.Clustername] = yarns[n.Clustername] + "\n" + n.ID
		}
	}
	//	yarns := map[string]string{
	//		"yarn1": "3cd9494cdc80\n0ebc7d9cf054",
	//		"yarn2": "falj2ljfao3k",
	//	}
	fmt.Println("Decommising the following nodemangers...")
	var wg sync.WaitGroup
	//traverse all the yarn clusters
	for k, v := range yarns {
		wg.Add(1)
		go decommis_yarn_nodes(k, v, &wg)
	}
	wg.Wait()

}

// Decommising the nodemanager of the nodefile
func Decommis_nodemanagers_through_file(nodefile string) {
	fi, err := os.Open(nodefile)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	buff := bufio.NewReader(fi)
	nodes := make([]string, 0)
	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		node := strings.Trim(line, "\n")
		nodes = append(nodes, node)
	}
	Decommis_nodemanagers(nodes)
}

//Decommising the nodemanagers of a yarn
func decommis_yarn_nodes(clustername string, nodemanagers string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	logger := utils.Logger()
	fmt.Println(clustername, nodemanagers)
	logger.WithFields(logrus.Fields{"Time": time.Now(), "Cluster": clustername, "Nodemanagers": strings.Replace(nodemanagers, "\n", ",", -1), "Action": "DECOM"}).Info("Decomissing nodemanagers " + nodemanagers)
	resource_manager_ip := utils.Clustername2ip(clustername)
	nodemanger_exclude_file := viper.GetString("clusters.nodemanager_exclude_file")
	command := "ssh -n root@" + resource_manager_ip + ` "echo -e '` + nodemanagers + `'>>` + nodemanger_exclude_file + `"`
	fmt.Println(command)
	//command := "ssh -n jingchao.song@" + "172.20.0.6" + ` "echo -e '` + nodemanagers + `'>>` + "/home/jingchao.song/docker/test.txt"+ `"`
	//TODO should return result and error handler
	utils.Run_command(command)
	command = "ssh -n root@" + resource_manager_ip + ` 'su - hadoop -c "yarn rmadmin -refreshNodes"'`
	fmt.Println(command)
	utils.Run_command(command)

}

//Offline the host, decommsing the nodemanagers and then delete the docker contianers.
func Offline_host(hostname string) {
	fmt.Println("Offline host", hostname, "...")
	containers := docker.Get_all_yarn_containers()
	nms := make([]string, 0)
	for _, c := range containers {
		host := c.Host
		if host == hostname {
			nms = append(nms, c.Name)
		}
	}
	Decommis_nodemanagers(nms)
	docker.Delete_containers_on_host(hostname)
}

//Create a new nodemanager and start it
func Create_new_nodemanager(nm_config Nodemanager_config) {
	swarm_master_ip := viper.GetString("clusters.swarm_master_ip")
	swarm_master_port := viper.GetString("clusters.swarm_master_port")
	endpoint := "tcp://" + swarm_master_ip + ":" + swarm_master_port
	client, err := dockerclient.NewDockerClient(endpoint, nil)
	logger := utils.Logger()
	if err != nil {
		panic(err)
	}
	if err != nil {
		fmt.Println("Cannot connect to the swarm master.")
	}
	fmt.Println("Creating new nodemanager container...")
	env := []string{
		"HA=" + get_nodemanager_config(nm_config.HA, "HA"),
		"NAMESERVICE=" + get_nodemanager_config(nm_config.NAMESERVICE, "NAMESERVICE"),
		"ACTIVE_NAMENODE_IP=" + get_nodemanager_config(nm_config.ACTIVE_NAMENODE_IP, "ACTIVE_NAMENODE_IP"),
		"STANDBY_NAMENODE_IP=" + get_nodemanager_config(nm_config.STANDBY_NAMENODE_IP, "STANDBY_NAMENODE_IP"),
		"ACTIVE_NAMENODE_ID=" + get_nodemanager_config(nm_config.ACTIVE_NAMENODE_ID, "ACTIVE_NAMENODE_ID"),
		"STANDBY_NAMENODE_ID=" + get_nodemanager_config(nm_config.STANDBY_NAMENODE_ID, "STANDBY_NAMENODE_ID"),
		"HA_ZOOKEEPER_QUORUM=" + get_nodemanager_config(nm_config.HA_ZOOKEEPER_QUORUM, "HA_ZOOKEEPER_QUORUM"),
		"NAMENODE_IP=" + get_nodemanager_config(nm_config.NAMENODE_IP, "NAMENODE_IP"),
		"RESOURCEMANAGER_IP=" + get_nodemanager_config(nm_config.RESOURCEMANAGER_IP, "RESOURCEMANAGER_IP"),
		"YARN_RM1_IP=" + nm_config.YARN_RM1_IP,
		"YARN_RM2_IP=" + nm_config.YARN_RM2_IP,
		"YARN_JOBHISTORY_IP=" + nm_config.YARN_JOBHISTORY_IP,
		"CPU_CORE_NUM=" + get_nodemanager_config(nm_config.CPU_CORE_NUM, "CPU_CORE_NUM"),
		"NODEMANAGER_MEMORY_MB=" + get_nodemanager_config(nm_config.NODEMANAGER_MEMORY_MB, "NODEMANAGER_MEMORY_MB"),
		"YARN_CLUSTER_ID=" + nm_config.YARN_CLUSTER_ID,
		"YARN_ZK_DIR=" + nm_config.YARN_ZK_DIR,
		//"PATH=/usr/local/hadoop/bin:/usr/local/hadoop/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/java/bin://usr/local/java/jre/bin",
	}
	hostConifg := dockerclient.HostConfig{
		CpuShares:   nm_config.Limit_cpus,
		Memory:      nm_config.Limit_memory_mb * 1024 * 1024, //transform to Byte
		NetworkMode: nm_config.Network_mode,
	}
	var config *dockerclient.ContainerConfig
	config = new(dockerclient.ContainerConfig)
	config.Image = get_nodemanager_config(nm_config.Image, "image")
	config.Env = env

	//inherit Cmd and Entrypoint settings from docker Image or set them on config file
	config.Cmd = viper.GetStringSlice("nodemanager.cmd")
	config.Entrypoint = viper.GetStringSlice("nodemanager.entrypoint")
	config.HostConfig = hostConifg
	id, err := client.CreateContainer(config, "", nil)
	if err != nil {
		logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": "-", "Action": "CREATE"}).Error(err)
		panic(err)
	}
	container_name := id[0:12]
	fmt.Println("Container", container_name, "created.")
	logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": container_name, "Action": "CREATE"}).Info("Create a new nodemanager docker container")

	if nm_config.Container_name != "" {
		err = client.RenameContainer(container_name, nm_config.Container_name)
		if err != nil {
			logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": container_name, "Action": "RENAME"}).Error(err)
			panic(err)
		}
		fmt.Println("Rename container name to", nm_config.Container_name)
		logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": container_name, "Action": "RENAME"}).Info("Rename container "+container_name+" name to ", nm_config.Container_name)
	}
	err = client.StartContainer(id, nil)
	if err != nil {
		logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": container_name, "Action": "START"}).Error(err)
		panic(err)
	}
	fmt.Println("Started.")
	logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": container_name, "Action": "START"}).Info("Start container " + container_name)

}

//Get nodemanger configuration item from config file
//If no config specify throught the command line,use the default settings
func get_nodemanager_config(nm string, config string) string {
	if nm == "" {
		return viper.GetString("nodemanager." + config)
	}
	return nm
}
