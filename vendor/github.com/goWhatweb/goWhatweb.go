package goWhatweb

import (
	"fmt"
	"github.com/goWhatweb/engine"
	"github.com/goWhatweb/until"
	"log"
	"regexp"
	"sync"
	"time"
)

var ()

type url_cms struct {
	Domain string
	Cms    string
}

func Gww(domains []string) []url_cms {
	//domains := []string{"https://www.hacking8.com", "https://x.hacking8.com"}

	// 加载指纹
	sortPairs, webdata := until.ParseCmsDataFromFile("cms.json")
	var wg sync.WaitGroup

	// 开始并发相关
	ResultChian := make(chan string)
	fmt.Println("Load url:", domains)
	for _, domain := range domains {
		go func(d string) {
			newWorker := engine.NewWorker(7, d, &wg, ResultChian)
			if !newWorker.Checkout() {
				return
			}
			newWorker.Start()
			for _, v := range sortPairs {
				tmp_job := engine.JobStruct{d, v.Path, webdata[v.Path]}
				//fmt.Println(tmp_job)
				newWorker.Add(tmp_job)
			}
		}(domain)
	}
	time.Sleep(time.Second * 1)
	var uc_list []url_cms
	go func() {
		for {
			r := <-ResultChian
			re1 := regexp.MustCompile("Domain:(.+?) Cms:(.+?) Path")
			if re1.MatchString(r) {
				var temp url_cms
				s := re1.FindStringSubmatch(r)
				temp.Domain = string(s[1])
				temp.Cms = string(s[2])
				uc_list = append(uc_list, temp)
				// fmt.Println(string(s[1]) + "\n" + string(s[2]) + "\n完成一次循环")
				log.Println(r)
			} else {
				log.Println(r)
			}
			// fmt.Printf("%T", r)

		}
	}()
	// log.Println("初始化完成")
	wg.Wait()
	return uc_list
}
