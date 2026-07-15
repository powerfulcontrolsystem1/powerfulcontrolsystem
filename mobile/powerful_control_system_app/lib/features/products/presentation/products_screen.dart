import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers.dart';

class ProductsScreen extends ConsumerWidget {
  const ProductsScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final company = ref.watch(selectedCompanyProvider).value;
    if (company == null) {
      return const Scaffold(
          body: Center(
              child: Text('Selecciona una empresa para ver productos.')));
    }
    final products = ref.watch(productsProvider(company.id));
    return Scaffold(
        appBar: AppBar(title: Text('Productos · ${company.name}')),
        body: products.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (_, __) =>
              const Center(child: Text('No fue posible cargar el catálogo.')),
          data: (items) => RefreshIndicator(
              onRefresh: () async =>
                  ref.invalidate(productsProvider(company.id)),
              child: ListView.builder(
                  itemCount: items.length,
                  itemBuilder: (context, index) {
                    final product = items[index];
                    return ListTile(
                        title: Text(product.name),
                        subtitle: Text('Existencia: ${product.stock}'),
                        trailing: Text('\$${product.price}'));
                  })),
        ));
  }
}
