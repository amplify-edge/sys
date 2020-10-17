local accMixin = import "../../sys-account/service/go/template.sysaccount.libsonnet";
local loadVar = import "../../sys-core/service/go/mixin.loadfn.libsonnet";

local cfg = {
    sysAccountConfig: accMixin.sysAccountConfig {
        unauthenticatedRoutes: accMixin.UnauthenticatedRoutes + [
            "/v2.services.AccountService/ListAccounts"
        ],
    }
};

std.manifestYamlDoc(cfg)