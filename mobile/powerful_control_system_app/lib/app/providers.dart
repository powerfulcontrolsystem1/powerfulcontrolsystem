import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../authentication/data/auth_repository.dart';
import '../authentication/domain/mobile_account.dart';
import '../config/app_environment.dart';
import '../core/network/api_client.dart';
import '../core/storage/secure_session_store.dart';
import '../features/companies/data/company_repository.dart';
import '../features/companies/domain/company.dart';
import '../features/products/data/product_repository.dart';
import '../features/products/domain/product.dart';

final environmentProvider = Provider((ref) => AppEnvironmentConfig.fromDartDefines());
final sessionStoreProvider = Provider((ref) => SecureSessionStore());
final apiClientProvider = Provider((ref) => ApiClient(ref.watch(environmentProvider), ref.watch(sessionStoreProvider)));
final authRepositoryProvider = Provider((ref) => AuthRepository(ref.watch(apiClientProvider), ref.watch(sessionStoreProvider)));
final companyRepositoryProvider = Provider((ref) => CompanyRepository(ref.watch(apiClientProvider)));
final productRepositoryProvider = Provider((ref) => ProductRepository(ref.watch(apiClientProvider)));
final currentAccountProvider = FutureProvider<MobileAccount?>((ref) => ref.watch(authRepositoryProvider).restore());
final companiesProvider = FutureProvider<List<Company>>((ref) => ref.watch(companyRepositoryProvider).list());
final selectedCompanyProvider = FutureProvider<Company?>((ref) async {
  final id = int.tryParse(await ref.watch(sessionStoreProvider).readCompanyId() ?? '');
  if (id == null) return null;
  final companies = await ref.watch(companiesProvider.future);
  for (final company in companies) {
    if (company.id == id) return company;
  }
  return null;
});
final productsProvider = FutureProvider.family<List<Product>, int>((ref, companyId) => ref.watch(productRepositoryProvider).list(companyId));
