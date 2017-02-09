package docker

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/rootsongjc/magpie/utils"
	"github.com/samalba/dockerclient"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScaleResult struct {
	Scaled []string
	Errors []string
}

//Convert yarn name to short container ID.
func Convert_name2ID(name string) string {
	list := Get_all_yarn_containers()
	var container Yarn_docker_container
	for i := range list {
		c := list[i]
		if c.Name == name {
			container = c
		}
	}
	return container.ID
}

//Scale yarn cluster with the base container
func Scale_yarn_cluster(clustername string, numInstances int) {
	logger := utils.Logger()
	fmt.Println("Scaling yarn cluster", clustername, "with", numInstances, "containers...")
	var (
		errChan = make(chan (error))
		resChan = make(chan (string))
		result  = ScaleResult{Scaled: make([]string, 0), Errors: make([]string, 0)}
		lock    sync.Mutex // when set container affinities to swarm cluster, must use mutex
	)
	swarm_master_ip := viper.GetString("clusters.swarm_master_ip")
	swarm_master_port := viper.GetString("clusters.swarm_master_port")
	endpoint := "tcp://" + swarm_master_ip + ":" + swarm_master_port
	client, err := dockerclient.NewDockerClient(endpoint, nil)
	if err != nil {
		panic(err)
	}
	if err != nil {
		fmt.Println("Cannot connect to the swarm master.")
	}
	name := viper.GetString("base_container." + clustername)
	id := Convert_name2ID(name)
	if id == "" {
		fmt.Println("The base container ", name, " does not exist.")
		os.Exit(1)
	}
	logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": id, "Action": "SCALE"}).Info("Scale yarn clsuter " + clustername + " with base container " + id + " to " + strconv.Itoa(numInstances) + " nodemanagers")
	containerInfo, err := client.InspectContainer(id)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	//from shipyard
	nodes, err := ParseClusterNodes()
	containers := Get_all_yarn_containers()
	var hostname = ""
	var cpuNum int64
	for i := range containers {
		if containers[i].Name == name {
			hostname = containers[i].Host
		}
	}
	//Get the swarm node cpu number, for example 12 or 25
	for _, node := range nodes {
		if node.Name == hostname {
			cpuNum, err = strconv.ParseInt(strings.Split(node.ReservedCPUs, " ")[2], 10, 64)
			if err != nil {
				panic(err)
			}
		}
	}
	for i := 0; i < numInstances; i++ {
		go func(instance int) {
			//log.Debugf("scaling: id=%s #=%d", containerInfo.ID, instance)
			config := containerInfo.Config
			// clear hostname to get a newly generated
			config.Hostname = ""
			hostConfig := *containerInfo.HostConfig
			hostConfig.CpuShares = containerInfo.HostConfig.CpuShares * cpuNum / 1024.0
			config.HostConfig = hostConfig
			lock.Lock()
			defer lock.Unlock()
			id, err := client.CreateContainer(config, "", nil)
			fmt.Println("New container created", id)
			logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": id, "Action": "CREATE"}).Info("Create container " + id)
			if err != nil {
				errChan <- err
				return
			}
			if err := client.StartContainer(id, nil); err != nil {
				errChan <- err
				return
			}
			resChan <- id
		}(i)
	}

	for i := 0; i < numInstances; i++ {
		select {
		case id := <-resChan:
			result.Scaled = append(result.Scaled, id)
		case err := <-errChan:
			result.Errors = append(result.Errors, strings.TrimSpace(err.Error()))
		}
	}
	//Rename new containers
	//New name format like: yarn2-20161017132558-7
	if len(result.Errors) == 0 {
		containers := result.Scaled
		for i, c := range containers {
			oldname := c[0:12]
			var newname string
			newname = clustername + "-" + time.Now().Format("20060102150405") + "-" + strconv.Itoa(i)
			fmt.Println("Rename container", oldname, "as", newname)
			logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": id, "Action": "RENAME"}).Info("Rename container " + oldname + " to " + newname)
			client.RenameContainer(oldname, newname)
		}
	}

}
