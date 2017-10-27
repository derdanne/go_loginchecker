package main

import (
	"os/exec"
	"os"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
	"fmt"
	"strings"
	"bufio"
	"bytes"
)

type config struct {
	Mailfrom string `yaml:"mailfrom"`
	Mailsubject string `yaml:"mailsubject"`
	Mailtos []string `yaml:"mailtos"`
	Ipaddresses []string `yaml:"ipaddresses"`
}

func (c *config) getConfig(configfile string) *config {
	yamlFile, readErr := ioutil.ReadFile(configfile)
	if readErr != nil {
		log.Printf("yamlFile.Get err   #%v ", readErr)
		panic(readErr)
	}

	yamlErr := yaml.Unmarshal(yamlFile, &c)
	if yamlErr != nil {
		log.Fatalf("Unmarshal: %v", yamlErr)
		panic(yamlErr)
	}

	return c
}

func getWho() string {

	var (
		whoOut []byte
		execErr error
	)

	who, lookErr := exec.LookPath("who")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-u"}

	whoOut, execErr = exec.Command(who, args...).Output();
	if execErr != nil {
		panic(execErr)
	}
	return string(whoOut)
}

func isAllowed(address string, allowedAddresses []string) bool {
	for _, allowedAddress := range allowedAddresses {
		if allowedAddress == address {
			return false
		}
	}
	return true
}

func hostname() string {
	uname, lookErr := exec.LookPath("uname")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-n"}

	hostname, execErr := exec.Command(uname, args...).Output();
	if execErr != nil {
		panic(execErr)
	}

	return string(hostname)
}



func sendMail(recipient string, mailfrom string, mailsubject string, mailbody string) {

	var sendmailBuffer bytes.Buffer

	sendmailBuffer.WriteString("\nTo: ")
	sendmailBuffer.WriteString(recipient)
	sendmailBuffer.WriteString("\nSubject: ")
	sendmailBuffer.WriteString(mailsubject)
	sendmailBuffer.WriteString("\nLoginchecker triggered an alert\n")
	sendmailBuffer.WriteString("\n\n\n")
	sendmailBuffer.WriteString(mailbody)
	sendmailBuffer.WriteString("\n.")

	fmt.Println(sendmailBuffer.String())

	mail, lookErr := exec.LookPath("sendmail")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-t", "-f", mailfrom}

	sendmail := exec.Command(mail, args...);
	sendmail.Stdin = strings.NewReader(sendmailBuffer.String())
	var out bytes.Buffer
	sendmail.Stdout = &out
	execErr := sendmail.Run()

	if execErr != nil {
		log.Fatal(execErr)
	}
	fmt.Printf("in all caps: %q\n", out.String())

	fmt.Println(recipient, mailfrom, mailsubject, sendmailBuffer.String())
}


func main () {
	var (
		config config
		line string
	)

	configfile := "./config.yml"
	if  len(os.Args) > 1 {
		arg := os.Args[1]
		configfile = arg
	}

	config.getConfig(configfile)

	alertMailtos := config.Mailtos
	allowedAddresses := config.Ipaddresses
	mailFrom := config.Mailfrom
	mailSubject := config.Mailsubject + " @ " + hostname()
	fmt.Println("Allowed addresses: ", allowedAddresses)

	who := getWho()
	scanner := bufio.NewScanner(strings.NewReader(who))
	for scanner.Scan() {
		line = scanner.Text()
		whoIsLoggedIn := strings.Fields(line)
		if isAllowed(strings.Trim(whoIsLoggedIn[6], "(,)"), allowedAddresses) {
			mailBody := whoIsLoggedIn[0] + " is not allowed from " + strings.Trim(whoIsLoggedIn[6], "(,)")
			for _, alertMailto := range alertMailtos {
				sendMail(alertMailto, mailFrom, mailSubject, mailBody)
			}
		}
	}
}