package capsule

import "forum/pkg/event"

var Bus = event.NewBus(10, "./forum_bus.json")
