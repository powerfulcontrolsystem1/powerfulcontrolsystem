import 'package:flutter/material.dart';

import '../routing/app_router.dart';

class PowerfulControlSystemApp extends StatelessWidget {
  const PowerfulControlSystemApp({super.key});

  @override
  Widget build(BuildContext context) => MaterialApp.router(
        title: 'Powerful Control System',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xff0a84d8), brightness: Brightness.light), useMaterial3: true),
        darkTheme: ThemeData(colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xff2aa8ff), brightness: Brightness.dark), useMaterial3: true),
        themeMode: ThemeMode.system,
        routerConfig: appRouter,
      );
}
