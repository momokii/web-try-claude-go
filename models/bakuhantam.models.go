package models

type BakuHantamTopicList struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Link     string `json:"link"`
}

type BHTopicDetail struct {
	TweetOwner     string `json:"tweet_owner"`
	TweetOwnerData string `json:"tweet_owner_data"`
	TweetOwnerTime string `json:"tweet_owner_time"`
	QuotedUser     string `json:"quoted_user"`
	QuotedData     string `json:"quoted_data"`
}
