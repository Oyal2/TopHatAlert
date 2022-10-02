package TopHat

import (
	"net/http"
	"time"
)

type SettingsInfo struct {
	SessionID string `json:"sessionID"`
}

type TopHatInfo struct {
	Webhook string
	CourseID int
	PublicCode string
	Omit 	map[interface{}]struct{}
	Client *http.Client
	Now 	time.Time
}


type Thumbnail struct {
	URL string `json:"url"`
}
type Fields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}
type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}
type Author struct {
	Name    string `json:"name"`
	IconUrl string `json:"icon_url"`
}
type Embeds struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Color     int       `json:"color"`
	Timestamp string    `json:"timestamp"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Image	  Image 	`json:"image"`
	Fields    []Fields  `json:"fields"`
	Author    Author    `json:"author"`
	Footer    Footer    `json:"footer"`
}
type Image struct {
	URL string `json:"url"`
}
type Webhook struct {
	Username  string   `json:"username"`
	AvatarURL string   `json:"avatar_url"`
	Embeds    []Embeds `json:"embeds"`
}