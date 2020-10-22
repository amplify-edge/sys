local dbcfg = import "vendor/config/mixin.db.libsonnet";

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