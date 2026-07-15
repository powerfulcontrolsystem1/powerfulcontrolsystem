import '../../core/network/api_client.dart';
import '../../core/errors/api_exception.dart';
import '../../core/storage/secure_session_store.dart';
import '../domain/mobile_account.dart';

class AuthRepository {
  AuthRepository(this._client, this._sessions);

  final ApiClient _client;
  final SecureSessionStore _sessions;

  Future<MobileAccount> login(
      {required String email,
      required String password,
      String otpCode = '',
      String deviceName = 'PCS Mobile'}) async {
    final response = await _client.post('/api/v1/auth/login', body: {
      'email': email,
      'password': password,
      'otp_code': otpCode,
      'device_name': deviceName
    });
    final data = response['data'] as Map<String, dynamic>;
    await _sessions.saveToken(data['access_token'] as String);
    return MobileAccount.fromJson(data['account'] as Map<String, dynamic>);
  }

  Future<MobileAccount?> restore() async {
    if ((await _sessions.readToken())?.isEmpty ?? true) return null;
    try {
      final response = await _client.get('/api/v1/me');
      return MobileAccount.fromJson(response['data'] as Map<String, dynamic>);
    } on ApiException catch (error) {
      if (error.statusCode == 401) {
        await _sessions.clearAll();
        return null;
      }
      rethrow;
    }
  }

  Future<void> logout() async {
    try {
      await _client.post('/api/v1/auth/logout');
    } finally {
      await _sessions.clearAll();
    }
  }
}
