local accMixin = import "../../sys-account/service/go/template.sysaccount.libsonnet";

local cfg = {
    sysAccountConfig: accMixin.sysAccountConfig {
        unauthenticatedRoutes: accMixin.UnauthenticatedRoutes + [
            "/v2.sys_core.services.DbAdminService/Backup",
            "/v2.sys_core.services.DbAdminService/ListBackup",
            "/v2.sys_core.services.DbAdminService/Restore",
        ],
        initialSuperUsers: [
            {
                email: "superadmin@getcouragenow.org",
                password: "superadmin",
            }
        ],
    }
};

std.manifestYamlDoc(cfg)