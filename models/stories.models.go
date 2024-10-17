package models

type StoriesCreateInput struct {
	Theme    string `json:"theme"`
	Language string `json:"language"`
}

type StoriesCreateTitle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type StoriesCreateTitleFormat struct {
	Titles []StoriesCreateTitle `json:"titles"`
}

type StoriesCreateFirstPartInput struct {
	StoriesCreateInput
	Title       string `json:"title"`
	Description string `json:"description"`
}

type StoriesCreateParagraphContinueInput struct {
	StoriesCreateFirstPartInput
	Paragraph string `json:"paragraph"`
	Choice    string `json:"choice"`
}

type StoriesCreateParagraph struct {
	Paragraph string   `json:"paragraph"`
	Choices   []string `json:"choices"`
}
