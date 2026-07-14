import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureSessionStore {
  SecureSessionStore([FlutterSecureStorage? storage])
      : _storage = storage ?? const FlutterSecureStorage();

  static const _tokenKey = 'pcs.mobile.bearer.v1';
  static const _companyKey = 'pcs.mobile.company.v1';
  final FlutterSecureStorage _storage;

  Future<String?> readToken() => _storage.read(key: _tokenKey);
  Future<void> saveToken(String token) => _storage.write(key: _tokenKey, value: token);
  Future<void> clearToken() => _storage.delete(key: _tokenKey);
  Future<String?> readCompanyId() => _storage.read(key: _companyKey);
  Future<void> saveCompanyId(int id) => _storage.write(key: _companyKey, value: '$id');
  Future<void> clearCompanyId() => _storage.delete(key: _companyKey);
  Future<void> clearAll() => _storage.deleteAll();
}
