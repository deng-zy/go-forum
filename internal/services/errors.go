package services

import "errors"

var (
	ErrNameDuplicate    = errors.New("forum name duplicate")
	ErrForumIdDuplicate = errors.New("forum_id duplicate")
	ErrInternal         = errors.New("net worker error")
)
