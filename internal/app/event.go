package app

import (
	"forum/internal/pkg/constants/event"
	"forum/internal/services"

	"github.com/gordon-zhiyong/beehive"
)

//RunBus bootstrap event listener
func EventSubecribe() {
	beehive.Subscribe(event.EventUpdateForum, services.ForumService.BuildInfoCache)
	beehive.Subscribe(event.EventNewForum,
		services.ForumService.BuildInfoCache,
		services.ForumService.BuildListCache)
	beehive.Subscribe(event.EventDeleteForum,
		services.ForumService.BuildListCache,
		services.ForumService.BuildInfoCache)
	beehive.Subscribe(event.EventRefreshForumCache, services.ForumService.BuildCache)
}
