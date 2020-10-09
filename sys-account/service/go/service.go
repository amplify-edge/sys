package service

import (
	"context"
	"github.com/sirupsen/logrus"
)

type SysAccountService struct {
	logger *logrus.Entry
	authInterceptorFunc func(context.Context) (context.Context, error)
	port int
	ProxyService *pkg
}

