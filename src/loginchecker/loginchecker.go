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

type configMail struct {
	From       string   `yaml:"from,omitempty"`
	FromName   string   `yaml:"from_name,omitempty"`
	Subject    string   `yaml:"subject,omitempty"`
	Recipients []string `yaml:"recipients,omitempty"`
}

type configSlack struct {
	WebhookUrl string `yaml:"webhook_url,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
	Author     string `yaml:"author,omitempty"`
	Message    string `yaml:"message,omitempty"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
}

type config struct {
	AllowedAddresses []string    `yaml:"allowed_addresses,omitempty"`
	AllowedUsers     []string    `yaml:"allowed_users,omitempty"`
	RecheckTime      int64       `yaml:"recheck_time,omitempty"`
	GraceTime        int64       `yaml:"grace_time,omitempty"`
	Mail             configMail  `yaml:"mail,omitempty"`
	Slack            configSlack `yaml:"slack,omitempty"`
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

func isNotAllowedIp(address string, allowedNetworks []string) bool {
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

func isNotAllowedUser(user string, allowedUsers []string) bool {
	for _, allowedUser := range allowedUsers {
		if allowedUser == user {
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

	hostFqdn, execErr := exec.Command(hostname, args...).Output()
	if execErr != nil {
		panic(execErr)
	}

	return strings.Trim(string(hostFqdn), "\n")
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
		Color:      "danger",
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
	enableMail := true
	enableSlack := true

	configfile := "./config.yml"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		configfile = arg
	}
	config.getConfig(configfile)

	if config.Mail.Recipients == nil || config.Mail.From == "" || config.Mail.FromName == "" || config.Mail.Subject == "" {
		println("Alerting via email is turned off. If you want to enable email alerting, make sure you set the needed config params!")
		enableMail = false
	}

	if config.Slack.WebhookUrl == "" || config.Slack.Channel == "" || config.Slack.Author == "" || config.Slack.Message == "" || config.Slack.Username == "" || config.Slack.IconEmoji == "" {
		println("Alerting via slack is turned off. If you want to enable slack alerting, make sure you set the needed config params!")
		enableSlack = false
	}

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

			if isNotAllowedIp(address, config.AllowedAddresses) || isNotAllowedUser(username, config.AllowedUsers) {
				notify := false

				timeDetected := time.Now()
				if timeLogged, ok := logged[uniqeUser]; ok {
					if timeDetected.Sub(timeLogged) > time.Duration(config.GraceTime)*time.Second {
						logged[uniqeUser] = timeDetected
						notify = true
					}
				} else {
					logged[uniqeUser] = timeDetected
					notify = true
				}

				if notify {
					message := "Unauthorized access: User " + username + " is not allowed to access " + hostName + " from IP " + address + "!"

					if enableMail {
						for _, alertMailto := range config.Mail.Recipients {
							sendMail(alertMailto, config.Mail.From, config.Mail.FromName, config.Mail.Subject+" @ "+hostName, message)
						}
					}

					if enableSlack {
						sendSlack(config.Slack.WebhookUrl, config.Slack.Channel, config.Slack.Author, config.Slack.Message+" @ "+hostName, config.Slack.Username, config.Slack.IconEmoji, message)
					}

					println(message)
				}
			}
		}
		time.Sleep(time.Duration(config.RecheckTime) * time.Second)
	}
}
