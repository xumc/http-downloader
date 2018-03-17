package main

import (
	"strconv"
	"strings"
	"io"
	"os"
	"fmt"
	"net/http"
	"errors"
)

const (
	threadCount int64 = 5
)

type filePart struct {
	fileData []byte
	start int64
}
func createFileByURL(url string) (*os.File, int64, error) {
	splitedURL := strings.Split(url, "/")
	fileName := splitedURL[len(splitedURL) - 1]

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New("http code returned:" + strconv.Itoa(resp.StatusCode))
	}

	file, err := os.Create(fileName)
	if err != nil {
		return nil, 0, err
	}

	err = file.Truncate(resp.ContentLength)
	if err != nil {
		return nil, 0, err
	}

	return file, resp.ContentLength, nil
}

func downloadPart(url string, fileChan chan filePart, start, end int64) {
	fmt.Println(start, " => ", end)
	// build get request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Add("RANGE", "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end, 10))

	// get response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		fmt.Println("http code:" + resp.Status)
		os.Exit(1)
	}

	var data []byte
	for {
		buffer := make([]byte, 1024)
		nr, err := resp.Body.Read(buffer)
		if nr > 0 {
			data = append(data, buffer[:nr]...)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	part := filePart{fileData: data, start: start}
	fileChan <- part
}

func main() {
	url := os.Args[1]

	file, fileSize, err := createFileByURL(url)
	if err != nil {
		fmt.Println("created file error ", err)
		os.Exit(1)
	}

	fileChan := make(chan filePart)

	var block int64
	if fileSize % threadCount == 0 {
		block = fileSize / threadCount  
	} else {
		block = fileSize / threadCount + 1
	} 

	for i := int64(0); i < threadCount; i++ {
    	start := i * block;  
		var end int64
		if start + block >= fileSize {
			end = fileSize
		} else {
			end = start + block - 1
		}

		go downloadPart(url, fileChan, start, end)
	}
	
	for i := threadCount; i > 0; i-- {
		select {
		case part := <- fileChan:
			file.Seek(part.start, 0)
			file.Write(part.fileData)
		}
	}

	if err := file.Close(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	

	// fileNamePart := fileName + ".part"

	// // part file exist or not
	// var filePartEixst bool
	// fileInfo, err := os.Stat(fileNamePart)
	// if err != nil {
	// 	filePartEixst = false
	// } else {
	// 	filePartEixst = !fileInfo.IsDir()
	// }

	// // calculate offset 
	// var offset int64
	// if filePartEixst {
	// 	offset = fileInfo.Size()
	// } else {
	// 	offset = 0
	// }

	// // build get request
	// req, err := http.NewRequest("GET", url, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// req.Header.Add("RANGE", "bytes=" + strconv.FormatInt(offset, 10) + "-")

	// // get response
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// defer resp.Body.Close()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
	// 	fmt.Println("http code:" + resp.Status)
	// 	os.Exit(1)
	// }

	// // save(append) response body to file
	// var file *os.File
	// if !filePartEixst {
	// 	file, err = os.Create(fileNamePart)
	// } else {
	// 	file, err =  os.OpenFile(fileNamePart, os.O_RDWR|os.O_APPEND, 0660)
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// _, err = io.Copy(file, resp.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	file.Close()
	// 	os.Exit(1)
	// }
	// file.Close()

	// // rename part file to file
	// err = os.Rename(fileNamePart, fileName)
	// if err != nil {
	// 	fmt.Println("rename error")
	// 	os.Exit(1)
	// }
}
