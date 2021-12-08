package main

import (
	"GoUtils/download_utils"
	"fmt"
)

func main() {
	var url string = "https://dl.softmgr.qq.com/original/im/QQ9.5.0.27852.exe"
	var download_folder string = "./download_folder/"
	// var httpDownload *download_utils.HttpDownloader

	httpDownload := download_utils.NewDownloader(url, 4)
	fmt.Printf("Bool:%v\nContent:%d\n", httpDownload.GetAcceptRanges(), httpDownload.GetContentLength())

	httpDownload.Download(download_folder)
}
