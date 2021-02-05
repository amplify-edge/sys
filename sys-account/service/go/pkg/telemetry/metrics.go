package telemetry

import (
	"github.com/VictoriaMetrics/metrics"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging"
)

const (
	METRICS_REGISTERED_USERS = "sys_account_registered_users"
	METRICS_VERIFIED_USERS   = "sys_account_verified_users"
	METRICS_JOINED_PROJECT   = "sys_account_joined_projects"
	JoinProjectLabel         = `%s{org_id="%s", project_id="%s"}`
)

type SysAccountMetrics struct {
	RegisteredUserMetrics *metrics.Counter
	VerifiedUserMetrics   *metrics.Counter
}

func NewSysAccountMetrics(logger logging.Logger) *SysAccountMetrics {
	logger.Info("Registering Sys-Account Metrics")
	return &SysAccountMetrics{
		RegisteredUserMetrics: metrics.NewCounter(METRICS_REGISTERED_USERS),
		VerifiedUserMetrics:   metrics.NewCounter(METRICS_VERIFIED_USERS),
	}
}
