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
		log.Fatalln("Dial,error occurred:", err)
	}
	//create sftp client with ssh
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		log.Fatalln("NewClient,error occurred:", err)
	}
	cliConf.sshClient = sshClient
	cliConf.sftpClient = sftpClient
}

func (cliConf *ClientConfig) Reconnect() {
	var (
		//sshClient  *ssh.Client
		//sftpClient *sftp.Client
		err error
	)
	config := ssh.ClientConfig{
		User: cliConf.Username,
		Auth: []ssh.AuthMethod{ssh.Password(cliConf.Password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", cliConf.Host, cliConf.Port)
	if cliConf.sshClient, err = ssh.Dial("tcp", addr, &config); err != nil {
		log.Fatalln("Dial, error occurred:", err)
	}
	//create sftp client with ssh
	if cliConf.sftpClient, err = sftp.NewClient(cliConf.sshClient); err != nil {
		log.Fatalln("NewClient,error occurred:", err)
	}
	//cliConf.sshClient = sshClient
	//cliConf.sftpClient = sftpClient
}

func (cliConf *ClientConfig) RunShell(shell string) string {
	var session *ssh.Session
	var err error
	//get session
	println("执行命令：", shell)
	session, err = cliConf.sshClient.NewSession()
	if err != nil {
		log.Fatalln("NewSession,error occurred:", err)
	}
	//execute cmd
	output, _ := session.CombinedOutput(shell)

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
		dstFile.Close()
		srcFile.Close()
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
		srcFile.Close()
		dstFile.Close()
	}()

	if _, err1 := srcFile.WriteTo(dstFile); err1 != nil {
		log.Fatalln("file write error:", err1)
	}
	ret := fmt.Sprintf("down load: %s", dst)
	fmt.Println(ret)
	return ret
}

// test factory tools
func quitExe(control *ClientConfig) {
	os.Exit(0)
}

func genGDC(control *ClientConfig) {
	control.RunShell(" python3 /userdata/lfi/camera_params/gdc_map.py /userdata/lfi/camera_params/preposition_intrinsic.json /userdata/lfi/camera_params/layout_preposition.json /userdata/lfi/camera_params/gdc_map_preposition.txt")
}

func checkFile(control *ClientConfig) {
	control.RunShell("ls -al /userdata/lfi/camera_params/")
	control.RunShell("cat  /userdata/lfi/json_config.json | grep code")
}

func uploadCameraFile(control *ClientConfig) {
	//preSn, panoSn = httpService.getPrePanoCameraSn()
	//preFile = "./camera/preposition/" + preSn + ".json"
	//panoFile = "./camera/panoramic/" + panoSn + ".yaml"
	//control.reconnect()
	//control.UploadFile(preFile, "/userdata/lfi/camera_params/preposition_intrinsic.json")
	//control.UploadFile(panoFile, "/userdata/lfi/camera_params/panoramic_intrinsic.yaml")
}

func copyGDCScript(control *ClientConfig) {
	control.RunShell("cp /root/novabot/ota_lib/camera_params/gdc_map.py  /userdata/lfi/camera_params/")
	control.RunShell("cp /root/novabot/ota_lib/camera_params/layout_preposition.json  /userdata/lfi/camera_params/")
}

func upgrade(control *ClientConfig) {
	log.Println("开始手动OTA升级...")

	//control.reconnect()
	control.UploadFile("d:/lfimvpfactory20231014475.deb", "/root/lfimvpfactory20231014475.deb")
	print("文件上传成功,开始解压文件...")
	control.RunShell("dpkg -x lfimvpfactory20231014475.deb novabot.new")
	print("文件解压完成，等待重启...\n")
	control.RunShell("ls")
	control.RunShell("cp /root/novabot.new/scripts/run_ota.sh /userdata/ota/")
	control.RunShell("echo 1 > /userdata/ota/upgrade.txt")
	control.RunShell("cat /userdata/ota/upgrade.txt")
	time.Sleep(5 * time.Second)

	print("远程命令执行完成，开始重启机器")
	control.RunShell("reboot -f")
	print("等待30s后继续执行数据清理...")
	time.Sleep(30 * time.Second)
	control.Reconnect()
	control.RunShell("rm -rf /root/lfimvpfactory20231014475.deb")
	control.RunShell("rm -rf /root/novabot.bak")
	control.RunShell("cat /userdata/ota/upgrade.txt")
	control.RunShell("ls -alh /root/novabot/")
	time.Sleep(40)
	checkFile(control)
}

func getMac(control *ClientConfig) {
	control.RunShell("bash /usr/bin/startbt6212.sh & ") //获取蓝牙MAC，此处会一直阻塞在这里，需优化执行脚本
	control.RunShell("cat /bl.cfg | grep \"BD Address\"  | awk '{print $3}'")
}

func startService(control *ClientConfig) {
	control.RunShell("ps -aux | grep tof_camera_node")
	/*num := input("是否需要重启StartService 1: 重启，0，不重启")
	if num == 1 {
		print("重启服务")
		control.remoteCmd("~/novabot/scripts/start_service.sh")
	}
	else {
		print("不需要重启服务")
	}*/
}

func recharge(control *ClientConfig) {
	control.RunShell("/root/novabot/debug_sh/test_recharge.sh ")
}

func agingTest(control *ClientConfig) {
	control.RunShell("nohup python3 /root/novabot/debug_sh/chassis_Aging_Test.py&")
}

func model2User(control *ClientConfig) {
	control.RunShell("sed -i 's/flag=true/flag=false/' /root/novabot/test_scripts/factory_test/start_test.sh ")
}

func model2Factory(control *ClientConfig) {
	control.RunShell("sed -i 's/flag=false/flag=true/' /root/novabot/test_scripts/factory_test/start_test.sh ")
}

func modelStatus(control *ClientConfig) {
	control.RunShell("grep 'flag=' /root/novabot/test_scripts/factory_test/start_test.sh ")
}

func checkGDC(control *ClientConfig) {
	control.UploadFile("d:/verify_txt.py", "/root/verify_txt.py")
	control.RunShell("python3 verify_txt.py -g /userdata/lfi/camera_params/gdc_map_preposition.txt")
}

func factory_tools() {
	client := new(ClientConfig)
	client.createClient("192.168.1.10", 22, "root", "root") //"lFi@NovaBot"
	//client.RunShell("ls -alh")
	switch_dict := map[int64]func(*ClientConfig){
		0:  quitExe,
		1:  genGDC,
		2:  uploadCameraFile,
		3:  upgrade,
		4:  checkFile,
		5:  getMac,
		6:  copyGDCScript,
		7:  startService,
		8:  recharge,
		9:  agingTest,
		10: model2User,
		11: model2Factory,
		12: modelStatus,
		13: checkGDC,
	}

	print("操作说明：\n",
		"0：退出\n",
		"1：生成GDC文件\n",
		"2：上传相机内参文件\n",
		"3：手动OTA升级\n",
		"4：文件完整性检测\n",
		"5: 获取X3蓝牙MAC\n",
		"6: 拷贝GDC执行脚本文件\n",
		"7: 检查是否需要重启Service\n",
		"8: TTT===自动回充测试\n",
		"9: TTT===老化测试\n",
		"10: ===>切换到用户模式\n",
		"11: ===>切换到工厂模式\n",
		"12: ===>工作模式状态查询\n",
		"13:GDC文件完整性检测\n",
	)
	var x int64
	fmt.Println("请输入操作选项:")
	fmt.Scan(&x)
	func_name := switch_dict[x]
	func_name(client)

}

func test() {
	client := new(ClientConfig)
	client.createClient("192.168.1.10", 22, "root", "lFi@NovaBot")
	//var x int
	//fmt.Println("input  a int number:")
	//fmt.Scan(&x)
	//fmt.Printf("读取到内容:%d\n", x)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		client.RunShell("ls")
		wg.Done()
	}()
	go func() {
		client.DownloadFile("/root/novabot/test.go", "D:/test3.go")
		wg.Done()
	}()
	go func() {
		client.UploadFile("D:/test.go", "/root/novabot/test.go")
		wg.Done()
	}()

	wg.Wait()
	log.Println("goroutines are all finished")
}

func main() {
	//test()
	factory_tools()
}
