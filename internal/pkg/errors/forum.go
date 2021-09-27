package errors

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	//ErrForumNameAlreadyExist forum name already exists
	ErrForumNameAlreadyExist = errors.New("forum name already exist")
	//ErrForumIDError forum ID already exists
	ErrForumIDError = errors.New("forum id already exist")
	//ErrForumNotExists forum not found
	ErrForumNotExists = gorm.ErrRecordNotFound
)
