package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"io/ioutil"
	"log"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.Dial("imap.vip.sina.com:143")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("dengyun@vip.sina.com", "b54aa91e652a203a"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	//mailboxes := make(chan *imap.MailboxInfo, 10)
	//done := make(chan error, 1)
	//go func () {
	//	done <- c.List("", "*", mailboxes)
	//}()
	//
	//log.Println("Mailboxes:")
	//for m := range mailboxes {
	//	log.Println("* " + m.Name)
	//}
	//
	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("Flags for INBOX:", mbox.Flags)

	// Get all messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 10 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 10
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Fetching all messages:")
	section := &imap.BodySectionName{}
	for msg := range messages {
		log.Println(msg.SeqNum)
		log.Println("* " + msg.Envelope.Subject)
		r := msg.GetBody(section)
		if r != nil {
			body, _ := ioutil.ReadAll(r)
			log.Println(body)
		}else{
			fmt.Println("body section not found")
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}