package scache

import (
	"errors"
)

var (
	ErrKeyNotFound = errors.New("key not found in cache")
	ErrKeyIsExist  = errors.New("key has exist in cache")
)
