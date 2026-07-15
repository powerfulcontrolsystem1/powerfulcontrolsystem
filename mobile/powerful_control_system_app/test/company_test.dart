import 'package:flutter_test/flutter_test.dart';
import 'package:powerful_control_system_app/features/companies/domain/company.dart';

void main() {
  test('company maps only the safe mobile fields', () {
    final company = Company.fromJson(
        {'id': 12, 'nombre': 'PCS', 'tipo': 'Tecnología', 'estado': 'activo'});
    expect(company.id, 12);
    expect(company.name, 'PCS');
  });
}
