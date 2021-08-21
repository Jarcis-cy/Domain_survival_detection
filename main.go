package main

import (
	"Domain_survival_detection/golimit"
	"bufio"
	"flag"
	"fmt"
	"github.com/imroc/req"
	"github.com/schollz/progressbar/v3"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Urlstat struct {
	url      string
	title    string
	statcode int
	url_ip   string
}

func url_request(url string, pproxy string, sstime int) (title string, stacode int, err error) {
	tmp_url := url
	header := req.Header{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36 Edg/92.0.902.73",
	}
	if pproxy != "" {
		req.SetProxyUrl(pproxy)
	}
	req.SetTimeout(time.Duration(sstime) * time.Second)
	r, err := req.Get(tmp_url, header)
	if err != nil {
		var s = [2]int{0, 0}
		return string(s[1]), 0, err
	} else {
		resp := r.Response()
		data := string(r.Bytes())
		sta_code := resp.StatusCode
		re := regexp.MustCompile("<title>(.+)</title>")
		if re.MatchString(data) {
			s := re.FindStringSubmatch(data)
			return string(s[1]), sta_code, err
		} else {
			var s = [2]int{0, 0}
			return string(s[1]), sta_code, err
		}
	}
}

func file_operation(filepath string) []string {
	file, _ := os.Open(filepath)

	scanner := bufio.NewScanner(file) // 接受io.Reader类型参数 返回一个bufio.Scanner实例
	var ss []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "http") { //前缀是http开头
			ss = append(ss, line)
		} else {
			line = "http://" + line
			ss = append(ss, line)
		}
	}
	file.Close()
	return ss
}

func writerCSV(path string, totsl []Urlstat) {

	//OpenFile读取文件，不存在时则创建，使用追加模式
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("文件打开失败！")
	}
	defer file.Close()
	file.WriteString("\xEF\xBB\xBF")
	//创建写入接口
	file.WriteString("目标URL,目标Title,响应状态码,IP\n")
	//写入一条数据，传入数据为切片(追加模式)
	for i := 0; i < len(totsl); i++ {
		line := totsl[i].url + "," + totsl[i].title + "," + strconv.Itoa(totsl[i].statcode) + "," + totsl[i].url_ip + "\n"
		file.WriteString(line)
	}

	log.Println("\n数据写入完成...\n")
}

func data_processing(routineCT int, ss []string, pproxy string, sstime int) []Urlstat {
	g := golimit.NewG(routineCT) // 创建go程
	wg := &sync.WaitGroup{}
	bar := progressbar.Default(int64(len(ss))) // 设置进度条
	// 开始访问目标
	var totalsl []Urlstat
	for i := 0; i < len(ss); i++ {
		wg.Add(1)
		task := ss[i]
		g.Run(func() {
			defer func() { // 当go程崩溃时触发
				if err := recover(); err != nil {
					wg.Done()
					fmt.Println(err)
				}
			}()
			title, stacode, err := url_request(task, pproxy, sstime)
			if err != nil {
				// fmt.Printf("\nerror : %s无法访问\n", task)
				var tmpn Urlstat
				tmpn.url = task
				tmpn.title = "无法访问"
				tmpn.statcode = 0
				totalsl = append(totalsl, tmpn)
			} else {
				var tmpn Urlstat
				tmpn.url = task
				tmpn.title = title
				tmpn.statcode = stacode
				if strings.HasPrefix(task, "https") {
					task = string([]byte(task)[8:])
				} else {
					task = string([]byte(task)[7:])
				}
				addr, err := net.ResolveIPAddr("ip", task)
				if err != nil {
					fmt.Println(err)
					tmpn.url_ip = "NULL"
					totalsl = append(totalsl, tmpn)
				} else {
					tmpn.url_ip = addr.String()
					totalsl = append(totalsl, tmpn)
				}
			}
			wg.Done()
			bar.Add(1)
		})
	}
	wg.Wait()
	return totalsl
}

func set_flag() (string, int, int, string, string) {
	tmpcsv := strconv.Itoa(int(time.Now().Unix())) + ".csv"
	var routineCountTotal int
	var filepath string
	var csvpath string
	var pproxy string
	var stime int
	flag.StringVar(&filepath, "r", "", "传入待测试地址文件,默认为空")
	flag.IntVar(&routineCountTotal, "g", 3, "线程数")
	flag.IntVar(&stime, "t", 5, "设置访问超时时长")
	flag.StringVar(&csvpath, "o", tmpcsv, "传入生成的csv文件的地址,默认为当前路径")
	flag.StringVar(&pproxy, "p", "", "设置代理地址,默认为空;例：http://127.0.0.1:10809")
	flag.Parse()
	return filepath, routineCountTotal, stime, csvpath, pproxy
}

func main() {
	// 设置flag
	filepath, routineCountTotal, stime, csvpath, pproxy := set_flag()
	// 处理URL
	ss := file_operation(filepath)
	// 创建go程并处理数据
	totalsl := data_processing(routineCountTotal, ss, pproxy, stime)
	// fmt.Println(totalsl[1])
	writerCSV(csvpath, totalsl)
}
