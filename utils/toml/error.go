package toml

import (
	"errors"
)

var (
	// ErrorParamNotExists error for not exists params when read
	ErrorParamNotExists = errors.New("param not exists")
	// ErrorParamAlreadyExists error for already exists params when add
	ErrorParamAlreadyExists = errors.New("param already exists")
	// ErrorSectionNotExists error for not exists section when search
	ErrorSectionNotExists = errors.New("section not exists")
	// ErrorSectionAlreadyExists error for already exists section when add
	ErrorSectionAlreadyExists = errors.New("section already exists")
)