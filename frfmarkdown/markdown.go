package frfmarkdown

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TParams struct {
	Feedpath     string
	Step         int
	Singlemode   bool
	IndexPrefix  string
	IndexPostfix string
	ImagePrefix  string
}

var Params TParams

var RXL = map[string]string{} //TXL format keys

var tagReplacer *strings.Replacer

var (
	RegJsReplace *regexp.Regexp
	RegToken     *regexp.Regexp
	RegUser      *regexp.Regexp
	RegHashtag   *regexp.Regexp
	RegUrl       *regexp.Regexp
)

func initRegs() {
	RegJsReplace = regexp.MustCompile(`"([a-zA-Z]{2,})":(\d{1,})`)
	RegUser = regexp.MustCompile(`(?s)([^a-zA-Z0-9\/]|^)@([a-zA-Z0-9\-]+)`)                                   //@username
	RegHashtag = regexp.MustCompile(`(?s)([^a-zA-Z0-9\/?'"]|^)#([^\s\x20-\x2F\x3A-\x3F\x5B-\x5E\x7B-\xBF]+)`) //hashtag
	RegUrl = regexp.MustCompile(`(?s)(https?:[^\s^\*(),\[{\]};'"><]+)`)                                       //url
}

func striptags(text string) string {
	/*	outtext := tagReplacer.Replace(text)
		return strings.Replace(outtext, "\n", "<br>", -1) */
	return strings.Replace(text, "\n", "<br>", -1)
}

func GenMdTplPage(template string, tstrings TXLines) string {
	tout := template
	for key, val := range tstrings {
		tout = strings.Replace(tout, "$"+key, val, -1)
	}
	return tout
}

func makeixlink(link string) string {
	return Params.IndexPrefix + link + Params.IndexPostfix
}

func makemdlink(key, name, item string) string {
	out := RXL[key]
	out = strings.Replace(out, "$name", name, -1)
	out = strings.Replace(out, "$item", item, -1)
	return out
}

func highlighter(text, pen string) string {
	out := RegUser.ReplaceAll([]byte(text), []byte(RXL["user"]))
	return string(out)
}

func MkQLink(link string) string {
	return makemdlink("qlink", "", link)
}

func (post *XPost) getgroups() TGList {
	var groups = TGList{}
	var tmpgroups = TGList{}
	for _, p := range post.PostJson.Subscribers {
		if strings.EqualFold(p.Type, "group") {
			tmpgroups[p.ID] = GroupSType{p.Username, p.IsPrivate, p.IsProtected}
		}
	}
	for _, p := range post.PostJson.Subscriptions {
		if tmpgroups[p.User].Id != "" {
			groups[p.ID] = GroupSType{tmpgroups[p.User].Id, tmpgroups[p.User].IsPrivate, tmpgroups[p.User].IsProtected}
		}
	}
	return groups
}

func (post *XPost) genCommentsMd(pen string) string {
	cText := ""
	for _, p := range post.PostJson.Comments { //frfcmts.Comments
		clikes := ""
		if p.Likes != "0" {
			clikes = p.Likes
		}
		cc := GenMdTplPage(Templates.Comment,
			TXLines{"comment": highlighter(striptags(p.Body), ""), "author": post.usrindex[p.CreatedBy].Id,
				"clikes": clikes})
		cText += cc
	}
	return cText
}

func (post *XPost) genLikesMd() string {
	if len(post.PostJson.Posts.Likes) == 0 {
		return ""
	}
	likesText := "Likes: "
	for _, p := range post.PostJson.Posts.Likes {
		likesText += post.usrindex[p].Id + ", "
	}
	likesText = strings.TrimSuffix(likesText, ", ") + "\n"
	return likesText
}

func (post *XPost) genAttachText() string {
	if len(post.PostJson.Attachments) == 0 {
		return ""
	}
	attachText := ""
	prefix := Params.ImagePrefix
	for _, p := range post.PostJson.Attachments {
		attachFile := ""
		file := p.ID + path.Ext(p.URL)
		if p.MediaType != "image" {
			attachFile += prefix + "media_" + file
		} else {
			attachFile += prefix + "image_" + file
		}
		attachText += makemdlink("attach", "", attachFile) + " "
	}
	return attachText
}

func (post *XPost) genGroupText(groups TGList) string {
	ghtml := ""
	gcnt := 0
	pscnt := len(post.PostJson.Posts.PostedTo)
	for _, p := range post.PostJson.Posts.PostedTo {
		if groups[p].Id != "" {
			ghtml += groups[p].Id + ":"
			gcnt++
		}
	}
	if (pscnt > 1) && (pscnt > gcnt) {
		ghtml = "+" + ghtml
	}
	return ghtml
}

func (post *XPost) getBLineItems() (string, string) {
	createdby := post.PostJson.Posts.CreatedBy
	uuname := post.usrindex[createdby].Id
	utime, _ := strconv.ParseInt(post.PostJson.Posts.CreatedAt, 10, 64)
	xtime := time.Unix(utime/1000, 0).Format("2006-01-02")
	return uuname, xtime
}

func (post *XPost) ToMarkdown(id, pen string) (string, string) {
	//users
	post.usrindex = TGList{}
	for _, p := range post.PostJson.Users { //frfusr.Users
		post.usrindex[p.ID] = GroupSType{p.Username, p.IsPrivate, p.IsProtected}
	}
	// groups
	ghtml := post.genGroupText(post.getgroups())
	//b-line
	uuname, xtime := post.getBLineItems() //auname := ghtml + uuname
	//likes
	likesText := post.genLikesMd()
	//attach
	attachText := post.genAttachText()
	//comments
	commText := post.genCommentsMd(pen)
	//assembly
	tmap := TXLines{"text": highlighter(striptags(post.PostJson.Posts.Body), ""), //post
		"auname": ghtml + uuname, "aname": uuname, "id": id, "time_html": xtime, //b-line
		"attach_html": attachText, "likes_html": likesText, "comm_html": commText, //attachs,likes,comments
	}
	//	fmt.Printf("t=%v\nmap=%v\n", Templates.Item, tmap)
	return GenMdTplPage(Templates.Item, tmap), xtime
}

func MkMdPage(id string, mdText string, isIndex bool, maxeof int, feedname string, title string) string {
	nav := ""
	if isIndex {
		ids, _ := strconv.Atoi(id)
		if ids != 0 {
			nav += makemdlink("link", "Previous", makeixlink(strconv.Itoa(ids-Params.Step))) + " "
		}
		if ids < maxeof {
			nav += makemdlink("link", "Next", makeixlink(strconv.Itoa(ids+Params.Step))) + " "
		}
	}
	title = makemdlink("title", "", title)
	outfiletext := GenMdTplPage(Templates.File, TXLines{"title": title, "html_text": mdText, "pager": nav, "feedname": feedname})
	return outfiletext
}

func LoadJson(filename string) *XPost {
	npost := new(XPost)
	fbin, _ := ioutil.ReadFile(filename)
	json.Unmarshal(fbin, &npost.PostJson)
	return npost
}

func init() {
	//	tagReplacer = strings.NewReplacer("<", "&lt;", ">", "&gt;")
	initRegs()
}
