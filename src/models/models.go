package models

type Message struct {
	Name string `json:"name,omitempty"` // only sent in initial message

	// group and room sent with message to designate targets
	Group string `json:"group,omitempty"` // usually for control data / screenshots OR audio
	Room  string `json:"room,omitempty"`  // usually for audio

	// Message data
	Recipient string `json:"recipient"`
	Kind      string `json:"kind,omitempty"`
	N         int    `json:"n"`
	Z         int    `json:"z"`
	Fast      bool   `json:"fast,omitempty"`
	Twitch    bool   `json:"twitch"`
	Img       string `json:"img,omitempty"`
	Audio     string `json:"audio,omitempty"`
}
