{
    local cfg = self,
    CoreDb:: {
      name: "getcouragenow.db",
      encryptKey: "testkey!@",
      rotationDuration: 1,
      dbDir: "./db"
    },
    CoreCron:: {
      backupSchedule: "@daily",
      rotateSchedule: "@every 24h",
      backupDir: cfg.CoreDb.dbDir + "/backup"
    },
    sysCoreConfig: {
        db: self.CoreDb,
        cron: self.CoreCron,
    }
}

