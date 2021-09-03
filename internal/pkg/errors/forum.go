package errors

import "github.com/pkg/errors"

var (
	ErrForumNameAlreadyExist = errors.New("forum name already exist")
	ErrForumIDError          = errors.New("forum id already exist")
)
