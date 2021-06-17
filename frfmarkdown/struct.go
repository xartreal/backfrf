package frfmarkdown

type TPostJson struct {
	Users []struct {
		ID          string `json:"id"`
		Username    string `json:"username"`
		ScreenName  string `json:"screenName"`
		IsPrivate   string `json:"isPrivate"`
		IsProtected string `json:"isProtected"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
		Type        string `json:"type"`
	} `json:"users"`
	Subscriptions []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		User string `json:"user"`
	} `json:"subscriptions"`
	Subscribers []struct {
		ID             string   `json:"id"`
		Username       string   `json:"username"`
		ScreenName     string   `json:"screenName"`
		IsPrivate      string   `json:"isPrivate"`
		IsProtected    string   `json:"isProtected"`
		CreatedAt      string   `json:"createdAt"`
		UpdatedAt      string   `json:"updatedAt"`
		Type           string   `json:"type"`
		IsRestricted   string   `json:"isRestricted,omitempty"`
		Administrators []string `json:"administrators,omitempty"`
	} `json:"subscribers"`
	Posts struct {
		ID                     string   `json:"id"`
		Body                   string   `json:"body"`
		CommentsDisabled       string   `json:"commentsDisabled"`
		CreatedAt              string   `json:"createdAt"`
		UpdatedAt              string   `json:"updatedAt"`
		CommentLikes           string   `json:"commentLikes"`
		OwnCommentLikes        string   `json:"ownCommentLikes"`
		OmittedCommentLikes    string   `json:"omittedCommentLikes"`
		OmittedOwnCommentLikes string   `json:"omittedOwnCommentLikes"`
		CreatedBy              string   `json:"createdBy"`
		PostedTo               []string `json:"postedTo"`
		Comments               []string `json:"comments"`
		Attachments            []string `json:"attachments"`
		Likes                  []string `json:"likes"`
		OmittedComments        string   `json:"omittedComments"`
		OmittedLikes           string   `json:"omittedLikes"`
	} `json:"posts"`
	Comments []struct {
		ID         string `json:"id"`
		Body       string `json:"body"`
		CreatedAt  string `json:"createdAt"`
		UpdatedAt  string `json:"updatedAt"`
		HideType   string `json:"hideType"`
		Likes      string `json:"likes"`
		HasOwnLike bool   `json:"hasOwnLike"`
		SeqNumber  string `json:"seqNumber"`
		PostID     string `json:"postId"`
		CreatedBy  string `json:"createdBy"`
	} `json:"comments"`
	Attachments []struct {
		ID           string `json:"id"`
		FileName     string `json:"fileName"`
		FileSize     string `json:"fileSize"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnailUrl"`
		MediaType    string `json:"mediaType"`
		CreatedAt    string `json:"createdAt"`
		UpdatedAt    string `json:"updatedAt"`
		CreatedBy    string `json:"createdBy"`
	} `json:"attachments"`
}

type TXLines map[string]string

type GroupSType struct {
	Id          string
	IsPrivate   string
	IsProtected string
}

type TGList = map[string]GroupSType

type XPost struct {
	PostJson TPostJson
	usrindex TGList
	//	feedpath string
}

type THtmlTemplate struct {
	Comment string
	Item    string
	File    string
	Cal     string
}

var Templates *THtmlTemplate
