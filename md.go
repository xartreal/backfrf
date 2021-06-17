// md
package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"backfrf.v8md/frfmarkdown"
)

func loadmdtemplates(singlemode bool) {
	frfmarkdown.Templates = &frfmarkdown.THtmlTemplate{
		Comment: loadtfile("template/template_comment.md"),
		Item:    loadtfile("template/template_item.md"),
		File:    loadtfile("template/template_file.md"),
	}
	//set params
	frfmarkdown.Params = frfmarkdown.TParams{Feedpath: RunCfg.feedpath, Step: Config.step,
		Singlemode: singlemode, IndexPrefix: "index_", IndexPostfix: ".md",
		ImagePrefix: Config.mdmedia}
	//	jpath = RunCfg.feedpath + "json/posts_"
}

func loadTXL(name string) {
	fbin, err := ioutil.ReadFile(name)
	if err != nil {
		outerror(2, "FATAL: File %s not found or read error", name)
	}
	lines := strings.Split(string(fbin), "\n")
	for _, lval := range lines {
		if !strings.Contains(lval, "=") {
			continue
		}
		s := strings.Split(strings.TrimSpace(lval), "=")
		frfmarkdown.RXL[s[0]] = s[1]
	}
}

func checkTXL(keys []string) {
	for _, v := range keys {
		if _, ok := frfmarkdown.RXL[v]; !ok {
			outerror(1, "FATAL: Required TXL key '%s' not found\n", v)
		}
	}
}

func genPmarkdown(list []string, id string, isindex bool, title string, pen string, maxeof int) string {
	maxx := len(list) - 1
	//	outtext := "### " + title + " \n"
	outtext := ""
	prefix := ""
	for i := 0; i < maxx; i++ {
		if len(list[i]) < 2 {
			continue
		}
		//	fmt.Printf("l=%v\n", list[i])
		xtext, xtime := frfmarkdown.LoadJson(RunCfg.jpath+list[i]).ToMarkdown(list[i], "")
		//		if Config.allhtml == 1 {
		if Config.mddate == 1 {
			prefix = xtime + "-"
		}
		afname := "md/" + prefix + list[i] + ".md"
		ioutil.WriteFile(RunCfg.feedpath+afname, []byte(xtext), 0644)
		//		}
		//outtext += `[[` + prefix + list[i] + ".md]] \n"
		outtext += frfmarkdown.MkQLink(prefix+list[i]+".md") + " \n"
	}
	//	ptitle := RunCfg.feedname + " - " + title + " - backfrf"
	return frfmarkdown.MkMdPage(id, outtext, isindex, maxeof, RunCfg.feedname, title)
}

func genmd(offset int, maxeof int) {
	toffset := strconv.Itoa(offset)
	data, _ := ioutil.ReadFile(RunCfg.list + toffset)
	list := strings.Split(string(data), "\n")
	outname := "md/index_" + toffset + ".md"
	//genPmarkdown(list, toffset, true, "offset "+toffset, "", maxeof)
	outfiletext := genPmarkdown(list, toffset, true, "offset "+toffset, "", maxeof)
	ioutil.WriteFile(RunCfg.feedpath+outname, []byte(outfiletext), 0644)
	fmt.Printf("\roffset %d done", offset)
}
