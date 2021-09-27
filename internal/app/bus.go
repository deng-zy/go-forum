package app

import (
	"forum/internal/pkg/capsule"
	"forum/internal/pkg/constants"
	"forum/internal/services"
)

//RunBus bootstrap event listener
func RunBus() {
	subscribe()
	capsule.Bus.Bootstrap()
}

//subscribe event
func subscribe() {
	capsule.Bus.Subscribe(constants.EventUpdateForum, services.ForumService.BuildInfoCache)
	capsule.Bus.Subscribe(constants.EventNewForum, services.ForumService.BuildInfoCache,
		services.ForumService.BuildListCache)
	capsule.Bus.Subscribe(constants.EventDeleteForum, services.ForumService.BuildListCache,
		services.ForumService.BuildInfoCache)
	capsule.Bus.Subscribe(constants.EventRefreshForumCache, services.ForumService.BuildCache)
}
