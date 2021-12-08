package download_utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	userAgent = `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36`
	//设置代理
)

type HttpDownloader struct {
	url           string
	filename      string
	contentLength int
	acceptRanges  bool
	// 是否支持断点续传
	numThreads int
	// 同时下载线程数
}

func (h *HttpDownloader) GetAcceptRanges() bool {
	return h.acceptRanges
}
func (h *HttpDownloader) GetContentLength() int {
	return h.contentLength
}

func check(e error) {
	if e != nil {
		log.Println(e)
		panic(e)
	}
}

func NewDownloader(url string, numThreads int) *HttpDownloader {
	var urlSplits []string = strings.Split(url, "/")
	var filename string = urlSplits[len(urlSplits)-1]

	res, err := http.Head(url)
	check(err)

	httpDownload := new(HttpDownloader)
	httpDownload.url = url
	httpDownload.contentLength = int(res.ContentLength)
	httpDownload.numThreads = numThreads
	httpDownload.filename = filename

	if len(res.Header["Accept-Ranges"]) != 0 && res.Header["Accept-Ranges"][0] == "bytes" {
		httpDownload.acceptRanges = true
	} else {
		httpDownload.acceptRanges = false
	}

	return httpDownload
}

// 下载综合调度
func (h *HttpDownloader) Download(folder string) {
	f, err := os.Create(folder + h.filename)
	check(err)
	defer f.Close()

	if h.acceptRanges == false {
		fmt.Println("The file does not support multi-threaded downloading, single-threaded downloading:")
		resp, err := http.Get(h.url)
		check(err)
		save2file(folder, h.filename, 0, resp)
	} else {
		var wg sync.WaitGroup
		for _, ranges := range h.Split() {
			fmt.Printf("Multi-threaded downloading:%d-%d\n", ranges[0], ranges[1])
			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				h.download(start, end, folder)
			}(ranges[0], ranges[1])
		}
		wg.Wait()
	}
}

// 下载文件分段
func (h *HttpDownloader) Split() [][]int {
	ranges := [][]int{}
	blockSize := h.contentLength / h.numThreads
	for i := 0; i < h.numThreads; i++ {
		var start int = i * blockSize
		var end int = (i+1)*blockSize - 1
		if i == h.numThreads-1 {
			end = h.contentLength - 1
		}
		ranges = append(ranges, []int{start, end})
	}
	return ranges
}

// 多线程下载
func (h *HttpDownloader) download(start, end int, folder string) {
	req, err := http.NewRequest("GET", h.url, nil)
	check(err)
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", start, end))
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	check(err)
	defer resp.Body.Close()

	save2file(folder, h.filename, int64(start), resp)
}

// 保存文件
func save2file(folder, filename string, offset int64, resp *http.Response) {
	// os.Mkdir(folder, os.ModePerm)
	saveFilePath := folder + filename
	f, err := os.OpenFile(saveFilePath, os.O_WRONLY, 0660)
	check(err)
	f.Seek(offset, 0)
	defer f.Close()

	content, err := ioutil.ReadAll(resp.Body)
	check(err)
	f.Write(content)
}
