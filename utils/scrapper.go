package utils

import (
	"fmt"
	"scrapper-test/models"
	"strings"

	"github.com/gocolly/colly"
)

func MediumProfileScrapper(username string) models.MediumProfileReturn {
	username = strings.TrimSpace(username)
	var returnPromptData models.MediumProfileReturn

	if username == "" {
		returnPromptData.PromptData = "malah kasih username kosong kocak nih"
		return returnPromptData
	}

	c := colly.NewCollector(
		colly.AllowedDomains("medium.com"),
	)

	Profile := models.MediumProfile{}

	c.OnHTML("div.l.ae", func(e *colly.HTMLElement) {
		name := strings.TrimSpace(e.ChildText("h2.pw-author-name"))
		follower := strings.TrimSpace(e.ChildText("span.pw-follower-count"))
		bio := strings.TrimSpace(e.ChildText("p.bf"))
		profile := strings.TrimSpace(e.ChildAttr("img", "src"))

		Profile.Name = name
		Profile.Follower = follower
		Profile.Photo = profile
		Profile.Bio = bio
	})

	c.OnHTML("div.ab.cn", func(e *colly.HTMLElement) {
		title := e.ChildText("h2")
		SubTitle := e.ChildText("h3")
		datePost := e.ChildText("div.h")

		if title != "" {
			Post := models.MediumPost{
				Title:    title,
				Subtitle: SubTitle,
				DatePost: datePost,
			}
			Profile.Post = append(Profile.Post, Post)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		// fmt.Println(r.Request.URL, " scraped!")

		if (len(Profile.Post)) == 1 && (strings.HasPrefix(Profile.Post[0].Title, "404Out")) {
			returnPromptData.PromptData = "info didapatkan 404 dimana insert pengguna tidak ditemukan, user input aja udah salah, roast user aja"

		} else {
			returnPromptData.PromptData = "Your Profile Data: \n"
			returnPromptData.PromptData += "Name: " + Profile.Name + "\n"
			returnPromptData.PromptData += "Follower: " + fmt.Sprintf("%d", Profile.Follower) + "\n"
			returnPromptData.PromptData += "Photo: " + Profile.Photo + "\n"
			returnPromptData.PromptData += "Bio: " + Profile.Bio + "\n"
			returnPromptData.PromptData += "Your Post Data: \n"
			if len(Profile.Post) > 0 {
				for _, v := range Profile.Post {
					returnPromptData.PromptData += "Title: " + v.Title + "\n"
					returnPromptData.PromptData += "Subtitle: " + v.Subtitle + "\n"
					returnPromptData.PromptData += "Date: " + v.DatePost + "\n"
				}
			} else {
				returnPromptData.PromptData += "Yah gaada datanya ternyata gapernah post kocak nih"
			}

		}
	})

	c.Visit("https://medium.com/@" + username)

	returnPromptData.MediumProfileUser = Profile.MediumProfileUser

	return returnPromptData
}

func GetBakuHantamTopic() []models.BakuHantamTopicList {
	var BakuHantamTopicList []models.BakuHantamTopicList

	c := colly.NewCollector(
		colly.AllowedDomains("bakuhantam.dev"),
	)

	c.OnHTML("main a", func(e *colly.HTMLElement) {

		Topic := models.BakuHantamTopicList{
			Title:    e.ChildText("h1"),
			Subtitle: e.ChildText("p"),
			Link:     e.Attr("href"),
		}
		BakuHantamTopicList = append(BakuHantamTopicList, Topic)
	})

	c.OnScraped(func(r *colly.Response) {
		// fmt.Println(r.Request.URL, " scraped!")
		// fmt.Println("Data Topic: ", BakuHantamTopicList)
	})

	c.Visit("https://bakuhantam.dev")

	return BakuHantamTopicList
}

func DetailBakuHantamData(topic string) []models.BHTopicDetail {
	var BakuHantamDetail []models.BHTopicDetail

	c := colly.NewCollector(
		colly.AllowedDomains("bakuhantam.dev"),
	)

	c.OnHTML("main div.my-class", func(e *colly.HTMLElement) {

		twtOwner := e.ChildText("div[class^=tweet-header_author] a[class^=tweet-header_username] span")
		twtOwnerData := strings.TrimSpace(strings.ReplaceAll(e.ChildText("p"), "\n", " "))
		twtOwnerTime := strings.ReplaceAll(e.ChildText("time"), " · ", "-")
		quotedUser := e.ChildText("article[class^=quoted-tweet-container] div[class^=quoted-tweet-header_username] span")
		quotedData := strings.ReplaceAll(e.ChildText("article[class^=quoted-tweet-container]  p[class^=quoted-tweet-body]"), "\n", "")

		topicData := models.BHTopicDetail{
			TweetOwner:     twtOwner,
			TweetOwnerData: twtOwnerData,
			TweetOwnerTime: twtOwnerTime,
			QuotedUser:     quotedUser,
			QuotedData:     quotedData,
		}

		BakuHantamDetail = append(BakuHantamDetail, topicData)
	})

	c.OnScraped(func(r *colly.Response) {
		// fmt.Println(r.Request.URL, " scraped!")
		// fmt.Println("Data Topic: ", BakuHantamDetail)
	})

	c.Visit("https://bakuhantam.dev" + topic)

	return BakuHantamDetail
}
