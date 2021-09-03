package services

import (
	"forum/internal/handlers/dashboard/request"
	"forum/internal/pkg/config"
	"forum/internal/repositories"
	"forum/pkg/model"
	"forum/pkg/snowflake"
	"time"

	"github.com/spf13/viper"
)

type forumService struct {
}

var snowFlake *snowflake.Node
var dataCenter int64
var node int64
var epoch time.Time

func init() {
	config.Load()
	node = viper.GetInt64("id.node")
	epoch = viper.GetTime("id.epoch")
	dataCenter = viper.GetInt64("id.dataCenter")

	var err error
	snowFlake, err = snowflake.NewNode(node, dataCenter, epoch)
	if err != nil {
		panic(err)
	}
}

var forumRepository = repositories.ForumRepository
var ForumService = newForumService()

func newForumService() *forumService {
	return &forumService{}
}

func (f *forumService) Create(form *request.Forum) error {
	if forumRepository.Exists("name", form.Name) {
		return ErrNameDuplicate
	}

	forumID := snowFlake.Generate()
	if forumRepository.Exists("forum_id", forumID) {
		return ErrForumIdDuplicate
	}

	model := &model.Forum{
		Name:    form.Name,
		ForumId: uint64(forumID),
		Intro:   form.Intro,
		Sort:    form.Sort,
		Parent:  form.Parent,
	}

	return forumRepository.Create(model)
}

func (f *forumService) All(limit, offset int) []*model.Forum {
	return forumRepository.All(limit, offset)
}

func (f *forumService) Delete(forum uint64) error {
	return forumRepository.Delete(forum)
}
