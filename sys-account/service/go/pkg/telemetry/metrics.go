package telemetry

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/sirupsen/logrus"
)

const (
	METRICS_REGISTERED_USERS = "sys_account_registered_users"
	METRICS_VERIFIED_USERS   = "sys_account_verified_users"
	METRICS_JOINED_PROJECT   = "sys_account_joined_projects"
)

type SysAccountMetrics struct {
	RegisteredUserMetrics *metrics.Counter
	VerifiedUserMetrics   *metrics.Counter
	JoinedProjectMetrics  *metrics.Histogram
}

func NewSysAccountMetrics(logger *logrus.Entry) *SysAccountMetrics {
	logger.Infof("Registering Sys-Account Metrics")
	return &SysAccountMetrics{
		RegisteredUserMetrics: metrics.NewCounter(METRICS_REGISTERED_USERS),
		VerifiedUserMetrics:   metrics.NewCounter(METRICS_VERIFIED_USERS),
		JoinedProjectMetrics:  metrics.NewHistogram(METRICS_JOINED_PROJECT),
	}
}
