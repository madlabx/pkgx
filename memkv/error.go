package memkv

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrExpired           = errors.New("expired record")
	ErrInvalidRecordType = errors.New("invalid record type")
	ErrNotFound          = gorm.ErrRecordNotFound
)
