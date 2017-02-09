package utils

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/exec"
)

// Convert yarn clustername to IP address
func Clustername2ip(clustername string) string {
	ip := viper.GetString("resource_managers." + clustername)
	return ip
}

//Execute the command
func Run_command(command string) {
	cmd := exec.Command("/bin/bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) == "" {
		return
	}
	fmt.Println(string(out))
}

//Get a logger
func Logger() *logrus.Logger {
	logfile := "magpie.log"
	var logger = logrus.New()
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_RDWR, 0666)
	logger.Out = f
	if err != nil {
		os.Create(logfile)
		fmt.Println("Logfile", logfile, "not exist,created automatically.")
	}
	return logger
}
