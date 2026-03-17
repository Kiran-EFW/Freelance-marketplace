import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import 'main.dart';
import 'router.dart';

/// Root widget for the Seva Provider app.
class SevaProviderApp extends ConsumerWidget {
  const SevaProviderApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final storageService = ref.watch(storageServiceProvider);
    final themeMode = _resolveThemeMode(storageService.themeMode);
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Seva Provider',
      debugShowCheckedModeBanner: false,
      theme: SevaTheme.light,
      darkTheme: SevaTheme.dark,
      themeMode: themeMode,
      routerConfig: router,
    );
  }

  ThemeMode _resolveThemeMode(String mode) {
    switch (mode) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }
}
