package main

import (
	"flag"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
)

func init()  {
	log.SetFlags(log.Llongfile|log.LstdFlags)
}
var (
	server, email, password string
	workernum               int
)

func PutData(email string) (err error) {

	var f *os.File
	//write csv record
	csvFile := "./data.csv"
	f, err = os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE, 0755)

	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write([]byte(email+"\n"))
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	//获取命令行参数
	flag.StringVar(&server, "server", "", "imap服务地址(包含端口)")
	flag.StringVar(&email, "email", "", "邮箱名")
	flag.StringVar(&password, "password", "", "密码")
	flag.IntVar(&workernum, "workernum", 32, "并发数:")
	flag.Parse()
	if flag.NFlag() < 3 {
		flag.PrintDefaults()
		log.Fatal("参数缺失")
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

	// Select INBOX
	mbox, err := imapClient.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	// Get the last message
	if mbox.Messages == 0 {
		log.Fatal("No message in mailbox")
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}
	messages := make(chan *imap.Message, 10)
	go func() {
		if err := imapClient.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()
	var i uint32
	i = 1
	for  {
		if i == mbox.Messages {
			break
		}
		msg := <-messages
		if msg == nil {
			log.Fatal("Server didn't returned message")
		}

		r := msg.GetBody(&section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		// Process each message's part
		//for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			b, _ := ioutil.ReadAll(p.Body)
			reEmail := `[\w\.]+@\w+\.[a-z]{2,3}(\.[a-z]{2,3})?`
			re := regexp.MustCompile(reEmail)
			//-1 表示匹配所有  如果输入5  表示只匹配5个
			email := re.FindStringSubmatch(string(b))
			if len(email) > 0 {
				err = PutData(email[0])
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(strconv.Itoa(int(i))+": "+email[0])
			}
		//}
		i++
	}
}