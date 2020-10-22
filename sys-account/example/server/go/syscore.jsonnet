local coreTpl = import "../../../../sys-core/service/go/template.syscore.libsonnet";
local loadVar = import "vendor/config/mixin.loadfn.libsonnet";

local cfg = {
    sysCoreConfig: {
       db: coreTpl.CoreDB {
           name: "gcn.db",
           encryptKey: loadVar(prefixName="SYS_CORE", env="DB_ENCRYPT_KEY").val,
           dbDir: "./bin-all/db",
       },
       cron: coreTpl.CoreCron {
           backupSchedule: "@daily",
       }
    }
};

std.manifestYamlDoc(cfg)