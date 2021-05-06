// services
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	regJsReplace *regexp.Regexp
	regToken     *regexp.Regexp
)

func initRegs() {
	regJsReplace = regexp.MustCompile(`"([a-zA-Z]{2,})":(\d{1,})`)
	regToken = regexp.MustCompile(`"authToken":"(.*)"`)
}

func checkIsFeed(feedname string) bool {
	if strings.Contains(feedname, "filter:") || (feedname == "home") {
		return true
	}
	res := httpget("https://freefeed.net/v1/users/"+feedname, false)
	if strings.Contains(string(res), `{"err"`) {
		return false
	}
	//	fmt.Printf("%v\n", string(res))
	return strings.Contains(string(res), feedname)
}

func MkFeedItems(feedname string) {
	if !isexists(RunCfg.feedpath) {
		fmt.Printf("Creating directory for feed [%s]\n", feedname)
		if os.MkdirAll(RunCfg.feedpath+"html", 0755) != nil {
			outerror(1, "\n! Can't create feed directory\n")
		}
		os.MkdirAll(RunCfg.feedpath+"json", 0755)
		os.MkdirAll(RunCfg.feedpath+"db", 0755)
		os.MkdirAll(RunCfg.feedpath+"index", 0755)    //list_*
		os.MkdirAll(RunCfg.feedpath+"media", 0755)    // image_*, media_*
		os.MkdirAll(RunCfg.feedpath+"timeline", 0755) //timeline_*
	}
	// create timeline db (if not exists) & open
	dbname := RunCfg.feedpath + "db/timeline.db"
	if !isexists(dbname) {
		createDB(dbname, "posts", &TimelineDB)
	}
	openDB(dbname, "posts", &TimelineDB)
	// create ext db (if not exists) & open
	dbname = RunCfg.feedpath + "db/media.db"
	if !isexists(dbname) {
		createDB(dbname, "ext", &ExtDB)
	}
	openDB(dbname, "ext", &ExtDB)
}

