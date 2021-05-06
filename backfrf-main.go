package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func helpprog() {
	fmt.Print("Usage: \n\n",
		"get feed&html: backfrf feed <feed>\n",
		"get single post: backfrf get <addr>\n",
		"get jsons: backfrf json <feed>\n",
		"get jsons only: backfrf jsononly <feed>\n",
		"rebuild html: backfrf html <feed>\n",
		"check integrity: backfrf check <feed>\n",
		"find lost posts: backfrf lost <feed>\n",
		"reindex timelines: backfrf reindex <feed>\n\n")
	os.Exit(1)
	//	fmt.Printf("init configuration: backfrf init <username> <password>\n")
}

func message(text, feedname string) {
	fmt.Printf(text, feedname)
	if !isexists(RunCfg.feedpath + "json/profile") {
		outerror(1, "Feed '%s' not found\n", feedname)
	}
}

func checkArgs() (string, string) {
	if len(os.Args) != 3 {
		helpprog()
	}
	return os.Args[1], os.Args[2]
}

var TimelineDB KVBase
var ExtDB KVBase

func main() {
	fmt.Println("backfrf " + myversion + "\n")
	// command line parsing
	if len(os.Args) < 3 {
		helpprog()
	}

	initRegs()

	// init command (deprecated)
	if strings.EqualFold(os.Args[1], "init") {
		mktoken(os.Args[2], os.Args[3])
	}
	cmd, feedname := checkArgs()
	ReadConf()
	if feedname == "@myname" {
		feedname = Config.myname
	}

	// commands
	switch strings.ToLower(cmd) {
	case "get": // get command
		getaddr := os.Args[2]
		if strings.HasPrefix(getaddr, "https:/") {
			getaddr = filepath.Base(getaddr)
		}
		//		ReadConf()
		loadtemplates(true)
		// Check auth
		checktoken(true)
		RunCfg.feedpath = "posts/" + getaddr
		if os.MkdirAll(RunCfg.feedpath, 0755) != nil {
			outerror(1, "\nFATAL: Can't create directory %s\n", RunCfg.feedpath)
		}
		fmt.Printf("Get %s ", getaddr)
		getPost(getaddr, true)
		fmt.Printf("\n\n")
		os.Exit(0)
	case "html": // html command
		MkFeedPath(feedname)
		message("Creating html, feed '%s'\n", feedname)
		openDB(RunCfg.feedpath+"db/media.db", "ext", &ExtDB)
		rebuildHtml()
		closeDB(&ExtDB)
		os.Exit(0)
	case "lost": //find lost posts
		MkFeedPath(feedname)
		message("Checking feed '%s'\n", feedname)
		findlost()
		os.Exit(0)
	case "check": //check integrity command
		Config.debugmode = 1
		MkFeedPath(feedname)
		message("Checking feed '%s'\n", feedname)
		if !isexists(RunCfg.feedpath + "db/media.db") {
			createDB("db/media.db", "ext", &ExtDB)
		}
		checkfeed()
		os.Exit(0)
	case "reindex": // reindex lists
		MkFeedPath(feedname)
		message("Rebuilding timeline lists in feed [%s]\n", feedname)
		rebuildLists()
		os.Exit(0)
	default:
	}

	// backup mode commands
	switch strings.ToLower(cmd) {
	case "json": // json command
		fmt.Printf("Get jsons for feed '%s'\n", feedname)
	case "jsononly": // json-only command
		fmt.Printf("Get jsons only for feed '%s'\n", feedname)
		RunCfg.jsononly = true
	case "feed":
		fmt.Printf("Get data for feed '%s'\n", feedname)
	default:
		fmt.Printf("Unknown command '%s'\n", cmd)
		os.Exit(1)
	}

	backup(feedname)
	if strings.ToLower(cmd) == "feed" { // index json & make html
		rebuildHtml()
	}
	fmt.Printf("Done\n")
}