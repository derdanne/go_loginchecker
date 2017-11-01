package main

import (
	"log"
	"os/exec"
	"strings"
)

type Host struct {
	Hostname string
}

func detectHostname() *Host {
	hostnameCmd, lookErr := exec.LookPath("hostname")
	args := []string{"-f"}
	if lookErr != nil {
		log.Print(lookErr)
		hostnameCmd, lookErr = exec.LookPath("uname")
		args = []string{"-n"}
	}

	hostname, execErr := exec.Command(hostnameCmd, args...).Output()
	if execErr != nil {
		log.Print(execErr)
	}

	return &Host{
		Hostname: strings.Trim(string(hostname), "\n"),
	}
}
