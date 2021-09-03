package model

type ForumResponse struct {
	ForumID   uint64 `json:"forum_id"`
	Name      string `json:"name"`
	Sort      uint   `json:"sort"`
	Parent    uint   `json:"parent"`
	Intro     string `json:"intro"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
