package main

import (
	"fmt"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/slice"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func mqtt_pub(pubtopic string, subtopic string, msg string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://119.23.212.113:1883") //.SetClientID("go_tester_man" + pubtopic)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln(token.Error())
	}

	/*data, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("msg is not json format")
		return
	}*/
	var ch chan int
	ch = make(chan int, 1)

	client.Subscribe(subtopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		log.Println("接收数据已完成")
		defer func() {
			if err := recover(); err != nil {
				println("通道已关闭")
			}
		}()
		ch <- 1
	})
	if token := client.Publish(pubtopic, 1, false, string(msg)); token.Wait() && token.Error() != nil {
		log.Fatalln("publish msg error")
	}
	log.Print("数据发送完成!!!!")
	log.Println("等待数据接收")
	<-ch
	close(ch)
	log.Println("数据处理已完成")
}

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

func (cliConf *ClientConfig) createClient(host string, port int64, username, password string) bool {
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
		log.Print("Dial,error occurred:", err)
		return false
	}
	//create sftp client with ssh
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		log.Print("NewClient,error occurred:", err)
		return false
	}
	cliConf.sshClient = sshClient
	cliConf.sftpClient = sftpClient
	return true
}

func (cliConf *ClientConfig) Reconnect() bool {
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
		log.Print("Dial, error occurred:", err)
		return false
	}
	//create sftp client with ssh
	if cliConf.sftpClient, err = sftp.NewClient(cliConf.sshClient); err != nil {
		log.Print("NewClient,error occurred:", err)
		return false
	}
	return true
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
func quitExe() {
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
	println("开始手动OTA升级...")

	//control.reconnect()
	control.UploadFile("d:/lfimvpfactory20231014475.deb", "/root/lfimvpfactory20231014475.deb")
	println("文件上传成功,开始解压文件...")
	control.RunShell("dpkg -x lfimvpfactory20231014475.deb novabot.new")
	println("文件解压完成，等待重启...")
	control.RunShell("ls")
	control.RunShell("cp /root/novabot.new/scripts/run_ota.sh /userdata/ota/")
	control.RunShell("echo 1 > /userdata/ota/upgrade.txt")
	control.RunShell("cat /userdata/ota/upgrade.txt")
	time.Sleep(5 * time.Second)

	println("远程命令执行完成，开始重启机器")
	control.RunShell("reboot -f")
	println("等待30s后继续执行数据清理...")
	time.Sleep(30 * time.Second)
	control.Reconnect()
	control.RunShell("rm -rf /root/lfimvpfactory20231014475.deb")
	control.RunShell("rm -rf /root/novabot.bak")
	control.RunShell("cat /userdata/ota/upgrade.txt")
	control.RunShell("ls -alh /root/novabot/")
	time.Sleep(40 * time.Second)
	checkFile(control)
}

func upgrade2(control *ClientConfig) {

	println("开始升级...")
	control.RunShell("rm  -rf /root/novabot")
	control.RunShell("sync")
	control.UploadFile("d:/lfimvp-factory-20231114489.deb", "/root/lfimvp-factory-20231114489.deb")
	control.RunShell("sync")
	println("文件上传成功")
	control.RunShell("dpkg -x lfimvp-factory-20231114489.deb novabot")
	control.RunShell("sync")
	control.RunShell("sleep 2s")
	println("文件解压完成")
	control.RunShell("rm -rf /root/lfimvp-factory-20231114489.deb")
	control.RunShell("cat /root/novabot/Readme.txt")
	control.RunShell("/root/novabot/scripts/start_service.sh")
	println("完成升级")

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

func change2UserModle_charging() {
	var sn string
	fmt.Println("请扫描充电桩SN编码:")
	fmt.Scanln(&sn)

	pubtopic := "tools/charging_send_mqtt/" + sn
	subtopic := "tools/charging_receive_mqtt/" + sn
	log.Print("publisher topic:", pubtopic, " ,subscriber topic:", subtopic)
	msg := "{\"set_test_info\":{\"id\":4,\"value\":null}}\n"
	/*type Person struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}
	p := Person{Name: "liqiang", Age: "111"}
	datas, err := json.Marshal(p)
	if err != nil {
		log.Print("error in marshalling json")
	}*/
	mqtt_pub(pubtopic, subtopic, msg)
}

func factory_test_charge() {
	var x int
	switch_dict := map[int64]func(){
		1: change2UserModle_charging,
	}
	println("=======>>>>>>>>进入充电桩检测模式")
	fmt.Println("1: 充电桩切换到用户模式")
	fmt.Println("请选择功能\n" +
		"（如输入1）:")
	fmt.Scanln(&x)
	switch_dict[int64(x)]()

}

func factory_tools() {
	//client.RunShell("ls -alh")
	var x int
	switch_dict := map[int64]func(){
		0: quitExe,
		1: factory_test_x3,
		2: factory_test_charge,
	}
	for true {
		fmt.Println("<<<<<开始测试>>>>>")
		fmt.Println("0: 退出程序")
		fmt.Println("1: X3板")
		fmt.Println("2: 充电桩")
		fmt.Println("请选择需要操作的设备：\n" +
			"（如输入1:代表X3板，检测前请通过网线直连X3板，2：代表充电桩，检测前确保网络工装电脑网络通畅）")
		fmt.Scanln(&x)
		switch_dict[int64(x)]()
	}
}

func factory_test_x3() {
	switch_dict := map[int64]func(*ClientConfig){
		1:  genGDC,
		2:  uploadCameraFile,
		3:  upgrade2,
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
	var x int64
	client := new(ClientConfig)
	bret := client.createClient("192.168.1.10", 22, "root", "lFi@NovaBot") //"lFi@NovaBot"
	if !bret {
		log.Print("ssh connect error.")
	}
	println("=======>>>>>>>>进入X3机器检测模式")
	print("操作说明：\n",
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

	fmt.Println("请输入操作选项:")
	fmt.Scanln(&x)
	func_name := switch_dict[int64(x)]
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
