package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"forum/internal/handlers/dashboard/request"
	"forum/internal/pkg/capsule"
	"forum/internal/pkg/config"
	"forum/internal/pkg/constants"
	"forum/internal/repositories"
	"forum/pkg/event"
	"forum/pkg/model"
	"forum/pkg/snowflake"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
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
var redisClient = capsule.RedisClient()

//ForumService forum service instance
var ForumService = newForumService()

var forumInfoCacheKey = "forum:info"
var forumListCacheKey = "forum:list"

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

	err := forumRepository.Create(model)
	if err != nil {
		return err
	}
	capsule.Bus.Publish(constants.EventNewForum, model.ForumId)
	return nil
}

func (f *forumService) Update(forumID uint64, form *request.Forum) error {
	info, err := forumRepository.Info(forumID)
	if err != nil {
		return err
	}

	exists, err := forumRepository.InfoWithName(form.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if exists.ForumId != 0 && exists.ForumId != forumID {
		return ErrNameDuplicate
	}

	info.Name = form.Name
	info.Intro = form.Intro
	info.Sort = form.Sort
	info.Parent = form.Parent

	err = forumRepository.Update(forumID, info)
	if err != nil {
		return err
	}

	capsule.Bus.Publish(constants.EventUpdateForum, forumID)

	return nil
}

func (f *forumService) Delete(forum uint64) error {
	err := forumRepository.Delete(forum)
	if err != nil {
		return err
	}

	capsule.Bus.Publish(constants.EventNewForum, forum)
	return nil
}

func (f *forumService) All(limit, offset int) []*model.Forum {
	return forumRepository.All(limit, offset)
}

func (f *forumService) Show(forumID uint64) (forum *model.Forum, err error) {
	forum, err = forumRepository.Info(forumID)
	return
}

func (f *forumService) BuildInfoCache(e *event.Event) error {
	forum := e.Data.(uint64)
	topic := e.Topic
	ctx := context.Background()

	if topic == constants.EventDeleteForum {
		_, err := redisClient.HDel(ctx, forumInfoCacheKey, strconv.FormatUint(forum, 10)).Result()
		return err
	}

	info, err := forumRepository.Info(forum)
	if err != nil {
		return err
	}

	cache, err := json.Marshal(info)
	_, err = redisClient.HSet(ctx, forumInfoCacheKey, forum, cache).Result()
	if err != nil {
		return err
	}

	return nil
}

func (f *forumService) BuildListCache(e *event.Event) error {
	fmt.Println("recv event", e.Topic, e.Data)
	forums := forumRepository.AllForumID()
	forumIDLi := make([]interface{}, len(forums))
	for i, forumID := range forums {
		forumIDLi[i] = forumID
	}
	fmt.Println("forumIDLi", forumIDLi)

	ctx := context.Background()
	_, err := redisClient.SAdd(ctx, forumListCacheKey, forumIDLi...).Result()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (f *forumService) BuildCache(e *event.Event) error {
	capsule.Bus.Publish(constants.EventNewForum, nil)
	forumIDLi := forumRepository.AllForumID()
	for _, forumID := range forumIDLi {
		capsule.Bus.Publish(constants.EventUpdateForum, forumID)
	}
	return nil
}
