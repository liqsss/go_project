package main

import (
	"fmt"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func testcode() {
	fmt.Println("hello world")
	result1 := slice.AppendIfAbsent([]string{"a", "b"}, "b")
	result2 := slice.AppendIfAbsent([]string{"a", "b"}, "c")

	fmt.Println(result1)
	fmt.Println(result2)

	macAddrs := netutil.GetMacAddrs()
	fmt.Println(macAddrs)
}

type ClientConfig struct {
	Host       string       //ip
	Port       int64        // 端口
	Username   string       //用户名
	Password   string       //密码
	sshClient  *ssh.Client  //ssh client
	sftpClient *sftp.Client //sftp client
	LastResult string       //最近一次运行的结果
}

func (cliConf *ClientConfig) createClient(host string, port int64, username, password string) {
	var (
		sshClient  *ssh.Client
		sftpClient *sftp.Client
		err        error
	)
	cliConf.Host = host
	cliConf.Port = port
	cliConf.Username = username
	cliConf.Password = password
	cliConf.Port = port
	config := ssh.ClientConfig{
		User: cliConf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", cliConf.Host, cliConf.Port)
	if sshClient, err = ssh.Dial("tcp", addr, &config); err != nil {
		log.Fatalln("error occurred:", err)
	}
	//create sftp client with ssh
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		log.Fatalln("error occurred:", err)
	}
	cliConf.sshClient = sshClient
	cliConf.sftpClient = sftpClient
}

func (cliConf *ClientConfig) RunShell(shell string) string {
	var session *ssh.Session
	var err error
	//get session
	session, err = cliConf.sshClient.NewSession()
	if err != nil {
		log.Fatalln("error occurred:", err)
	}
	//execute cmd
	output, e := session.CombinedOutput(shell)
	if e != nil {
		log.Fatalln("error occurred:", err)
	}
	fmt.Println(string(output))
	return string(output)
}

func (cliConf *ClientConfig) UploadFile(src, dst string) string {
	dstFile, e := cliConf.sftpClient.Create(dst)
	if e != nil {
		log.Fatalln("create reomote file error")
	}

	srcFile, err := os.Open(src)
	if err != nil {
		log.Fatalln("open local file error")
	}
	defer func() {
		defer dstFile.Close()
		defer srcFile.Close()
	}()

	dstFile.ReadFrom(srcFile)
	return cliConf.RunShell(fmt.Sprintf("ls -al %s", dst))
}

func (cliConf *ClientConfig) DownloadFile(src, dst string) string {
	srcFile, err := cliConf.sftpClient.Open(src)
	if err != nil {
		log.Fatalln("read file error：", err)
	}

	dstFile, e := os.Create(dst)
	if e != nil {
		log.Fatalln("file create error:", e)
	}
	defer func() {
		defer srcFile.Close()
		defer dstFile.Close()
	}()

	if _, err1 := srcFile.WriteTo(dstFile); err1 != nil {
		log.Fatalln("file write error:", err1)
	}
	ret := fmt.Sprintf("down load: %s", dst)
	fmt.Println(ret)
	return ret
}

func main() {
	client := new(ClientConfig)
	client.createClient("192.168.1.10", 22, "root", "lFi@NovaBot")
	//var x int
	//fmt.Println("input  a int number:")
	//fmt.Scan(&x)
	//fmt.Printf("读取到内容:%d\n", x)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		client.RunShell("ls -al")
		wg.Done()
	}()

	go func() {
		client.UploadFile("D:/db_novabot.sql", "/root/novabot/nova.sql")
		wg.Done()
	}()
	go func() {
		client.DownloadFile("/root/novabot/test.go", "D:/test3.go")
		wg.Done()
	}()

	wg.Wait()
}
