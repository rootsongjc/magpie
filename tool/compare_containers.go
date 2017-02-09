package tool

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/rootsongjc/magpie/utils"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func Compare_yarn_docker_cluster(clustername string) {
	endpoint := "tcp://" + viper.GetString("clusters.swarm_master_ip") + ":" + viper.GetString("clusters.swarm_master_port")
	client, err := dc.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	dockercontainers := set.New()
	containers, err := client.ListContainers(dc.ListContainersOptions{})
	for j := range containers {
		// Names format: [/bj-yh-dc-datanode-141.tendcloud.com/yarn6-20160901-nm141]
		c := strings.Split(containers[j].Names[0], "/")[2]
		if strings.HasPrefix(c, clustername) {
			hostname := containers[j].ID
			dockercontainers.Add(hostname[0:12])
		}
	}
	url := "http://" + utils.Clustername2ip(clustername) + ":8088/ws/v1/cluster/nodes"
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
	nodes, _ := js.Get("nodes").Get("node").Array()
	nodemanagers := set.New()
	for j := range nodes {
		state := nodes[j].(map[string]interface{})["state"]
		hostname := nodes[j].(map[string]interface{})["nodeHostName"].(string)
		if state == "RUNNING" {
			nodemanagers.Add(hostname)
		}
	}
	var result string
	result = clustername + "\t" + strconv.Itoa(nodemanagers.Size()) + "\t" + strconv.Itoa(dockercontainers.Size())
	if dockercontainers.Size() > nodemanagers.Size() {
		diff := set.Difference(dockercontainers, nodemanagers)
		result = result + "\t" + diff.String()
	} else if dockercontainers.Size() == nodemanagers.Size() {
		result = result + "\t" + "OK"
	} else if dockercontainers.Size() < nodemanagers.Size() {
		diff := set.Difference(nodemanagers, dockercontainers)
		result = result + "\t" + diff.String()
	}
	fmt.Println(result)

}
