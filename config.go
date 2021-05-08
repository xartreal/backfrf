// config
package main

import (
	"strconv"

	"github.com/vaughan0/go-ini"
)

var myversion = "0.9.3"
var useragent = "ARL backfrf/" + myversion

var RunCfg struct {
	jsononly bool
	feedpath string
	feedname string
	extflag  bool //true if mediafilter empty
	// fast path
	jpath     string
	timeline  string
	list      string
	mediapath string
}

var Config struct {
	myauth    string
	myname    string
	step      int
	logstat   int
	loadmedia int
	debugmode int
	allhtml   int
	maxlast   int
	archive   int
	filter    string
}

func getIniVar(file ini.File, section string, name string) string {
	rt, _ := file.Get(section, name)
	if len(rt) < 1 {
		outerror(2, "FATAL: Variable '%s' not defined\n", name)
	}
	return rt
}

func getIniNum(file ini.File, section string, name string) int {
	rt, err := strconv.Atoi(getIniVar(file, section, name))
	if err != nil {
		return 0
	}
	return rt
}

func ReadConf() {
	file, err := ini.LoadFile("backfrf.ini")
	if err != nil {
		outerror(1, "\n! Configuration not found\n")
	}
	Config.myauth = getIniVar(file, "credentials", "auth")
	Config.myname = getIniVar(file, "credentials", "myname")

	Config.step = getIniNum(file, "default", "step")
	Config.debugmode = getIniNum(file, "default", "debug")

	Config.logstat = getIniNum(file, "user", "logstat")
	Config.loadmedia = getIniNum(file, "user", "loadmedia")
	Config.allhtml = getIniNum(file, "user", "allhtml")
	Config.maxlast = getIniNum(file, "user", "maxlast")
	Config.archive = getIniNum(file, "user", "archive")
	Config.filter, _ = file.Get("user", "filter")
}
