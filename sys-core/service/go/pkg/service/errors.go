package service

import (
	"fmt"
)

const (
	moduleName = "sys-core-db: "
)

type ErrorReason int

const (
	errRegisterModelEmpty = iota
)

type Error struct {
	Reason ErrorReason
	Err    error
}

func (err Error) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("%s (%v)", err.description(), err.Err)
	}
	return err.description()
}

func (err Error) description() string {
	return moduleName + err.desc()
}

func (err Error) desc() string {
	switch err.Reason {
	case errRegisterModelEmpty:
		return "registering model to db invalid, empty map"
	default:
		return "unknown error occurred"
	}
}
