// funcs
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func httpget(url string, withtoken bool) []byte {
	var query = []byte("")
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(query))
	if withtoken {
		req.Header.Set("X-Authentication-Token", Config.myauth)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	req.Header.Set("User-Agent", useragent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return []byte{0}
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func httpfile(url string) []byte {
	var query = []byte("")
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(query))
	//	req.Header.Set("X-Authentication-Token", auth)
	req.Header.Set("User-Agent", useragent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err)
		return []byte{0}
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func MkFeedPath(feedname string) {
	feedpath := ""
	if strings.Contains(feedname, "filter:") {
		feedpath = strings.Replace(feedname, ":", "/", -1)
	} else {
		feedpath = "feeds/" + feedname
	}
	feedpath = strings.Replace(feedpath, ":", "-", -1) //todo: subfeeds?
	RunCfg.feedpath = feedpath + "/"
}

func loadtfile(name string) string {
	fbin, err := ioutil.ReadFile(name)
	if err != nil {
		outerror(2, "FATAL: Template load error: '%s'\n", name)
	}
	return string(fbin)
}

func isexists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func outerror(code int, format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(code)
}
