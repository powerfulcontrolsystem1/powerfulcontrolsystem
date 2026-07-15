import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/providers.dart';

class CompanySelectorScreen extends ConsumerWidget {
  const CompanySelectorScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final companies = ref.watch(companiesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Seleccionar empresa')),
      body: companies.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => const Center(
            child: Text('No fue posible cargar las empresas autorizadas.')),
        data: (items) => items.isEmpty
            ? const Center(
                child: Text('No tienes empresas activas disponibles.'))
            : ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: items.length,
                separatorBuilder: (_, __) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final company = items[index];
                  return Card(
                      child: ListTile(
                          leading: const Icon(Icons.business),
                          title: Text(company.name),
                          subtitle: Text(company.type.isEmpty
                              ? 'Empresa activa'
                              : company.type),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () async {
                            await ref
                                .read(sessionStoreProvider)
                                .saveCompanyId(company.id);
                            if (context.mounted) context.go('/dashboard');
                          }));
                },
              ),
      ),
    );
  }
}
