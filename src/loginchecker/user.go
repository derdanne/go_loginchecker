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

func (user *User) updateTimeNotified(time time.Time) {
	user.TimeNotified = time
}

func createUsersSlice() []*User {
	users := make([]*User, 0)
	return users
}

func detectUser(userSlice []*User) []*User {
	tmpUserSlice := createUsersSlice()

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
		timestamp := time.Now()
		tmpUserSlice = append(tmpUserSlice, &User{Username: username, Hostname: hostname, TimeDetected: timestamp, TimeNotified: timestamp})
	}

	for _, user := range userSlice {
		if isUserNotInSlice(user.Username, user.Hostname, tmpUserSlice) {
			userSlice = removeUserFromSlice(user.Username, user.Hostname, userSlice)
		}
	}

	for _, user := range tmpUserSlice {
		if isUserNotInSlice(user.Username, user.Hostname, userSlice) {
			userSlice = append(userSlice, &User{Username: user.Username, Hostname: user.Hostname, TimeDetected: user.TimeDetected, TimeNotified: user.TimeNotified})
		}
	}
	tmpUserSlice = nil
	return userSlice
}

func isUserNotInSlice(username string, hostname string, userSlice []*User) bool {
	for _, user := range userSlice {
		if user.Username == username && user.Hostname == hostname {
			return false
		}
	}
	return true
}

func removeUserFromSlice(username string, hostname string, userSlice []*User) []*User {
	k := 0
	for _, user := range userSlice {
		if user.Username != username || user.Hostname != hostname {
			userSlice[k] = user
			k++
		}
	}
	return userSlice[:k]
}


