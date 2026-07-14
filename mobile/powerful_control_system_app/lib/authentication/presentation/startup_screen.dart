import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/providers.dart';

class StartupScreen extends ConsumerStatefulWidget {
  const StartupScreen({super.key});

  @override
  ConsumerState<StartupScreen> createState() => _StartupScreenState();
}

class _StartupScreenState extends ConsumerState<StartupScreen> {
  @override
  void initState() {
    super.init();
    Future.microtask(_restore);
  }

  Future<void> _restore() async {
    final account = await ref.read(currentAccountProvider.future).catchError((_) => null);
    if (!mounted) return;
    if (account == null) {
      context.go('/login');
      return;
    }
    final company = await ref.read(selectedCompanyProvider.future).catchError((_) => null);
    if (!mounted) return;
    context.go(company == null ? '/companies' : '/dashboard');
  }

  @override
  Widget build(BuildContext context) => const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
}
