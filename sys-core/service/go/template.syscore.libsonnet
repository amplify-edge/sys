local dbcfg = import "vendor/github.com/getcouragenow/sys-share/sys-core/service/config/mixins/mixin.db.libsonnet";

{
    local cfg = self,
    CoreDB:: dbcfg.DB {
      name: "core.db",
    },
    CoreCron:: dbcfg.Cron,
    sysCoreConfig: {
        db: self.CoreDB,
        cron: self.CoreCron,
    }
}