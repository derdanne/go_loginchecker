package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/monochromegane/slack-incoming-webhooks"
	"gopkg.in/yaml.v2"
)

type config struct {
	Mailfrom         string   `yaml:"mail_from,omitempty"`
	Mailfromname     string   `yaml:"mail_from_name,omitempty"`
	Mailsubject      string   `yaml:"mail_subject,omitempty"`
	Mailtos          []string `yaml:"mail_to,omitempty"`
	Allowedaddresses []string `yaml:"allowed_addresses,omitempty"`
	Rechecktime      int64    `yaml:"recheck_time,omitempty"`
	Gracetime        int64    `yaml:"grace_time,omitempty"`
	Slackwebhookurl  string   `yaml:"slack_webhook_url,omitempty"`
	Slackchannel     string   `yaml:"slack_channel,omitempty"`
	Slackauthor      string   `yaml:"slack_author,omitempty"`
	Slackmessage     string   `yaml:"slack_message,omitempty"`
	Slackusername    string   `yaml:"slack_username,omitempty"`
	Slackiconemoji   string   `yaml:"slack_icon_emoji,omitempty"`
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

	whoOut, execErr := exec.Command(who, args...).Output()
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

	hostFqdn, execErr := exec.Command(hostname, args...).Output()
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

	sendmail := exec.Command(mail, args...)
	sendmail.Stdin = strings.NewReader(sendmailBuffer.String())
	execErr := sendmail.Run()
	if execErr != nil {
		panic(execErr)
	}
}

func sendSlack(webhookUrl string, channel string, author string, message string, username string, iconEmoji string, attachementMessage string) {
	client := &slack_incoming_webhooks.Client{
		WebhookURL: webhookUrl,
	}

	attachment := &slack_incoming_webhooks.Attachment{
		AuthorName: author,
		Text:       attachementMessage,
		Color:      "red",
	}

	payload := &slack_incoming_webhooks.Payload{
		Text:      message,
		Channel:   channel,
		Username:  username,
		IconEmoji: iconEmoji,
	}
	payload.AddAttachment(attachment)

	execErr := client.Post(payload)
	if execErr != nil {
		panic(execErr)
	}
}

func main() {
	var (
		config config
	)

	hostName := hostname()

	configfile := "./config.yml"
	if len(os.Args) > 1 {
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
	mailSubject := config.Mailsubject + " @ " + hostName
	slackWebhookUrl := config.Slackwebhookurl
	slackChannel := config.Slackchannel
	slackAuthor := config.Slackauthor
	slackMessage := config.Slackmessage + " @ " + hostName
	slackUsername := config.Slackusername
	slackIconEmoji := config.Slackiconemoji

	logged := make(map[string]time.Time)
	for {
		who := getWho()
		scanner := bufio.NewScanner(strings.NewReader(who))
		for scanner.Scan() {
			line := scanner.Text()
			whoIsLoggedIn := strings.Fields(line)
			username := whoIsLoggedIn[0]
			address := strings.Trim(whoIsLoggedIn[6], "(,)")
			if strings.Contains(address, ":S") {
				address = string(strings.Split(address, ":")[0])
			}

			uniqeUser := username + address

			if isNotAllowed(address, allowedAddresses) {
				notify := false

				timeDetected := time.Now()
				if timeLogged, ok := logged[uniqeUser]; ok {
					if timeDetected.Sub(timeLogged) > graceTime {
						logged[uniqeUser] = timeDetected
						notify = true
					}
				} else {
					logged[uniqeUser] = timeDetected
					notify = true
				}

				if notify {
					message := username + " is not allowed from " + address
					for _, alertMailto := range alertMailtos {
						sendMail(alertMailto, mailFrom, mailFromName, mailSubject, message)
					}
					sendSlack(slackWebhookUrl, slackChannel, slackAuthor, slackMessage, slackUsername, slackIconEmoji, message)
				}
			}
		}
		time.Sleep(recheckTime)
	}
}
