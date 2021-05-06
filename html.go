// html
package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/xartreal/frfpanehtml"
)

var jpath string //json path

func loadtemplates(singlemode bool) {
	frfpanehtml.Templates = &frfpanehtml.THtmlTemplate{
		Comment: loadtfile("template/template_comment.html"),
		Item:    loadtfile("template/template_item.html"),
		File:    loadtfile("template/template_file.html"),
	}
	//set params
	frfpanehtml.Params = frfpanehtml.TParams{Feedpath: RunCfg.feedpath, Step: Config.step,
		Singlemode: singlemode, IndexPrefix: "index_", IndexPostfix: ".html"}
	jpath = RunCfg.feedpath + "json/posts_"
}

func genPhtml(list []string, id string, isindex bool, title string, pen string, maxeof int) string {
	maxx := len(list) - 1
	outtext := "<h2>" + title + "</h2>"
	for i := 0; i < maxx; i++ {
		if len(list[i]) < 2 {
			continue
		}
		//	fmt.Printf("l=%v\n", list[i])
		xtext := frfpanehtml.LoadJson(jpath+list[i]).ToHtml(list[i], "")
		if Config.allhtml == 1 {
			afname := "html/" + list[i] + ".html"
			ioutil.WriteFile(RunCfg.feedpath+afname, []byte(xtext), 0644)
		}
		outtext += xtext + "<hr>"
	}
	ptitle := RunCfg.feedname + " - " + title + " - backfrf"
	return frfpanehtml.MkHtmlPage(id, outtext, isindex, maxeof, RunCfg.feedname, ptitle)
}

func genhtml(offset int, maxeof int) {
	toffset := strconv.Itoa(offset)
	data, _ := ioutil.ReadFile(RunCfg.feedpath + "index/list_" + toffset)
	list := strings.Split(string(data), "\n")
	outname := "html/index_" + toffset + ".html"
	outfiletext := genPhtml(list, toffset, true, "offset "+toffset, "", maxeof)
	ioutil.WriteFile(RunCfg.feedpath+outname, []byte(outfiletext), 0644)
	fmt.Printf("\roffset %d done", offset)
}
