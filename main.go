package main

import (
	"strconv"
	"strings"
	"io"
	"os"
	"fmt"
	"net/http"
)

func main() {
	url := os.Args[1]

	splitedURL := strings.Split(url, "/")
	fileName := splitedURL[len(splitedURL) - 1]
	fileNamePart := fileName + ".part"

	// part file exist or not
	var filePartEixst bool
	fileInfo, err := os.Stat(fileNamePart)
	if err != nil {
		filePartEixst = false
	} else {
		filePartEixst = !fileInfo.IsDir()
	}

	// calculate offset 
	var offset int64
	if filePartEixst {
		offset = fileInfo.Size()
	} else {
		offset = 0
	}

	// build get request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Add("RANGE", "bytes=" + strconv.FormatInt(offset, 10) + "-")

	// get response
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		fmt.Println("http code:" + resp.Status)
		os.Exit(1)
	}

	// save(append) response body to file
	var file *os.File
	if !filePartEixst {
		file, err = os.Create(fileNamePart)
	} else {
		file, err =  os.OpenFile(fileNamePart, os.O_RDWR|os.O_APPEND, 0660)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		file.Close()
		os.Exit(1)
	}
	file.Close()

	// rename part file to file
	err = os.Rename(fileNamePart, fileName)
	if err != nil {
		fmt.Println("rename error")
		os.Exit(1)
	}
}
