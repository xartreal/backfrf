package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func checktoken(singlemode bool) {
	fmt.Print("Check token...")
	if !whoami(Config.myauth, singlemode) {
		outerror(1, " error!\n")
	}
	fmt.Printf(" ok\n")
}

func whoami(auth string, singlemode bool) bool {
	url := "https://freefeed.net/v2/users/whoami"
	body := httpget(url, true)
	if strings.Contains(string(body), `err":`) { // auth error
		ioutil.WriteFile("error.json", body, 0755)
		return false
	}
	if !singlemode {
		if err := ioutil.WriteFile(RunCfg.feedpath+"json/profile", body, 0755); err != nil {
			outerror(1, "Error: %q\n", err)
		}
	}
	return true
}

func getauth(user string, pass string) (string, bool) {
	resp, _ := http.PostForm("https://freefeed.net/v1/session", url.Values{"username": {user}, "password": {pass}})
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), `err":`) { // auth error
		fmt.Printf("Incorrect username or password\n")
		return "", false
	}
	res := regToken.FindStringSubmatch(string(body))
	if len(res) < 2 {
		outerror(2, "Error: AuthToken not found in server reply\n")
	}
	return res[1], true
}

func mktoken(uname, upass string) {
	if len(os.Args) != 4 {
		helpprog()
	}
	if isexists("backfrf.ini") { //save user section
		ReadConf()
	} else {
		Config.logstat = 0
		Config.loadmedia = 1
	}
	fmt.Printf("Init configuration\nGet token... ")
	myauth, ready := getauth(uname, upass)
	if !ready {
		outerror(1, "error")
	}
	fmt.Println("ok")
	// здесь мы сохраняем auth token
	iniout := "[credentials]\nauth=" + myauth + "\nmyname=" + os.Args[2] + "\n\n"
	iniout += "[default]\nstep=30\ndebug=1\n\n"
	iniout += "[user]\nlogstat=" + strconv.Itoa(Config.logstat) + "\nloadmedia=" + strconv.Itoa(Config.loadmedia)
	iniout += "\nallhtml=0\nmaxlast=0\n"
	ioutil.WriteFile("backfrf.ini", []byte(iniout), 0755)
	fmt.Println("")
	os.Exit(0)
}
