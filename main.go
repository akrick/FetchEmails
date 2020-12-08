package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func init()  {
	log.SetFlags(log.Llongfile|log.LstdFlags)
}
var (
	server, email, password string
	workernum               int
)

func determineEncoding(r io.Reader) encoding.Encoding  {
	bytes, err := bufio.NewReader(r).Peek(1024)
	if err != nil {
		panic(err)
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}
func main() {
	//获取命令行参数
	//go run main.go -server imap.sina.com:143 -email akrick@sina.com -password 22ef61051bfe5c09
	flag.StringVar(&server, "server", "", "imap服务地址(包含端口)")
	flag.StringVar(&email, "email", "", "邮箱名")
	flag.StringVar(&password, "password", "", "密码")
	flag.IntVar(&workernum, "workernum", 32, "并发数:")
	flag.Parse()
	if flag.NFlag() < 3 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if server == "" || email == "" || password == "" || workernum == 0 {
		log.Fatal("服务器地址,用户名,密码,参数错误")
	}
	//连接imap服务
	imapClient, err := client.Dial(server)
	if err != nil {
		log.Fatal(err)
	}
	//登陆
	if err := imapClient.Login(email, password); err != nil {
		log.Fatal(err)
	}
	//创建邮件夹目录
	_, err = os.Stat(email)
	if os.IsNotExist(err) {
		err = os.Mkdir(email, os.ModePerm)
		if err != nil{
			log. Fatal(err)
		}
	}
	// Select INBOX
	mbox, err := imapClient.Select("INBOX", false)
	if err != nil {
		log.Println("here we go")
		log.Fatal(err)
	}

	// Get the last message
	if mbox.Messages == 0 {
		log.Fatal("No message in mailbox")
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(1, mbox.Messages)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := imapClient.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	msg := <-messages
	if msg == nil {
		log.Fatal("Server didn't returned message")
	}

	r := msg.GetBody(&section)
	if r == nil {
		log.Fatal("Server didn't returned message body")
	}

	fmt.Println(r)
	os.Exit(0)

	// Create a new mail reader
	e := determineEncoding(r)
	//转为UTF8格式读
	utf8Reader := transform.NewReader(r, e.NewDecoder())
	ur, err := mail.CreateReader(utf8Reader)
	if err != nil {
		log.Fatal(err)
	}

	//Print some info about the message
	//header := ur.Header
	//if date, err := header.Date(); err == nil {
	//	log.Println("Date:", date)
	//}
	//if from, err := header.AddressList("From"); err == nil {
	//	log.Println("From:", from)
	//}
	//if to, err := header.AddressList("Reply to"); err == nil {
	//	log.Println("Reply to:", to)
	//}
	//if subject, err := header.Subject(); err == nil {
	//	log.Println("Subject:", subject)
	//}

	// Process each message's part
	for {
		p, err := ur.NextPart()
		if err == io.EOF {
			log.Println("done")
			break
		} else if err != nil {
			log.Fatal(err)
		}
		switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				log.Println("Body:", string(b))
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.Println("Got attachment:", filename)
		}
	}
}