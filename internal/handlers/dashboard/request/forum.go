package request

type Forum struct {
	Name   string `form:"name" json:"name" binding:"required,gt=0,lte=32"`
	Intro  string `form:"intro" json:"intro" binding:"required,gt=0,lte=200"`
	Sort   uint   `form:"sort" json:"sort" binding:"omitempty,required,numeric"`
	Parent uint   `form:"parent" json:"parent" binding:"omitempty,required,numeric"`
}
