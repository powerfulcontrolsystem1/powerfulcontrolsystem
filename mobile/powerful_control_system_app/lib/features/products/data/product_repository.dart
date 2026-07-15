import '../../../core/network/api_client.dart';
import '../domain/product.dart';

class ProductRepository {
  ProductRepository(this._client);
  final ApiClient _client;

  Future<List<Product>> list(int companyId, {String query = ''}) async {
    final response = await _client.get('/api/v1/empresa/productos', query: {
      'empresa_id': '$companyId',
      'limit': '50',
      'q': query,
      'fields': 'id,nombre,precio_venta,stock_actual'
    });
    final values = (response['data'] as List<dynamic>? ?? const []);
    return values
        .map((item) => Product.fromJson(item as Map<String, dynamic>))
        .toList(growable: false);
  }
}
