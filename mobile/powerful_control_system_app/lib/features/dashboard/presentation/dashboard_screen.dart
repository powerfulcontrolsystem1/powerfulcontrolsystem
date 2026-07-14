import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/providers.dart';
import '../../../core/connectivity/connectivity_controller.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final company = ref.watch(selectedCompanyProvider);
    final online = ref.watch(connectivityProvider).valueOrNull ?? true;
    return Scaffold(
      appBar: AppBar(title: const Text('Panel móvil'), actions: [IconButton(onPressed: () async { await ref.read(authRepositoryProvider).logout(); ref.invalidate(currentAccountProvider); ref.invalidate(companiesProvider); ref.invalidate(selectedCompanyProvider); if (context.mounted) context.go('/login'); }, icon: const Icon(Icons.logout), tooltip: 'Cerrar sesión')]),
      body: company.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => const Center(child: Text('No fue posible recuperar la empresa activa.')),
        data: (item) => ListView(padding: const EdgeInsets.all(16), children: [
          if (!online) const Card(child: ListTile(leading: Icon(Icons.cloud_off), title: Text('Sin conexión'), subtitle: Text('Las operaciones pendientes se confirmarán solamente con el backend.'))),
          Text(item?.name ?? 'Selecciona una empresa', style: Theme.of(context).textTheme.headlineSmall),
          const SizedBox(height: 16),
          Card(child: ListTile(leading: const Icon(Icons.inventory_2_outlined), title: const Text('Productos'), subtitle: const Text('Catálogo empresarial con permisos del servidor.'), onTap: item == null ? null : () => context.go('/products'))),
          Card(child: ListTile(leading: const Icon(Icons.business_outlined), title: const Text('Cambiar empresa'), onTap: () => context.go('/companies'))),
        ]),
      ),
    );
  }
}
