// structs
package main

type TMyStat struct {
	records        int
	newrecords     int
	changedrecords int
	newimages      int
}

var MyStat TMyStat

// timeline

type FrFfile struct {
	Attachments []struct {
		Id           string `json:"id"`
		Url          string `json:"url"`
		ThumbnailUrl string `json:"thumbnailUrl"`
		MediaType    string `json:"mediaType"`
	} `json:"attachments"`
}

type FrFjtml struct {
	Timelines struct {
		Posts []string `json:"posts"`
	} `json:"timelines"`
	Posts []struct {
		Id              string   `json:"id"`
		Body            string   `json:"body"`
		CreatedAt       string   `json:"createdAt"`
		UpdatedAt       string   `json:"updatedAt"`
		FriendfeedUrl   string   `json:"friendfeedUrl"`
		CreatedBy       string   `json:"createdBy"`
		Attachments     []string `json:"attachments"`
		Likes           []string `json:"likes"`
		Comments        []string `json:"comments"`
		OmittedComments string   `json:"omittedComments"`
	} `json:"posts"`
}

// posts

type FrFJSON struct {
	Posts struct {
		Id          string   `json:"id"`
		Body        string   `json:"body"`
		PostedTo    []string `json:"postedTo"`
		CreatedAt   string   `json:"createdAt"`
		UpdatedAt   string   `json:"updatedAt"`
		CreatedBy   string   `json:"createdBy"`
		Attachments []string `json:"attachments"`
		Likes       []string `json:"likes"`
		Comments    []string `json:"comments"`
	} `json:"posts"`
}

