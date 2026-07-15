class ApiException implements Exception {
  const ApiException(this.code, this.message,
      {this.requestId, this.statusCode});

  final String code;
  final String message;
  final String? requestId;
  final int? statusCode;

  bool get isUnauthenticated => code == 'unauthenticated' || statusCode == 401;
  bool get isTwoFactorRequired => code == 'two_factor_required';

  @override
  String toString() => 'ApiException($code, requestId: $requestId)';
}
