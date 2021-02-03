local dbcfg = import "vendor/github.com/getcouragenow/sys-share/sys-core/service/config/mixins/mixin.db.libsonnet";
local tokencfg = import "vendor/github.com/getcouragenow/sys-share/sys-core/service/config/mixins/mixin.jwt.libsonnet";
{
    local cfg = self,
    CoreDB:: dbcfg.DB {
         name: "sysaccount.db",
         encryptKey: "testEncryptKey",
         dbDir: "./db/sys",
         deletePrevious: true,
    },
     CoreCron:: dbcfg.Cron {
         backupSchedule: "@daily",
         rotateSchedule: "@every 24h",
         backupDir: cfg.CoreDB.dbDir + "/sysaccount-backup"
    },
    CoreMail:: {
      senderName: "gutterbacon",
      senderMail: "gutterbacon@example.com",
      productName: "SOME_PRODUCT",
      logoUrl: "https://via.placeholder.com/500x500?text=YOUR+LOGO+HERE",
      copyright: "SOME_COPYRIGHT_MSG",
      troubleContact: "SOME_TROUBLESHOOT_CONTACT_HERE",
      sendgrid: {
        apiKey: "SENDGRID_API_KEY_HERE",
      },
      smtp: {
        host: "smtp.gmail.com",
        port: 465,
        email: "winwisely268@gmail.com",
        password: "blablabla",
      }
    },
    FileDB:: dbcfg.DB {
        name: "sysfiles.db",
        encryptKey: "testEncryptKey",
        dbDir: "./db/sys",
        deletePrevious: true,
    },
    FileCron:: {
        backupSchedule: "@daily",
        rotateSchedule: "@every 24h",
        backupDir: cfg.FileDB.dbDir + "/sysfile-backup"
    },
    UnauthenticatedRoutes:: [
    	"/v2.sys_account.services.AuthService/Login",
    	"/v2.sys_account.services.AuthService/Register",
    	"/v2.sys_account.services.AuthService/ResetPassword",
    	"/v2.sys_account.services.AuthService/ForgotPassword",
    	"/v2.sys_account.services.AuthService/RefreshAccessToken",
    	"/v2.sys_account.services.AuthService/VerifyAccount",
    	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
    ],
    AccessToken:: tokencfg.Token {
        secret: "some_jwt_access_secret",
        expiry: 3600,
    },
    RefreshToken:: tokencfg.Token {
        secret: "some_jwt_refresh_secret",
        expiry: cfg.AccessToken.expiry * 100,
    },
    sysAccountConfig: {
        superUserFilePath: "./config/supers.yml",
        unauthenticatedRoutes: cfg.UnauthenticatedRoutes,
        sysCoreConfig: {
            db: cfg.CoreDB,
            cron: cfg.CoreCron,
        },
        jwt: {
            access: cfg.AccessToken,
            refresh: cfg.RefreshToken,
        },
        sysFileConfig: {
            db: cfg.FileDB,
            cron: cfg.FileCron,
        },
        mailConfig: cfg.CoreMail,
    }
}