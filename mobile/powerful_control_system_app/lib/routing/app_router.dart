import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../authentication/presentation/login_screen.dart';
import '../authentication/presentation/session_guard.dart';
import '../authentication/presentation/startup_screen.dart';
import '../features/companies/presentation/company_selector_screen.dart';
import '../features/dashboard/presentation/dashboard_screen.dart';
import '../features/products/presentation/products_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(path: '/', builder: (context, state) => const StartupScreen()),
    GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
    GoRoute(path: '/companies', builder: (context, state) => const SessionGuard(child: CompanySelectorScreen())),
    GoRoute(path: '/dashboard', builder: (context, state) => const SessionGuard(child: DashboardScreen())),
    GoRoute(path: '/products', builder: (context, state) => const SessionGuard(child: ProductsScreen())),
  ],
  errorBuilder: (context, state) => Scaffold(body: Center(child: Text('No se encontró la pantalla solicitada: ${state.uri.path}'))),
);
