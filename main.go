package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var (
	server, email, password string
	workernum               int
	imapClient              *client.Client
	mailDirs                []string
	mailSumNums             uint32
)

func main() {
	//获取命令行参数
	//go run main.go -server imap.vip.sina.com:143 -email dengyun@vip.sina.com -password b54aa91e652a203a
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
	os.Mkdir(email, os.ModePerm)
	mailboxes := make(chan *imap.MailboxInfo, 20)
	go func() {
		imapClient.List("", "*", mailboxes)
	}()
	//列取邮件夹
	for m := range mailboxes {
		mailDirs = append(mailDirs, m.Name)
	}

	for _, mailDir := range mailDirs {
		//选中每个邮件夹
		mbox, err := imapClient.Select(mailDir, false)
		if err != nil {
			log.Fatal(err)
		}
		mailDirNums := mbox.Messages
		log.Printf("%s : %d", mailDir, mailDirNums)
		fileDir := fmt.Sprintf("%s/%s_%d", email, mailDir, mailDirNums)
		//创建邮件夹目录
		os.Mkdir(fileDir, os.ModePerm)
		mailSumNums += mailDirNums
	}
	log.Printf("总邮件数 : %d", mailSumNums)
	for _, mailDir := range mailDirs {
		//选中每个邮件夹
		mbox, err := imapClient.Select(mailDir, false)
		if err != nil {
			log.Fatal(err)
		}
		//循环该邮件夹中的邮件
		seqset := new(imap.SeqSet)
		seqset.AddRange(1, mbox.Messages)
		section := &imap.BodySectionName{}
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, mbox.Messages)
		go func() {
			imapClient.Fetch(seqset, items, messages)
		}()

		for msg := range messages {
			mailFile := fmt.Sprintf("%s/%s_%d/%d.eml", email, mailDir, mbox.Messages, msg.SeqNum)
			r := msg.GetBody(section)
			if r == nil {
				log.Printf("%s-%dServer didn't returned message body", mailDir, msg.SeqNum)
			}
			if r != nil {
				body, err := ioutil.ReadAll(r)
				if err != nil {
					log.Printf("%s:%d ioutil.ReadAll error", mailDir, msg.SeqNum)
				}

				file6, err := os.OpenFile(mailFile, os.O_RDWR|os.O_CREATE, 0766)
				if err != nil {
					log.Printf("%s:%d os.OpenFile error %s", mailDir, msg.SeqNum, mailFile)
				}
				file6.Write(body)
				file6.Close()
				log.Printf("%s :第 %d ", mailDir, msg.SeqNum)
			}
		}
	}
}