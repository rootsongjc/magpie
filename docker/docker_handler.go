package docker

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/rootsongjc/magpie/utils"
	"github.com/samalba/dockerclient"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

//Yarn docker cluster state
type docker_cluster_state struct {
	cluster string
	total   int
	running int
	exited  int
}

var logger = utils.Logger()

//Yarn nodemanager docker container
type Yarn_docker_container struct {
	ID          string
	Clustername string //yarn1
	Name        string //yarn1-20160912-nm1
	Status      string //exited
	State       string //Exited (0) 3 hours ago
	Ip          string //172.18.12.31
	Host        string //bj-dc-datanode-078.tendcloud.com
}

//Get the yarn docker cluster status
func Get_docker_status(cluster_names []string) {
	fmt.Println("======================DOCKER CLSUTER STATUS=========================")
	fmt.Println("CLUSTER\tTOTAL\tRUNNING\tEXITED")
	yarn_containers := Get_all_yarn_containers()
	var total, running, exited int
	exited_containers := make([]Yarn_docker_container, 0)
	arr := make([]docker_cluster_state, len(cluster_names))
	for j := range yarn_containers {
		c := yarn_containers[j]
		for i := range cluster_names {
			arr[i].cluster = cluster_names[i]
			// name format: yarn6-20160901-nm139
			if c.Clustername == cluster_names[i] {
				arr[i].total += 1
				if c.State == "running" {
					arr[i].running += 1
				} else if c.State == "exited" {
					arr[i].exited += 1
					exited_containers = append(exited_containers, c)
				}
			}
		}
	}
	for i := range cluster_names {
		fmt.Println(arr[i].cluster, "\t", arr[i].total, "\t", arr[i].running, "\t", arr[i].exited)
		total += arr[i].total
		running += arr[i].running
		exited += arr[i].exited
	}
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("TOTAL", "\t", total, "\t", running, "\t", exited)
	fmt.Println("================EXITED DOCKER CONTAINERS DISTRIBUTION===============")
	if len(exited_containers) == 0 {
		fmt.Println("None")
	} else {
		fmt.Println("CLUSTER\tNAME\tSTATUS\tHOSTNAME")
		for i := range exited_containers {
			c := exited_containers[i]
			fmt.Println(c.Clustername, "\t", c.Name, "\t", c.Status, "\t", c.Host)
		}
	}
}

//Delete a docker container.
func Delete_container(id string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	client, err := Swarm_client()
	if err != nil {
		panic(err)
	}
	fmt.Println("Removal", id, "in progress...")
	err = client.RemoveContainer(dc.RemoveContainerOptions{ID: id, Force: true})
	if err != nil {
		logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": id, "Action": "DELETE"}).Error(err)
		panic(err)
	}
	fmt.Println("Remove", id, "OK")
	logger.WithFields(logrus.Fields{"Time": time.Now(), "ContainerID": id, "Action": "DELETE"}).Info("Delete container " + id)
}

//Delete all the docker containers on the host.
func Delete_containers_on_host(hostname string) {
	containers, err := Get_all_docker_containers()
	if err != nil {
		panic(err)
	} else {
		var wg sync.WaitGroup
		for i := range containers {
			c := Get_nodemanager_host(containers[i].Names[0])
			if c == hostname {
				id := containers[i].ID[0:12]
				fmt.Println("Delete docker contianer ID:", id, " NAME:", containers[i].Names[0])
				wg.Add(1)
				go Delete_container(id, &wg)
			}
		}
		wg.Wait()
	}
}

//Delete the containers of the file list
func Delete_container_file_list(path string) {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	containers := Get_all_yarn_containers()
	buff := bufio.NewReader(fi)
	var wg sync.WaitGroup
	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		name := strings.Trim(line, "\n")
		wg.Add(1)
		for i := range containers {
			if containers[i].Name == name {
				fmt.Println("Delete docker container:", name, " ID:", containers[i].ID)
				go Delete_container(containers[i].ID, &wg)

			}
		}
	}
	wg.Wait()

}

//Get container name for example: yarn1-20141231-nm1
func Get_container_name(longname string) string {
	return strings.Split(longname, "/")[2]
}

//Get nodemanager local machine's hostname for example:bj-dc-datanode-006.tendcloud.com
func Get_nodemanager_host(longname string) string {
	return strings.Split(longname, "/")[1]
}

//Convert the swarm docker struct to yarn_docker_container struct
func Convert_yarn_docker_container(docker_container dc.APIContainers) Yarn_docker_container {
	id := docker_container.ID[0:12]
	name := strings.Split(docker_container.Names[0], "/")[2]
	clustername := strings.Split(name, "-")[0]
	status := docker_container.Status
	state := docker_container.State
	//Exited container has no IPAddress
	var ip = "-"
	if state == "running" {
		ip = docker_container.Networks.Networks["mynet"].IPAddress
	}
	host := strings.Split(docker_container.Names[0], "/")[1]

	return Yarn_docker_container{ID: id,
		Name:        name,
		Clustername: clustername,
		Status:      status,
		State:       state,
		Ip:          ip,
		Host:        host}
}

