// timeline
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/xartreal/frfpanehtml"
)

var newitems []string

var nlflag bool

type TTimeline struct {
	feedname   string
	offset     int
	textoffset string
	lasterr    error
	body       []byte
}

func inmedialist(fext string) bool {
	if RunCfg.extflag {
		return true
	}
	if len(fext) == 0 {
		return false
	}
	lext := strings.Replace(fext, ".", "", 1)
	return strings.Contains(Config.filter, lext)
}

func intWriteFile(url string, id, fext string) {
	ioutil.WriteFile(url, []byte("media file not loaded"), 0755)
	MyStat.newimages++
}

func intGetPost(id, newmark, changemarker string) {
	TimelineDB.MyCollection.Set([]byte(id), []byte(newmark))
	if Config.debugmode == 1 {
		fmt.Printf("\n" + changemarker + id) //out post id
	}
	getPost(id, false)
	newitems = append(newitems, id)
	nlflag = true
}

func getFile(id string, url string, media string, singlemode bool) {
	var mtype string
	fext := path.Ext(url)
	fname := id + fext
	//	mediapath := RunCfg.feedpath + "media/"
	if strings.EqualFold(media, "image") {
		mtype = "image_"
	} else {
		mtype = "media_"
		if (!singlemode && Config.loadmedia != 1) || (Config.loadmedia == 1 && !inmedialist(fext)) {
			intWriteFile(RunCfg.mediapath+mtype+fname, id, fext)
			return
		}
	}
	fname = mtype + fname
	fnpath := RunCfg.mediapath + fname
	if isexists(fnpath) { // if file exists
		return
	}
	if RunCfg.jsononly {
		intWriteFile(fnpath, id, fext)
		return
	}

	body := httpfile(url)
	//fmt.Println("response Body:", string(body))
	if !singlemode {
		ioutil.WriteFile(fnpath, body, 0644)
	} else {
		ioutil.WriteFile(RunCfg.feedpath+"/"+fname, body, 0644)
	}
	if Config.debugmode == 1 {
		fmt.Printf("\nattach: " + mtype + id + fext)
	}
	MyStat.newimages++
}

func getPost(id string, singlemode bool) {
	url := "https://freefeed.net/v2/posts/" + id + "?maxComments=all&maxLikes=all"
	body := httpget(url, true)
	//fmt.Println("response Body:", string(body))
	//json correction
	outtext := regJsReplace.ReplaceAll(body, []byte(`"$1":"$2"`))

	if !singlemode {
		ioutil.WriteFile(RunCfg.jpath+id, outtext, 0644)
	} else {
		ioutil.WriteFile(RunCfg.feedpath+"/"+id+".json", outtext, 0644)
	}
	// attach
	frfz := new(FrFfile)
	json.Unmarshal(outtext, frfz)
	for _, p := range frfz.Attachments {
		getFile(p.Id, p.Url, p.MediaType, singlemode)
	}
	if singlemode {
		fx := RunCfg.feedpath + "/" + id
		text := frfpanehtml.LoadJson(fx+".json").ToHtml(id, "")
		outfiletext := frfpanehtml.MkHtmlPage(id, text, false, 0, "", "")
		outfiletext = strings.Replace(outfiletext, `href="./kube.min.css"`, `href="../../template/kube.min.css"`, -1)
		ioutil.WriteFile(fx+".html", []byte(outfiletext), 0644)

	}
}

func mkTimeline(feedname string) *TTimeline {
	ntimeline := new(TTimeline)
	ntimeline = &TTimeline{feedname: feedname, offset: 0, textoffset: "0"}
	return ntimeline
}

func (timeline *TTimeline) getTimeline(offset int) *TTimeline {
	xoffs := strconv.Itoa(offset)
	url := "https://freefeed.net/v2/timelines/" + timeline.feedname + "?offset=" + xoffs
	if RunCfg.metafeed {
		url = "https://freefeed.net/v2/search?qs=" + RunCfg.metaurl + "&offset=" + xoffs
	}

	body := httpget(url, true)
	// json correction
	timeline.body = regJsReplace.ReplaceAll(body, []byte(`"$1":"$2"`))

	timeline.lasterr = ioutil.WriteFile(RunCfg.timeline+xoffs, timeline.body, 0644)
	timeline.offset = offset
	timeline.textoffset = xoffs
	return timeline
}

func (timeline *TTimeline) processTimeline() int {
	if timeline.body[0] == 0 {
		fmt.Printf(" -- timeout?\n\n")
		return 0
	}
	tlist := ""
	feedlen := 0
	frf := new(FrFjtml)
	if nlflag {
		fmt.Printf("\n")
	}
	fmt.Printf("\roffset: %d ", timeline.offset)
	nlflag = false
	json.Unmarshal(timeline.body, frf)
	//	for idx, p := range frf.Timelines.Posts {
	for idx, p := range frf.Posts {
		if Config.archive != 1 && len(frf.Posts[idx].FriendfeedUrl) != 0 { //skip archive post
			continue
		}
		tlist += p.Id + "\n"
		cx, _ := strconv.Atoi(frf.Posts[idx].OmittedComments)
		cmark := strconv.Itoa(len(frf.Posts[idx].Comments) + cx)
		newmark := frf.Posts[idx].UpdatedAt + cmark
		feedlen++
		if !isexists(RunCfg.jpath + p.Id) {
			intGetPost(p.Id, newmark, "")
			MyStat.newrecords++
		} else {
			oldmark, _ := TimelineDB.MyCollection.Get([]byte(p.Id))
			if !strings.EqualFold(string(oldmark), newmark) {
				intGetPost(p.Id, newmark, "* ")
				MyStat.changedrecords++
			}
		}
	}
	if len(tlist) > 0 {
		ioutil.WriteFile(RunCfg.list+timeline.textoffset, []byte(tlist), 0644)
	}
	MyStat.records += feedlen
	return feedlen
}
