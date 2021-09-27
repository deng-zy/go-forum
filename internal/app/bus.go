package app

import (
	"forum/internal/pkg/capsule"
	"forum/internal/pkg/constants/event"
	"forum/internal/services"
)

//RunBus bootstrap event listener
func RunBus() {
	subscribe()
	capsule.Bus.Bootstrap()
}

//subscribe event
func subscribe() {
	capsule.Bus.Subscribe(event.EventUpdateForum, services.ForumService.BuildInfoCache)
	capsule.Bus.Subscribe(event.EventNewForum,
		services.ForumService.BuildInfoCache,
		services.ForumService.BuildListCache)
	capsule.Bus.Subscribe(event.EventDeleteForum,
		services.ForumService.BuildListCache,
		services.ForumService.BuildInfoCache)
	capsule.Bus.Subscribe(event.EventRefreshForumCache, services.ForumService.BuildCache)
}
