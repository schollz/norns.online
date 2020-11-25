package models

const AUDIO_PACKET_SECONDS = 1

type Message struct {
	// server should be initialized with these
	Name  string `json:"name,omitempty"`  // only sent in initial message
	Group string `json:"group,omitempty"` // usually for control data / screenshots OR audio
	Room  string `json:"room,omitempty"`  // usually for audio

	// Message data
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Kind      string `json:"kind,omitempty"`
	N         int    `json:"n"`
	Z         int    `json:"z"`
	Fast      bool   `json:"fast,omitempty"`
	Twitch    bool   `json:"twitch"`
	Img       string `json:"img,omitempty"`
	Audio     string `json:"audio,omitempty"`
}
