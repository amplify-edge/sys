package telemetry

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (

	METRICS_REGISTERED_USERS = "sys_account_registered_users"
	METRICS_VERIFIED_USERS   = "sys_account_verified_users"
	METRICS_JOINED_PROJECT   = "sys_account_joined_projects"
	JoinProjectLabel         = `%s{org_id="%s", project_id="%s"}`
)

type SysAccountMetrics struct {
	RegisteredUserMetrics    prom.Counter
	VerifiedUserMetrics      prom.Counter
	UserJoinedProjectMetrics *prom.CounterVec
}

func NewSysAccountMetrics(logger *logrus.Entry) *SysAccountMetrics {
	logger.Infof("Registering Sys-Account Metrics")
	return &SysAccountMetrics{
		RegisteredUserMetrics: prom.NewCounter(prom.CounterOpts{
			Name: metricsRegisteredUsers,
			Help: "sys-account metrics for registered users (a counter)",
		}),
		VerifiedUserMetrics: prom.NewCounter(prom.CounterOpts{
			Name: metricsVerifiedUsers,
			Help: "sys-account metrics for verified users ( a counter)",
		}),
		UserJoinedProjectMetrics: prom.NewCounterVec(prom.CounterOpts{
			Name: metricsJoinedProject,
			Help: "sys-account metrics for whenever user joined project (a categorized counter)",
		}, []string{"project_id", "org_id"}),
	}
}

func (s *SysAccountMetrics) RegisterMetrics() {
	prom.MustRegister(s.UserJoinedProjectMetrics, s.RegisteredUserMetrics, s.VerifiedUserMetrics)
}
