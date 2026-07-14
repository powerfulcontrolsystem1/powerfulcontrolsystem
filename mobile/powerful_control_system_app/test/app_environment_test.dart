import 'package:flutter_test/flutter_test.dart';
import 'package:powerful_control_system_app/config/app_environment.dart';

void main() {
  test('production defaults to the official HTTPS endpoint', () {
    final config = AppEnvironmentConfig.fromDartDefines();
    expect(config.apiBaseUrl.scheme, 'https');
    expect(config.apiBaseUrl.host, 'powerfulcontrolsystem.com');
  });
}
