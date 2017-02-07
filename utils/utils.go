package utils

import (
	"fmt"
	docker "github.com/rootsongjc/magpie/docker"
	"github.com/spf13/viper"
	"os/exec"
)

// Convert yarn clustername to IP address
func Clustername2ip(clustername string) string {
	ip := viper.GetString("resource_managers." + clustername)
	return ip
}

//Lookup the IP address of hostname
//Hostname is the 12 byte short container ID.
func Lookup(hostname string, all bool) {
	list := docker.Get_all_yarn_containers()
	var container docker.Yarn_docker_container
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


//Execute the command
func Run_command(command string){
	cmd := exec.Command("/bin/bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	if string(out)==""{
		return
	}
	fmt.Println(string(out))
}
