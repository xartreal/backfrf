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
	if strings.Contains(feedname, "filter:") || (feedname == "home") || RunCfg.metafeed {
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
		os.MkdirAll(RunCfg.feedpath+"media", 0755)    // mediafiles
		os.MkdirAll(RunCfg.feedpath+"timeline", 0755) //timeline_*
	}
	// create timeline db (if not exists) & open
	dbname := RunCfg.feedpath + "db/timeline.db"
	if !isexists(dbname) {
		createDB(dbname, "posts", &TimelineDB)
	}
	openDB(dbname, "posts", &TimelineDB)
}

func backup(feedname string) {
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

	// close timeline  db
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
	hstart, maxeof := 0, Config.step
	loadtemplates(false)
	fmt.Printf("\nScanning...")
	for isexists(RunCfg.timeline + strconv.Itoa(hstart)) {
		hstart += Config.step
	}
	if hstart > Config.step {
		maxeof = hstart - Config.step
	}
	fmt.Printf(" last offset: %d\n", maxeof-Config.step)
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

func rebuildMarkdown() {
	hstart, maxeof := 0, Config.step
	if !isexists(RunCfg.feedpath + "md") {
		os.MkdirAll(RunCfg.feedpath+"md", 0755)
	}
	loadmdtemplates(false) //?
	loadTXL("template/template_md.txl")
	checkTXL([]string{"attach", "link", "qlink", "user", "title"})
	fmt.Printf("\nScanning...")
	for isexists(RunCfg.timeline + strconv.Itoa(hstart)) {
		hstart += Config.step
	}
	if hstart > Config.step {
		maxeof = hstart - Config.step
	}
	fmt.Printf(" last offset: %d\n", maxeof-Config.step)
	fmt.Printf("Building MD...\n")

	for i := 0; i < maxeof; i += Config.step {
		genmd(i, maxeof)
	}
	fmt.Printf("\n")
}

// special
func checkTimeline(offset int) (int, int) {
	frflen, errcnt := 0, 0
	text, _ := ioutil.ReadFile(RunCfg.timeline + strconv.Itoa(offset))
	frf := new(FrFjtml)
	if nlflag {
		fmt.Printf("\n")
	}
	fmt.Printf("\roffset: %d", offset)
	nlflag = false
	json.Unmarshal(text, frf)
	// check id from timeline
	for _, p := range frf.Timelines.Posts {
		if !isexists(RunCfg.jpath + p) {
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
	//mtype := ""
	for _, p := range frfz.Attachments {
		/*		if strings.EqualFold(p.MediaType, "image") {
					mtype = "image_"
				} else {
					mtype = "media_"
				}*/
		if !isexists(RunCfg.mediapath + p.Id + path.Ext(p.Url)) {
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
	for tmleof > 0 {
		tmleof, errinc = checkTimeline(offset)
		offset += Config.step
		errcnt += errinc
	}
	fmt.Printf("\n\nErrors detected: %d\n", errcnt)
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
	text, _ := ioutil.ReadFile(RunCfg.timeline + strconv.Itoa(offset))
	frf := new(FrFjtml)
	fmt.Printf("offset: %d\r", offset)
	json.Unmarshal(text, frf)

	frflen := 0
	//check "posts" files
	for _, p := range frf.Timelines.Posts {
		feedpostlist = append(feedpostlist, RunCfg.jpath+p)
		frflen++
	}
	return frflen
}

func findlost() {
	var outtext string
	postlist, _ = filepath.Glob(RunCfg.jpath + "*")
	fmt.Printf("Posts in feed directory: %d\n", len(postlist))
	lostcnt = 0
	tmleof, offset := 1, 0
	for tmleof > 0 {
		tmleof = getfeedlist(offset)
		offset += Config.step
	}
	fmt.Printf("Posts in timelines: %d\n", len(feedpostlist))
	for _, itm := range postlist {
		itm = strings.Replace(itm, `\`, `/`, -1)
		if !inlist(itm) {
			lostcnt++
			lostlist = append(lostlist, itm)
		}
	}
	if lostcnt < 1 {
		fmt.Printf("No lost posts found\n")
		return
	}

	fmt.Printf("------------\nFound %d lost posts\n", lostcnt)
	outtext = fmt.Sprintf("Found %d lost posts\n", lostcnt)
	for _, st := range lostlist {
		outtext += fmt.Sprintf("%s\n", st)
	}
	ioutil.WriteFile("lost.txt", []byte(outtext), 0644)
	fmt.Printf("Created file 'lost.txt'\n")
}

func rebuildLists() {
	type kv struct {
		Id    string
		Value int64
	}
	var RecList = []kv{}
	postlist, _ = filepath.Glob(RunCfg.jpath + "*")
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
	cc, step, lastoffset := 0, 0, 0
	//	lpath := RunCfg.feedpath + "index/list_"
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
		ioutil.WriteFile(RunCfg.list+strconv.Itoa(step), []byte(strings.Join(nlist, "\n")), 0644)
		step += Config.step
	}
	fmt.Printf("Posts processed: %d, last offset: %d\n", maxcnt, lastoffset)
}

func listFeeds(feedname string) {
	if !isexists("feeds") {
		outerror(1, "No feeds\n")
	}
	if feedname != "all" {
		outerror(1, "Incorrect list cmd\n")
	}

	fmt.Printf("Feeds:\n\n")
	files, _ := ioutil.ReadDir("feeds")
	for _, item := range files {
		if !item.IsDir() {
			continue
		}
		flag := ""
		if !isexists("feeds/" + item.Name() + "/json/profile") {
			flag = " # no json data"
		} else {
			jfiles, _ := ioutil.ReadDir("feeds/" + item.Name() + "/json")
			flag = fmt.Sprintf(" %d jsons", len(jfiles)-1)
		}
		fmt.Printf("%s: %s\n", item.Name(), flag)
	}

	fmt.Printf("\n")
}

func removeLegacy() {
	jf, _ := filepath.Glob(RunCfg.jpath + "posts_*")
	if len(jf) == 0 {
		return
	}
	fmt.Printf("Converting legacy prefixes: posts_ ")
	for _, f := range jf {
		nf := strings.Replace(f, "posts_", "", -1)
		if os.Rename(f, nf) != nil {
			outerror(1, "\nFATAL: Can't rename %s\n", f)
		}
	}
	fmt.Printf("image_ ")
	jim, _ := filepath.Glob(RunCfg.mediapath + "image_*")
	for _, f := range jim {
		nf := strings.Replace(f, "image_", "", -1)
		if os.Rename(f, nf) != nil {
			outerror(1, "\nFATAL: Can't rename %s\n", f)
		}
	}
	fmt.Printf("media_ \n")
	jim, _ = filepath.Glob(RunCfg.mediapath + "media_*")
	for _, f := range jim {
		nf := strings.Replace(f, "media_", "", -1)
		if os.Rename(f, nf) != nil {
			outerror(1, "\nFATAL: Can't rename %s\n", f)
		}
	}
}
