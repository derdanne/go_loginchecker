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
	"net"
)

type config struct {
	Mailfrom string `yaml:"mail_from"`
	Mailfromname string `yaml:"mail_from_name"`
	Mailsubject string `yaml:"mail_subject"`
	Mailtos []string `yaml:"mail_to"`
	Allowedaddresses []string `yaml:"allowed_addresses"`
	Rechecktime int64 `yaml:"recheck_time"`
	Gracetime int64 `yaml:"grace_time"`
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
	who, lookErr := exec.LookPath("who")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"-u"}

	whoOut, execErr := exec.Command(who, args...).Output();
	if execErr != nil {
		panic(execErr)
	}
	return string(whoOut)
}

func isNotAllowed(address string, allowedNetworks []string) bool {
	parsedAddress := net.ParseIP(address)
	if parsedAddress == nil {
		for _, allowedHostname := range allowedNetworks {
			if allowedHostname == address {
				return false
			}
		}
	} else {
		for _, allowedNetwork := range allowedNetworks {
			_, parsedAllowedNetworkCIDR, _ := net.ParseCIDR(allowedNetwork)
			if parsedAllowedNetworkCIDR == nil {
			} else {
				if parsedAllowedNetworkCIDR.Contains(parsedAddress) {
					return false
				}
			}
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

	hostFqdn, execErr := exec.Command(hostname, args...).Output();
	if execErr != nil {
		panic(execErr)
	}

	return string(hostFqdn)
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
		panic(execErr)
	}
}

func main () {
	var (
		config config
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
	allowedAddresses := config.Allowedaddresses
	mailFrom := config.Mailfrom
	mailFromName := config.Mailfromname
	mailSubject := config.Mailsubject + " @ " + hostname()

	logged := make(map[string]time.Time)

	for {
		who := getWho()
		scanner := bufio.NewScanner(strings.NewReader(who))
		for scanner.Scan() {
			line := scanner.Text()
			whoIsLoggedIn := strings.Fields(line)
			username := whoIsLoggedIn[0]
			address := strings.Trim(whoIsLoggedIn[6], "(,)")
                        uniqeUser := username + address

			if isNotAllowed(address, allowedAddresses) {
				sendmail := false

				timeDetected := time.Now()
				if timeLogged, ok := logged[uniqeUser]; ok {
					if (timeDetected.Sub(timeLogged) > graceTime) {
						logged[uniqeUser] = timeDetected
						sendmail = true
					}
				} else {
					logged[uniqeUser] = timeDetected
					sendmail = true
				}

				if sendmail {
					mailBody := username + " is not allowed from " + address
					for _, alertMailto := range alertMailtos {
						sendMail(alertMailto, mailFrom, mailFromName, mailSubject, mailBody)
					}
				}
			}
		}
		time.Sleep(recheckTime)
	}
}