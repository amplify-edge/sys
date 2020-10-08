import 'package:sys_account/core/shared_repositories/utilities.dart';
import 'package:sys_share_sys_account_service/sys_share_sys_account_service.dart'
    as rpc;
import 'package:grpc/grpc_web.dart';

class UserRepo extends BaseRepo {
  // TODO @winwisely268: this is ugly, ideally we want flu side interceptor
  // as well so each request will have authorization metadata.
  static Future<rpc.Account> getUser({String id, String accessToken}) async {
    final req = rpc.GetAccountRequest()..id = id;

    try {
      final md = BaseRepo.callMetadata;
      md.putIfAbsent("Authorization", () => accessToken);
      final resp = await accountClient(accessToken)
          .getAccount(req, options: CallOptions(metadata: md))
          .then((res) {
        return res;
      });
      return resp;
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static rpc.AccountServiceClient accountClient(String accessToken) {
    return rpc.AccountServiceClient(BaseRepo.channel);
  }
}