//Get yarn docker containers list
func Get_all_yarn_containers() []Yarn_docker_container {
	cluster_names := viper.GetStringSlice("clusters.cluster_name")
	list := make([]Yarn_docker_container, 0)
	containers, err := Get_all_docker_containers()
	if err != nil {
		fmt.Println("Cann connet to the swarm master.")
		panic(err)
	}
	for j := range containers {
		container := containers[j]
		// Names format: [/bj-yh-dc-datanode-141.tendcloud.com/yarn6-20160901-nm141]
		c := Get_container_name(container.Names[0])
		for i := range cluster_names {
			// name format: yarn6-20160901-nm139
			name := cluster_names[i]
			if strings.HasPrefix(c, name) {
				ydc := Convert_yarn_docker_container(container)
				list = append(list, ydc)
			}
		}
	}
	return list
}

//Get all the docker containers of swarm cluster
func Get_all_docker_containers() ([]dc.APIContainers, error) {
	client, err := Swarm_client()
	if err != nil {
		fmt.Println("Cann connet to the swarm master.")
		panic(err)
	}
	containers, err := client.ListContainers(dc.ListContainersOptions{All: true})
	return containers, err
}

//Get all the running docker containers of swarm cluster
func Get_running_docker_containers() ([]dc.APIContainers, error) {
	client, err := Swarm_client()
	if err != nil {
		fmt.Println("Cann connet to the swarm master.")
		panic(err)
	}
	containers, err := client.ListContainers(dc.ListContainersOptions{})
	return containers, err
}

//Make a swarm client
func Swarm_client() (*dc.Client, error) {
	swarm_master_ip := viper.GetString("clusters.swarm_master_ip")
	swarm_master_port := viper.GetString("clusters.swarm_master_port")
	endpoint := "tcp://" + swarm_master_ip + ":" + swarm_master_port
	client, err := dc.NewClient(endpoint)
	return client, err
}

//Parse swarm cluster nodes info
func ParseClusterNodes() ([]Node, error) {
	swarm_master_ip := viper.GetString("clusters.swarm_master_ip")
	swarm_master_port := viper.GetString("clusters.swarm_master_port")
	endpoint := "tcp://" + swarm_master_ip + ":" + swarm_master_port
	client, err := dockerclient.NewDockerClient(endpoint, nil)
	if err != nil {
		panic(err)
	}
	info, err := client.Info()
	if err != nil {
		return nil, err
	}
	driverStatus := info.DriverStatus
	nodes := []Node{}
	var node Node
	nodeComplete := false
	name := ""
	addr := ""
	containers := ""
	reservedCPUs := ""
	reservedMemory := ""
	labels := []string{}
	for _, l := range driverStatus {
		if len(l) != 2 {
			continue
		}
		label := l[0]
		data := l[1]

		// cluster info label i.e. "Filters" or "Strategy"
		if strings.Index(label, "\u0008") > -1 {
			continue
		}

		if strings.Index(label, " └") == -1 {
			name = label
			addr = data
		}

		// node info like "Containers"
		switch label {
		case " └ Containers":
			containers = data
		case " └ Reserved CPUs":
			reservedCPUs = data
		case " └ Reserved Memory":
			reservedMemory = data
		case " └ Labels":
			lbls := strings.Split(data, ",")
			labels = lbls
			nodeComplete = true
		default:
			continue
		}

		if nodeComplete {
			node = Node{
				Name:           name,
				Addr:           addr,
				Containers:     containers,
				ReservedCPUs:   reservedCPUs,
				ReservedMemory: reservedMemory,
				Labels:         labels,
			}
			nodes = append(nodes, node)

			// reset info
			name = ""
			addr = ""
			containers = ""
			reservedCPUs = ""
			reservedMemory = ""
			labels = []string{}
			nodeComplete = false
		}
	}
	return nodes, nil
}

//Show the swarm cluster status
func Get_swarm_nodes_status() {
	nodes, err := ParseClusterNodes()
	if err != nil {
		panic(err)
	}
	fmt.Println("Name\tAddr\tContainers\tReservererCPUs\tReserverdMemeory")
	fmt.Println("============================================================================================================")
	for _, n := range nodes {
		fmt.Println(n.Name, "\t", n.Addr, "\t", n.Containers, "\t", n.ReservedCPUs, "\t", n.ReservedMemory)
	}
}

//Lookup the IP address of hostname
//Hostname is the 12 byte short container ID.
func Lookup(hostname string, all bool) {
	list := Get_all_yarn_containers()
	var container Yarn_docker_container
	flag := false
	for i := range list {
		c := list[i]
		if c.ID == hostname {
			flag = true
			container = c
		}
	}
	if flag == false {
		fmt.Println("No such contianer.")
	} else if all == true {
		fmt.Println("ID:", container.ID)
		fmt.Println("CLUSTER:", container.Clustername)
		fmt.Println("NAME:", container.Name)
		fmt.Println("STATUS:", container.Status)
		fmt.Println("STATE:", container.State)
		fmt.Println("IP:", container.Ip)
		fmt.Println("HOST", container.Host)
	} else {
		fmt.Println("IP:", container.Ip)
	}
}
