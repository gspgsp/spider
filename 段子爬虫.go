package main

import (
	"fmt"
	"strconv"
	"net/http"
	"regexp"
	"strings"
	"os"
)

/**
发送请求
 */
func HttpGet(url string) (result string, err error)  {
	resp, err1 := http.Get(url)
	if err1 != nil{
		err = err1
		return
	}

	defer resp.Body.Close()

	buf := make([]byte, 4 * 1024)
	for{
		n, _ := resp.Body.Read(buf)
		if n == 0{
			break
		}
		result += string(buf[:n])
	}
	return
}

//存到文件
func storeToFile(i int , fileTitle, fileContent []string)  {
	//新建文件
	f, err := os.Create(strconv.Itoa(i)+".txt")
	if err != nil{
		fmt.Println("os.Create err = ", err)
		return
	}
	defer f.Close()

	//写入
	n :=len(fileTitle)
	for i := 0; i < n; i++{
		//写标题
		f.WriteString(fileTitle[i]+"\n")
		//写内容
		f.WriteString(fileContent[i]+"\n")
		f.WriteString("\n***********************************\n")
	}

}

/**
爬取单独页面
 */
func SpiderOneJoy(url string) (title, content string, err error) {

	result, err1 := HttpGet(url)
	if err1 != nil{
		err = err1
		return
	}

	//拿到title
	re1 := regexp.MustCompile(`<h1>(?s:(.*?))</h1>`)
	if re1 == nil{
		fmt.Println("regexp.MustCompile err")
		err = fmt.Errorf("%s", "regexp.MustCompile err")
		return
	}
	tmpTitle := re1.FindAllStringSubmatch(result, 1)
	for _, data := range tmpTitle{
		title = data[1]
		title = strings.Replace(title, " ", "", -1)
		title = strings.Replace(title, "\t", "", -1)
		break
	}

	//拿到content
	re2 := regexp.MustCompile(`<div class="content-txt pt10">(?s:(.*?))<a id="prev"`)
	if re2 == nil{
		fmt.Println("regexp.MustCompile err")
		err = fmt.Errorf("%s", "regexp.MustCompile err")
		return
	}
	tmpContent := re2.FindAllStringSubmatch(result, -1)
	for _, data := range tmpContent{
		content = data[1]
		content = strings.Replace(content, "\r", "", -1)
		content = strings.Replace(content, "\n", "", -1)
		content = strings.Replace(content, "\t", "", -1)
		content = strings.Replace(content, "&nbsp;", "", -1)
		content = strings.Replace(content, "<br>", "", -1)
		content = strings.Replace(content, "<br />", "", -1)
		content = strings.Replace(content, "<br><br />", "", -1)

		/*re3 := regexp.MustCompile(`<img oldsrc="(?s:(.*?))" width="510"`)
		if re3 != nil{
			content = re3.FindAllStringSubmatch(content, -1)
		}*/
		break
	}

	return
}

/**
爬取页面
 */
func SpiderPage(i int, page chan int) {
	//https://www.pengfu.com/index_1.html
	url := "https://www.pengfu.com/index_"+strconv.Itoa(i)+".html"
	fmt.Printf("正在爬取第%d个页面，网址为:%s\n",i, url)

	result, err := HttpGet(url)
	if err != nil{
		fmt.Println("HttpGet err = ", err)
		return
	}
	//fmt.Println("r = ", result)

	re := regexp.MustCompile(`<h1 class="dp-b"><a href="(?s:(.*?))"`)
	if re == nil{
		fmt.Println("regexp.MustCompile err")
		return
	}

	//取连接里的详细信息
	joyUrls := re.FindAllStringSubmatch(result,-1)
	//fmt.Println("joyUrls = ", joyUrls)

	//
	fileTitle := make([]string, 0)
	fileContent := make([]string, 0)
	//
	for _, data := range joyUrls{
		//fmt.Println("url = ", data[1])
		//根据每个连接 爬取到内容
		title, content, err := SpiderOneJoy(data[1])
		if err != nil{
			fmt.Println("SpiderOneJoy err", err)
			return
		}

		/*fmt.Println("the title is :", title)
		fmt.Println("the content is :", content)*/
		fileTitle = append(fileTitle, title)
		fileContent = append(fileContent, content)
	}

	/*fmt.Println("fileTitle = ", fileTitle)
	fmt.Println("fileContent = ", fileContent)*/
	storeToFile(i, fileTitle, fileContent)

	//每写完一页 就给管道一个当前页码的值
	page <- i

}

/**
处理函数
 */
func DoWork(start, end int)  {
	fmt.Printf("准备爬去第%d页到第%d的网址\n",start,end)

	page := make(chan int)


	for i := start; i <= end; i++ {
		go SpiderPage(i, page)
	}

	for i := start; i <= end; i++ {
		fmt.Printf("第%d个页面爬取完成\n", <-page)
	}
}

func main(){
	var start, end int
	fmt.Println("请输入起始页(>=1):")
	fmt.Scan(&start)
	fmt.Println("请输入终止页(>=起始页):")
	fmt.Scan(&end)

	//工作函数
	DoWork(start, end)
}