package main

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func sendMail(recipient string, mailfrom string, mailfromname string, mailsubject string, mailbody string) {
	var sendmailBuffer bytes.Buffer
	sendmailBuffer.WriteString("Subject: ")
	sendmailBuffer.WriteString(mailsubject)
	sendmailBuffer.WriteString("\n")
	sendmailBuffer.WriteString(mailbody)
	sendmailBuffer.WriteString("\n.")

	mail, lookErr := exec.LookPath("sendmail")
	if lookErr != nil {
		log.Print(lookErr)
	}

	args := []string{"-t", "-f", mailfrom, "-F", mailfromname, recipient}

	sendmail := exec.Command(mail, args...)
	sendmail.Stdin = strings.NewReader(sendmailBuffer.String())
	execErr := sendmail.Run()
	if execErr != nil {
		log.Print(execErr)
	}
}
