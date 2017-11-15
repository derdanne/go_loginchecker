package main

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func sendMail(recipient string, mailFrom string, mailFromName string, mailSubject string, mailBody string) {
	var sendMailBuffer bytes.Buffer
	sendMailBuffer.WriteString("Subject: ")
	sendMailBuffer.WriteString(mailSubject)
	sendMailBuffer.WriteString("\n")
	sendMailBuffer.WriteString(mailBody)
	sendMailBuffer.WriteString("\n.")

	mail, lookErr := exec.LookPath("sendmail")
	if lookErr != nil {
		log.Print(lookErr)
	}

	args := []string{"-t", "-f", mailFrom, "-F", mailFromName, recipient}

	sendMail := exec.Command(mail, args...)
	sendMail.Stdin = strings.NewReader(sendMailBuffer.String())
	execErr := sendMail.Run()
	if execErr != nil {
		log.Print(execErr)
	}
}
