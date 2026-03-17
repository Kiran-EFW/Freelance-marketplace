import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:seva_core/core.dart';
import 'app.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
    ),
  );

  final storageService = StorageService();
  await storageService.initialize();

  const baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://api.seva.app/v1',
  );
  final apiClient = ApiClient(baseUrl: baseUrl);

  final authRepository = AuthRepository(api: apiClient);
  final authService = AuthService(authRepository: authRepository);
  await authService.initialize();

  runApp(
    ProviderScope(
      overrides: [
        apiClientProvider.overrideWithValue(apiClient),
        authServiceProvider.overrideWithValue(authService),
        storageServiceProvider.overrideWithValue(storageService),
      ],
      child: const SevaProviderApp(),
    ),
  );
}

// ---------------------------------------------------------------------------
// Riverpod providers for dependency injection
// ---------------------------------------------------------------------------

final apiClientProvider = Provider<ApiClient>((ref) {
  throw UnimplementedError('Must be overridden at app startup');
});

final authServiceProvider = Provider<AuthService>((ref) {
  throw UnimplementedError('Must be overridden at app startup');
});

final storageServiceProvider = Provider<StorageService>((ref) {
  throw UnimplementedError('Must be overridden at app startup');
});

final locationServiceProvider = Provider<LocationService>((ref) {
  return LocationService();
});

final jobRepositoryProvider = Provider<JobRepository>((ref) {
  return JobRepository(api: ref.watch(apiClientProvider));
});

final providerRepositoryProvider = Provider<ProviderRepository>((ref) {
  return ProviderRepository(api: ref.watch(apiClientProvider));
});

final notificationRepositoryProvider = Provider<NotificationRepository>((ref) {
  return NotificationRepository(api: ref.watch(apiClientProvider));
});
