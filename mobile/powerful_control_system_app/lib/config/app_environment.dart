enum AppEnvironment { development, staging, production }

class AppEnvironmentConfig {
  const AppEnvironmentConfig({
    required this.environment,
    required this.apiBaseUrl,
    required this.requestTimeout,
    required this.enableNetworkLogs,
    required this.minimumBackendVersion,
  });

  final AppEnvironment environment;
  final Uri apiBaseUrl;
  final Duration requestTimeout;
  final bool enableNetworkLogs;
  final String minimumBackendVersion;

  factory AppEnvironmentConfig.fromDartDefines() {
    const rawEnvironment = String.fromEnvironment('PCS_ENV', defaultValue: 'production');
    const rawApiBaseUrl = String.fromEnvironment('PCS_API_BASE_URL', defaultValue: 'https://powerfulcontrolsystem.com');
    const timeoutSeconds = int.fromEnvironment('PCS_API_TIMEOUT_SECONDS', defaultValue: 20);
    const networkLogs = bool.fromEnvironment('PCS_NETWORK_LOGS', defaultValue: false);
    const minimumBackendVersion = String.fromEnvironment('PCS_MIN_BACKEND_VERSION', defaultValue: '1');
    final environment = switch (rawEnvironment) {
      'development' => AppEnvironment.development,
      'staging' => AppEnvironment.staging,
      _ => AppEnvironment.production,
    };
    final apiBaseUrl = Uri.tryParse(rawApiBaseUrl);
    final localDevelopmentHost = apiBaseUrl?.host == 'localhost' || apiBaseUrl?.host == '10.0.2.2';
    final allowsDevelopmentHttp = environment == AppEnvironment.development && apiBaseUrl?.scheme == 'http' && localDevelopmentHost;
    if (apiBaseUrl == null || apiBaseUrl.host.isEmpty || (apiBaseUrl.scheme != 'https' && !allowsDevelopmentHttp)) {
      throw ArgumentError('PCS_API_BASE_URL debe usar HTTPS, excepto localhost o 10.0.2.2 durante desarrollo.');
    }
    return AppEnvironmentConfig(
      environment: environment,
      apiBaseUrl: apiBaseUrl,
      requestTimeout: Duration(seconds: timeoutSeconds.clamp(5, 90)),
      enableNetworkLogs: networkLogs && environment != AppEnvironment.production,
      minimumBackendVersion: minimumBackendVersion,
    );
  }
}
