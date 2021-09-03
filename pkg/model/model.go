package model

import (
	"gorm.io/gorm"
)

type Forum struct {
	gorm.Model
	ForumId uint64 `gorm:"column:forum_id;type:bigint(20) unsigned;NOT NULL" json:"forum_id"`       // 社区id
	Name    string `gorm:"column:name;type:varchar(32);NOT NULL" json:"name"`                       // 名字
	Intro   string `gorm:"column:intro;type:varchar(200);NOT NULL" json:"intro"`                    // 介绍
	Sort    uint   `gorm:"column:sort;type:smallint(5) unsigned;default:0;NOT NULL" json:"sort"`    // 排序
	Parent  uint   `gorm:"column:parent;type:bigint(20) unsigned;default:0;NOT NULL" json:"parent"` // 父板块
}
