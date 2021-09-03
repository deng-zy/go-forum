package request

type Forum struct {
	Name   string `form:"name" binding:"required,gt=0,lte=32"`
	Intro  string `form:"intro" binding:"required,gt=0,lte=200"`
	Sort   uint   `form:"sort" binding:"omitempty,required,numeric"`
	Parent uint   `form:"parent" binding:"omitempty,required,numeric"`
}
