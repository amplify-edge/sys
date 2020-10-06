import 'package:grpc/grpc_web.dart';
import 'package:sys_share_sys_account_service/sys_share_sys_account_service.dart'
    as rpc;

class AuthRepo {
  static final client = _authClient();

  static Future<rpc.LoginResponse> loginUser(
      {String email, String password}) async {
    final req = rpc.LoginRequest()
      ..email = email
      ..password = password;

    try {
      final resp = await client.login(req).then((res) {
        print(res);
        return res;
      });
      return resp;
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static Future<rpc.RefreshAccessTokenResponse> renewAccessToken(
      {String refreshToken}) async {
    final req = rpc.RefreshAccessTokenRequest()..refreshToken = refreshToken;

    try {
      final resp = await client.refreshAccessToken(req);
      return resp;
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static rpc.AuthServiceClient _authClient() {
    final channel =
        GrpcWebClientChannel.xhr(Uri(host: "127.0.0.1", port: 8888));
    return rpc.AuthServiceClient(channel);
  }
}
//
