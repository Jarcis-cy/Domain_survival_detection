# Domain_survival_detection
### 域名存活检测
初学go语言，通过golang实现对大批量的域名进行存活检测，获取目标title，ip，cms等。探测cms是使用的[boy-hack](https://github.com/boy-hack)大佬的[goWhatweb](https://github.com/boy-hack/goWhatweb)实现的，后续该项目还会添加CDN探测等功能。

### 使用方法
`go run main.go -h` or `Uscan.exe -h`
```
  -c    当添加该参数时，启动cms识别功能
  -d    当添加该参数时，启动cdn识别功能
  -g int
        线程数 (default 3)
  -o string
        传入生成的csv文件的地址,默认为当前路径 (default "1631862565.csv")
  -r string
        传入待测试地址文件,默认为空
```
##### Example
`go run main.go -r text.txt -c -d`

##### 注意
需要将Uscan的可执行程序与cms.json,waf.txt置于同一目录下

----

### 更新
##### 2021/9/17 添加CDN识别功能
通过访问获取ip的网站，多地ping的形式进行CDN存在与否的识别，后续添加其他网站请求资源。
