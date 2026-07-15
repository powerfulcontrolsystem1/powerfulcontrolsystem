import '../../../core/network/api_client.dart';
import '../domain/company.dart';

class CompanyRepository {
  CompanyRepository(this._client);
  final ApiClient _client;

  Future<List<Company>> list() async {
    final response = await _client.get('/api/v1/empresas');
    final values = (response['data'] as List<dynamic>? ?? const []);
    return values
        .map((item) => Company.fromJson(item as Map<String, dynamic>))
        .where((item) => item.id > 0)
        .toList(growable: false);
  }
}
