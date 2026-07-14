import 'dart:convert';

import 'package:http/http.dart' as http;

import '../../config/app_environment.dart';
import '../errors/api_exception.dart';
import '../storage/secure_session_store.dart';

class ApiClient {
  ApiClient(this._config, this._sessions, {http.Client? httpClient}) : _http = httpClient ?? http.Client();

  final AppEnvironmentConfig _config;
  final SecureSessionStore _sessions;
  final http.Client _http;
  Future<String?>? _refreshInFlight;

  Uri uri(String path, [Map<String, String>? query]) => _config.apiBaseUrl.replace(
        path: path,
        queryParameters: query == null || query.isEmpty ? null : query,
      );

  Future<Map<String, dynamic>> get(String path, {Map<String, String>? query}) => _request('GET', path, query: query);
  Future<Map<String, dynamic>> post(String path, {Object? body, Map<String, String>? headers}) => _request('POST', path, body: body, headers: headers);

  Future<Map<String, dynamic>> _request(
    String method,
    String path, {
    Map<String, String>? query,
    Object? body,
    Map<String, String>? headers,
    bool allowRefresh = true,
  }) async {
    final token = await _sessions.readToken();
    final requestHeaders = <String, String>{'Accept': 'application/json', ...?headers};
    if (body != null) requestHeaders['Content-Type'] = 'application/json';
    if (token != null && token.isNotEmpty) requestHeaders['Authorization'] = 'Bearer $token';
    final request = http.Request(method, uri(path, query))
      ..headers.addAll(requestHeaders)
      ..body = body == null ? '' : jsonEncode(body);
    try {
      final streamed = await _http.send(request).timeout(_config.requestTimeout);
      final response = await http.Response.fromStream(streamed);
      final decoded = response.body.isEmpty ? <String, dynamic>{} : jsonDecode(response.body) as Map<String, dynamic>;
      if (response.statusCode == 401 && allowRefresh && token != null && token.isNotEmpty && path != '/api/v1/auth/refresh') {
        final refreshed = await _refreshAccessToken(token);
        if (refreshed != null) {
          return _request(method, path, query: query, body: body, headers: headers, allowRefresh: false);
        }
      }
      if (response.statusCode < 200 || response.statusCode >= 300 || decoded['ok'] != true) {
        final error = decoded['error'] as Map<String, dynamic>?;
        throw ApiException(error?['code']?.toString() ?? 'request_failed', error?['message']?.toString() ?? 'No fue posible completar la solicitud.', requestId: decoded['request_id']?.toString(), statusCode: response.statusCode);
      }
      return decoded;
    } on ApiException {
      rethrow;
    } on Exception {
      throw const ApiException('network_unavailable', 'No fue posible conectarse con Powerful Control System.');
    }
  }

  Future<String?> _refreshAccessToken(String currentToken) {
    return _refreshInFlight ??= _performRefresh(currentToken).whenComplete(() => _refreshInFlight = null);
  }

  Future<String?> _performRefresh(String currentToken) async {
    try {
      final request = http.Request('POST', uri('/api/v1/auth/refresh'))
        ..headers.addAll(<String, String>{
          'Accept': 'application/json',
          'Authorization': 'Bearer $currentToken',
        });
      final streamed = await _http.send(request).timeout(_config.requestTimeout);
      final response = await http.Response.fromStream(streamed);
      if (response.statusCode < 200 || response.statusCode >= 300 || response.body.isEmpty) {
        return null;
      }
      final decoded = jsonDecode(response.body) as Map<String, dynamic>;
      final data = decoded['data'] as Map<String, dynamic>?;
      final replacement = data?['access_token']?.toString();
      if (decoded['ok'] != true || replacement == null || replacement.length < 32) {
        return null;
      }
      await _sessions.writeToken(replacement);
      return replacement;
    } on Exception {
      return null;
    }
  }
}
