package main

import (
	"bufio"
	"os/exec"
	"strings"
	"time"
)

type User struct {
	Username     string
	Hostname     string
	TimeDetected time.Time
	TimeNotified time.Time
}

func createUsersSlice() []*User {
	users := make([]*User, 0)
	return users
}

func isNewUser(username string, hostname string, Users []*User) bool {
	for _, user := range Users {
		if user.Username == username && user.Hostname == hostname {
			return false
		}
	}
	return true
}

func (user *User) UpdateTimeNotified(time time.Time) {
	user.TimeNotified = time
}

func detectUser(userslice []*User) []*User {

	WhoCmd, lookErr := exec.LookPath("who")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-u"}

	whoOut, execErr := exec.Command(WhoCmd, args...).Output()
	if execErr != nil {
		panic(execErr)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(whoOut)))
	for scanner.Scan() {
		whoIsLoggedIn := strings.Fields(scanner.Text())
		username := whoIsLoggedIn[0]
		hostname := strings.Trim(whoIsLoggedIn[6], "(,)")
		if strings.Contains(hostname, ":S") {
			hostname = string(strings.Split(hostname, ":")[0])
		}

		if isNewUser(username, hostname, userslice) {
			timestamp := time.Now()
			userslice = append(userslice, &User{Username: username, Hostname: hostname, TimeDetected: timestamp, TimeNotified: timestamp})
		}
	}
	return userslice
}
