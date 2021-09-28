package main

import (
	"Domain_survival_detection/golimit"
	"Domain_survival_detection/pping"
	"bufio"
	"flag"
	"fmt"
	"Domain_survival_detection/goWhatweb"
	"Domain_survival_detection/goWhatweb/fetch"
	"github.com/schollz/progressbar/v3"
	"log"
	"math/rand"
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
	cms      string
	cdn      string
}

type url_cms struct {
	Domain string
	Cms    string
}

type url_cdn struct {
	Domain string
	Cdn    string
}


func url_request(url string) (string, int, error) {
	data, _, sta_code, err := fetch.Get(url)
	if err != nil {
		var s = [2]int{0, 0}
		return string(s[1]), 0, err
	} else {
		data := string(data)
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
			linea := "http://" + line
			lines := "https://" + line
			ss = append(ss, linea)
			ss = append(ss, lines)
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
	file.WriteString("目标URL,目标Title,响应状态码,IP,CMS,CDN\n")
	//写入一条数据，传入数据为切片(追加模式)
	for i := 0; i < len(totsl); i++ {
		line := totsl[i].url + "," + totsl[i].title + "," + strconv.Itoa(totsl[i].statcode) + "," + totsl[i].url_ip + "," + totsl[i].cms + "," + totsl[i].cdn + "\n"
		file.WriteString(line)
	}

	log.Println("\n数据写入完成...\n")
}

func data_processing(routineCT int, ss []string) []Urlstat {
	g := golimit.NewG(routineCT) // 创建go程
	wg := &sync.WaitGroup{}
	bar := progressbar.Default(int64(len(ss))) // 设置进度条
	// 开始访问目标
	var totalsl []Urlstat
	log.Println("开始访问目标......")
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
			title, stacode, err := url_request(task)
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
					// fmt.Println(err)
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

func get_random_ua() string {
	USER_AGENTS := []string{"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; AcooBrowser; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Acoo Browser; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)",
		"Mozilla/4.0 (compatible; MSIE 7.0; AOL 9.5; AOLBuild 4337.35; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN) AppleWebKit/523.15 (KHTML, like Gecko, Safari/419.3) Arora/0.3 (Change: 287 c9dfb30)",
		"Mozilla/5.0 (X11; U; Linux; en-US) AppleWebKit/527+ (KHTML, like Gecko, Safari/419.3) Arora/0.6",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.2pre) Gecko/20070215 K-Ninja/2.1.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN; rv:1.9) Gecko/20080705 Firefox/3.0 Kapiko/3.0",
		"Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5",
		"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.20 (KHTML, like Gecko) Chrome/19.0.1036.7 Safari/535.20",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52"}
	length := len(USER_AGENTS)
	index := rand.Intn(length)
	return USER_AGENTS[index]
}

func set_flag() (string, int, string, bool, bool) {
	tmpcsv := strconv.Itoa(int(time.Now().Unix())) + ".csv"
	var routineCountTotal int
	var filepath string
	var csvpath string
	var cmstf bool
	var cdntf bool
	flag.StringVar(&filepath, "r", "", "传入待测试地址文件,默认为空")
	flag.IntVar(&routineCountTotal, "g", 3, "线程数")
	flag.StringVar(&csvpath, "o", tmpcsv, "传入生成的csv文件的地址,默认为当前路径")
	flag.BoolVar(&cmstf, "c", false, "当添加该参数时，启动cms识别功能")
	flag.BoolVar(&cdntf, "d", false, "当添加该参数时，启动cdn识别功能")
	flag.Parse()
	return filepath, routineCountTotal, csvpath, cmstf, cdntf
}

func main() {
	// 设置flag
	filepath, routineCountTotal, csvpath, cmstf, cdntf := set_flag()
	// 处理URL
	if filepath == "" {
		fmt.Printf("请添加参数r，并添加要检测的url文本")
	} else {
		ss := file_operation(filepath)
		// 创建go程并处理数据
		totalsl := data_processing(routineCountTotal, ss)
		fmt.Println("目标请求完成")
		if cmstf {
			log.Println("开始进行cms探测......")
			uc_list := goWhatweb.Gww(ss)
			for i := 0; i < len(totalsl); i++ {
				for j := 0; j < len(uc_list); j++ {
					if totalsl[i].url == uc_list[j].Domain {
						totalsl[i].cms = uc_list[j].Cms
					}
				}
			}
			log.Println("cms探测完成")
			if cdntf {
				log.Println("开始进行cdn探测......")
				ucdn_list := pping.Pping(ss)
				for i := 0; i < len(totalsl); i++ {
					for j := 0; j < len(ucdn_list); j++ {
						if totalsl[i].url == ucdn_list[j].Domain {
							totalsl[i].cdn = ucdn_list[j].Cdn
						}
					}
				}
				log.Println("cdn探测完成")
			}
		} else {
			if cdntf {
				log.Println("开始进行cdn探测......")
				ucdn_list := pping.Pping(ss)
				for i := 0; i < len(totalsl); i++ {
					for j := 0; j < len(ucdn_list); j++ {
						if totalsl[i].url == ucdn_list[j].Domain {
							totalsl[i].cdn = ucdn_list[j].Cdn
						}
					}
				}
				log.Println("cdn探测完成")
			}
		}
		writerCSV(csvpath, totalsl)
	}
}
