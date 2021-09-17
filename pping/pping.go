package pping

import(
	"fmt"
	"net/http"
	"crypto/tls"
	"time"
	"io/ioutil"
	"math/rand"
	"strings"
	"regexp"
	"github.com/widuu/gojson"
	"Domain_survival_detection/golimit"
	"sync"
)

type url_cdn struct {
	Domain string
	Cdn    string
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

func Post(url_a string, datat string) (content []byte, header http.Header, statcode int, err error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
	}
	Client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transCfg,
	}

	req, err := http.NewRequest("POST", url_a, strings.NewReader(datat))
	req.Header.Set("Content-Type","application/x-www-form-urlencoded; charset=UTF-8")
	if err != nil {
		return nil, nil, 0, err
	}
	req.Header.Add("User-Agent", get_random_ua())
	resp, err2 := Client.Do(req)
	if err2 != nil {
		return nil, nil, 0, err
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	return bytes, resp.Header, resp.StatusCode, nil
}

func isValueInList(value string, list []string) bool {
    for _, v := range list {
        if v == value {
            return true
        }
    }
    return false
}
func Pping(Pp_url []string) []url_cdn {
	base_url := "https://www.wepcc.com/"
	query_url := "https://www.wepcc.com/check-ping.html"
	// add_url := "www.baidu.com"
	// all_url := base_url + add_url
	var ucdn_list []url_cdn
	for _, domain := range Pp_url {
		ping_url := domain
		datat := "host=" + ping_url + "&node=1%2C2%2C3%2C4%2C5%2C6"
		data, _, _, _ := Post(base_url, datat)
		addat := string(data)
		re := regexp.MustCompile("id=\".+?\"")

		// 创建go程
		g := golimit.NewG(10) 
		wg := &sync.WaitGroup{}

		var ip_list []string		
		if re.MatchString(addat) {
			s := re.FindAllString(addat,15)
			for i := 0; i < len(s)-1; i++ {
				wg.Add(1)
				g.Run(func() {
					defer func() { // 当go程崩溃时触发
						if err := recover(); err != nil {
						wg.Done()
						fmt.Println(err)
						}
					}()
					node_id := string([]byte(s[i])[4:36])
					check_data_ping := "node=" + node_id + "&host=" + ping_url
					data, _, _, _ := Post(query_url, check_data_ping)
					ip_data := string(data)
					ip := gojson.Json(ip_data).Get("data").Get("Ip").Tostring()
					if isValueInList(ip,ip_list)==false {
						ip_list = append(ip_list, ip)
					}
					wg.Done()
				})				 			
			}
			wg.Wait()
		}
		var temp url_cdn
		if len(ip_list) > 1 {
			temp.Domain = domain
			temp.Cdn = "存在"
		} else {
			temp.Domain = domain
			temp.Cdn = "不存在"
		}
		ucdn_list = append(ucdn_list, temp)
	}
	return ucdn_list
}