package models

type ImageAnalysisRes struct {
	HaveEmotion      bool     `json:"have_emotion"`
	ImageDescription string   `json:"image_description"`
	EmotionDetection string   `json:"emotion_detection"`
	ObjectDetection  []string `json:"object_detection"`
	VisualElement    []string `json:"visual_element"`
}

type CreativeContentData struct {
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

type CreativeContentRecommendationRes struct {
	CreativeContent []CreativeContentData `json:"creative_content"`
}

type CreateImageGenerator struct {
	Prompt string `json:"prompt"`
}

type CreateTTS struct {
	Prompt   string `json:"prompt"`
	Language string `json:"language"`
}
