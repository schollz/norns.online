package models

type Message struct {
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"`
	Recipient string `json:"recipient,omitempty"`

	Img    string `json:"img,omitempty"`
	Kind   string `json:"kind,omitempty"`
	N      int    `json:"n"`
	Z      int    `json:"z"`
	Fast   bool   `json:"fast,omitempty"`
	Twitch bool   `json:"twitch"`
	MP3    string `json:"mp3,omitempty"`
}
