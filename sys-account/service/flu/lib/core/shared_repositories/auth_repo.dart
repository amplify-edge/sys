import 'package:meta/meta.dart';
import 'package:grpc/grpc_web.dart';
import './utilities.dart';
import 'package:sys_share_sys_account_service/sys_share_sys_account_service.dart'
    as rpc;

class AuthRepo extends BaseRepo {
  static final client = _authClient();

  static Future<rpc.LoginResponse> loginUser(
      {String email, String password}) async {
    final req = rpc.LoginRequest()
      ..email = email
      ..password = password;

    try {
      final resp = await client
          .login(req, options: CallOptions(metadata: BaseRepo.callMetadata))
          .then((res) {
        print(res);
        return res;
      });
      return resp;
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static Future<rpc.RegisterResponse> registerAccount(
      {@required String email,
      @required String password,
      @required String passwordConfirm}) async {
    if (password != passwordConfirm) {
      rpc.RegisterResponse resp = rpc.RegisterResponse.getDefault()
        ..success = false
        ..errorReason = rpc.ErrorReason()
        ..errorReason.reason = "password mismatch";
      return resp;
    }

    try {
      final request = rpc.RegisterRequest()
        ..email = email
        ..password = password
        ..passwordConfirm = passwordConfirm;
      final resp = await client.register(request,
          options: CallOptions(metadata: BaseRepo.callMetadata));
      return resp;
    } catch (e) {
      throw e;
    }
  }

  static Future<rpc.ForgotPasswordResponse> forgotPassword(
      {@required String email}) async {
    final req = rpc.ForgotPasswordRequest()..email = email;
    try {
      return await client.forgotPassword(req,
          options: CallOptions(metadata: BaseRepo.callMetadata));
    } catch (e) {
      throw e;
    }
  }

  static Future<rpc.ResetPasswordResponse> resetPassword(
      {@required String email,
      @required String password,
      @required String passwordConfirm}) async {
    final req = rpc.ResetPasswordRequest()
      ..email = email
      ..password = password
      ..passwordConfirm = passwordConfirm;

    try {
      return await client.resetPassword(req,
          options: CallOptions(metadata: BaseRepo.callMetadata));
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static Future<rpc.RefreshAccessTokenResponse> renewAccessToken(
      {@required String refreshToken}) async {
    final req = rpc.RefreshAccessTokenRequest()..refreshToken = refreshToken;

    try {
      final resp = await client.refreshAccessToken(req,
          options: CallOptions(metadata: BaseRepo.callMetadata));
      return resp;
    } catch (e) {
      print(e);
      throw e;
    }
  }

  static rpc.AuthServiceClient _authClient() {
    return rpc.AuthServiceClient(BaseRepo.channel);
  }
}
