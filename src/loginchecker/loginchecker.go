package main

import (
	"os/exec"
	"os"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
	"strings"
	"bufio"
	"bytes"
	"time"
)

type config struct {
	Mailfrom string `yaml:"mailfrom"`
	Mailfromname string `yaml:"mailfromname"`
	Mailsubject string `yaml:"mailsubject"`
	Mailtos []string `yaml:"mailtos"`
	Ipaddresses []string `yaml:"ipaddresses"`
	Rechecktime int64 `yaml:"rechecktime"`
	Gracetime int64 `yaml:"gracetime"`
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

func isNotAllowed(address string, allowedAddresses []string) bool {
	for _, allowedAddress := range allowedAddresses {
		if allowedAddress == address {
			return false
		}
	}
	return true
}

func hostname() string {
	hostname, lookErr := exec.LookPath("hostname")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-f"}

	hostfqdn, execErr := exec.Command(hostname, args...).Output();
	if execErr != nil {
		panic(execErr)
	}

	return string(hostfqdn)
}



func sendMail(recipient string, mailfrom string, mailfromname string, mailsubject string, mailbody string) {

	var sendmailBuffer bytes.Buffer

	sendmailBuffer.WriteString("Subject: ")
	sendmailBuffer.WriteString(mailsubject)
	sendmailBuffer.WriteString("\nLoginchecker triggered an alert\n")
	sendmailBuffer.WriteString("\n\n")
	sendmailBuffer.WriteString(mailbody)
	sendmailBuffer.WriteString("\n.")

	mail, lookErr := exec.LookPath("sendmail")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-t", "-f", mailfrom, "-F", mailfromname, recipient}

	sendmail := exec.Command(mail, args...);
	sendmail.Stdin = strings.NewReader(sendmailBuffer.String())
	execErr := sendmail.Run()
	if execErr != nil {
		log.Fatal(execErr)
	}
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

	recheckTime := time.Duration(config.Rechecktime) * time.Second
	graceTime := time.Duration(config.Gracetime) * time.Second
	alertMailtos := config.Mailtos
	allowedAddresses := config.Ipaddresses
	mailFrom := config.Mailfrom
	mailFromName := config.Mailfromname
	mailSubject := config.Mailsubject + " @ " + hostname()

	println(mailFrom, mailFromName, mailSubject)

	logged := make(map[string]time.Time)

	for {
		who := getWho()
		scanner := bufio.NewScanner(strings.NewReader(who))
		for scanner.Scan() {
			line = scanner.Text()
			whoIsLoggedIn := strings.Fields(line)
			username := whoIsLoggedIn[0]
			address := strings.Trim(whoIsLoggedIn[6], "(,)")
                        uniqe_user := username + address

			if isNotAllowed(address, allowedAddresses) {
				mailBody := username + " is not allowed from " + address
				sendmail := false

				timeDetected := time.Now()
				if timeLogged, ok := logged[uniqe_user]; ok {
					if (timeDetected.Sub(timeLogged) > graceTime) {
						logged[uniqe_user] = timeDetected
						sendmail = true
					}

				} else {
					logged[uniqe_user] = timeDetected
					sendmail = true
				}

				if sendmail == true {
					for _, alertMailto := range alertMailtos {
						sendMail(alertMailto, mailFrom, mailFromName, mailSubject, mailBody)
					}
				}
			}
		}
		time.Sleep(recheckTime)
	}
}