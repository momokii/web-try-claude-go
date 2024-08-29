package models

type MediumProfileUser struct {
	Name     string `json:"name"`
	Follower string `json:"follower"`
	Photo    string `json:"photo"`
	Bio      string `json:"bio"`
}

type MediumPost struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	DatePost string `json:"date_post"`
}

type MediumProfile struct {
	MediumProfileUser
	Post []MediumPost `json:"post"`
}

type MediumProfileReturn struct {
	MediumProfileUser
	PromptData string `json:"prompt_data"`
}