func backup(feedname string) {
	/*	if feedname == "@myname" {
		feedname = Config.myname
	}*/
	if !checkIsFeed(feedname) {
		outerror(2, "ERROR: Feed '%s' not found or not available\n", feedname)
	}
	MkFeedPath(feedname)
	RunCfg.feedname = feedname
	feedname = strings.Replace(feedname, ":", "/", -1)

	// make structure
	MkFeedItems(feedname)

	checktoken(false) // Check auth

	fmt.Printf("\n")

	MyStat = TMyStat{changedrecords: 0, newimages: 0, newrecords: 0, records: 0}

	tmleof, offset := 1, 0
	RunCfg.extflag = len(Config.filter) == 0

	for tmleof > 0 {
		if (Config.maxlast > 0) && (offset > Config.maxlast) {
			break
		}
		tmleof = mkTimeline(feedname).getTimeline(offset).processTimeline()
		offset += Config.step
	}
	// save new record list (for z-indexing)
	itemslist := strings.Join(newitems, "\n") + "\n"
	ioutil.WriteFile(RunCfg.feedpath+"db/newitems", []byte(itemslist), 0644)

	// close timeline & media db
	closeDB(&ExtDB)
	closeDB(&TimelineDB)
	//stat
	outstat := fmt.Sprintf("\nTotal: %d records, new: %d, changed: %d\n%d new media files\n",
		MyStat.records, MyStat.newrecords, MyStat.changedrecords, MyStat.newimages)
	fmt.Println(outstat)
	if Config.logstat == 1 {
		timestat := time.Now().Format(time.RFC822)
		f, _ := os.OpenFile(RunCfg.feedpath+"stat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString("\n" + timestat + "\n" + outstat)
		f.Close()
	}
}

func rebuildHtml() {
	hstart := 0
	maxeof := Config.step
	loadtemplates(false)
	//fmt.Printf("%d - %d -%d\n", hstart, step, maxeof)
	fmt.Printf("\nScanning...")
	for isexists(RunCfg.feedpath + "timeline/timeline_" + strconv.Itoa(hstart)) {
		hstart += Config.step
		//fmt.Printf("s: %d\n", hstart)
	}
	if hstart > Config.step {
		maxeof = hstart - Config.step
	}
	fmt.Printf(" %d records \n", maxeof-Config.step)
	fmt.Printf("HTMLing...\n")

	for i := 0; i < maxeof; i += Config.step {
		genhtml(i, maxeof)
	}
	if !isexists(RunCfg.feedpath + "html/kube.min.css") {
		fbin, _ := ioutil.ReadFile("template/kube.min.css")
		ioutil.WriteFile(RunCfg.feedpath+"html/kube.min.css", fbin, 0644)
	}
	fmt.Printf("\n")
}

// special
func checkTimeline(offset int) (int, int) {
	errcnt := 0
	text, _ := ioutil.ReadFile(RunCfg.feedpath + "timeline/timeline_" + strconv.Itoa(offset))
	frf := new(FrFjtml)
	if nlflag {
		fmt.Printf("\n")
	}
	fmt.Printf("\roffset: %d", offset)
	nlflag = false
	json.Unmarshal(text, frf)
	frflen := 0
	//check "posts" files
	jpath := RunCfg.feedpath + "json/posts_"
	for _, p := range frf.Timelines.Posts {
		if !isexists(jpath + p) {
			if Config.debugmode == 1 {
				fmt.Printf("\npost: %s", p)
			}
			getPost(p, false)
			nlflag = true
			errcnt++
		}
		frflen++
	}
	frfz := new(FrFfile)
	json.Unmarshal(text, frfz)
	//	frfzlen := 0
	mtype := ""
	mpath := RunCfg.feedpath + "media/"
	for _, p := range frfz.Attachments {
		if strings.EqualFold(p.MediaType, "image") {
			mtype = "image_"
		} else {
			mtype = "media_"
		}
		if !isexists(mpath + mtype + p.Id + path.Ext(p.Url)) {
			//fmt.Printf(" image/media: %s\n", p.Id)
			getFile(p.Id, p.Url, p.MediaType, false)
			nlflag = true
			errcnt++
		}
	}
	MyStat.records += frflen
	return frflen, errcnt
}

func checkfeed() {
	tmleof := 1
	offset, errcnt, errinc := 0, 0, 0
	nlflag = true
	//	fmt.Println("")
	openDB(RunCfg.feedpath+"db/media.db", "ext", &ExtDB)
	for tmleof > 0 {
		tmleof, errinc = checkTimeline(offset)
		offset += Config.step
		errcnt += errinc
	}
	fmt.Println("")
	closeDB(&ExtDB)
	fmt.Printf("\nErrors detected: %d\n", errcnt)
}

var postlist []string
var feedpostlist []string
var lostlist []string
var lostcnt int

func inlist(instr string) bool {
	for _, v := range feedpostlist {
		if v == instr {
			return true
		}
	}
	return false
}

func getfeedlist(offset int) int {
	text, _ := ioutil.ReadFile(RunCfg.feedpath + "timeline/timeline_" + strconv.Itoa(offset))
	frf := new(FrFjtml)
	fmt.Printf("offset: %d\r", offset)
	json.Unmarshal(text, frf)

	frflen := 0
	//check "posts" files
	jpath := RunCfg.feedpath + "json/posts_"
	for _, p := range frf.Timelines.Posts {
		feedpostlist = append(feedpostlist, jpath+p)
		frflen++
	}
	return frflen
}

func findlost() {
	var outtext string
	postlist, _ = filepath.Glob(RunCfg.feedpath + "json/posts_*")
	fmt.Printf("Posts in feed directory: %d\n", len(postlist))
	lostcnt = 0
	tmleof, offset := 1, 0
	for tmleof > 0 {
		tmleof = getfeedlist(offset)
		offset += Config.step
	}
	fmt.Printf("Posts in timelines: %d\n", len(feedpostlist))
	for _, itm := range postlist {
		if !inlist(itm) {
			lostcnt++
			lostlist = append(lostlist, itm)
		}
	}
	if lostcnt > 0 {
		fmt.Printf("------------\nFound %d lost posts\n", lostcnt)
		outtext = fmt.Sprintf("Found %d lost posts\n", lostcnt)
		for _, st := range lostlist {
			outtext += fmt.Sprintf("%s\n", st)
		}
		ioutil.WriteFile("lost", []byte(outtext), 0644)
		fmt.Printf("Created file 'lost'\n")
	} else {
		fmt.Printf("No lost posts found\n")
	}
}

func rebuildLists() {
	type kv struct {
		Id    string
		Value int64
	}
	var RecList = []kv{}
	postlist, _ = filepath.Glob(RunCfg.feedpath + "json/posts_*")
	// get timemarks
	for i := 0; i < len(postlist); i++ {
		fbin, _ := ioutil.ReadFile(postlist[i])
		frf := new(FrFJSON)
		json.Unmarshal(fbin, frf)
		tm, _ := strconv.ParseInt(frf.Posts.UpdatedAt, 10, 64)
		RecList = append(RecList, kv{frf.Posts.Id, tm})
	}
	// sort timemarks
	sort.Slice(RecList, func(i, j int) bool {
		return RecList[i].Value > RecList[j].Value
	})
	// make new lists
	maxcnt := len(RecList)
	cc := 0
	step := 0
	lastoffset := 0
	lpath := RunCfg.feedpath + "index/list_"
	for cc < maxcnt {
		nlist := []string{}
		lc := 0
		for lc < Config.step {
			nlist = append(nlist, RecList[cc].Id)
			lc++
			cc++
			if cc == maxcnt {
				break
			}
		}
		lastoffset = step
		ioutil.WriteFile(lpath+strconv.Itoa(step), []byte(strings.Join(nlist, "\n")), 0644)
		step += Config.step
	}
	fmt.Printf("Posts processed: %d, last offset: %d\n", maxcnt, lastoffset)
}
