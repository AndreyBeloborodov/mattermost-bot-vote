package main

// Vote Структура для голосования
type Vote struct {
	ID        string         `json:"id"`
	Question  string         `json:"question"`
	Options   []string       `json:"options"`
	Votes     map[string]int `json:"votes"`
	CreatorID string         `json:"creator_id"`
	Closed    bool           `json:"closed"`
}
