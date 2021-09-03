package repositories

import (
	"fmt"
	"forum/internal/pkg/capsule"
	"forum/pkg/model"

	"gorm.io/gorm"
)

type forumRepository struct {
	db *gorm.DB
}

func newForumRepository() *forumRepository {
	return &forumRepository{
		db: capsule.DBConn(),
	}
}

var ForumRepository = newForumRepository()

func (f *forumRepository) Create(form *model.Forum) (err error) {
	err = f.db.Create(form).Error
	return
}

func (f *forumRepository) Exists(column string, value interface{}) bool {
	var count int64 = 0
	f.db.Model(&model.Forum{}).Where(fmt.Sprintf("%s=?", column), value).Count(&count)
	return count > 0
}

func (f *forumRepository) All(limit, offset int) []*model.Forum {
	forums := []*model.Forum{}
	f.db.Model(&forums).Limit(limit).Offset(offset).Find(&forums)
	return forums
}

func (f *forumRepository) Delete(forum uint64) error {
	err := f.db.Where("forum_id=?", forum).Delete(&model.Forum{}).Error
	return err
}
