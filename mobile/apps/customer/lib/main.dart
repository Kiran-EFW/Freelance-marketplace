import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:seva_core/core.dart';
import 'app.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase.
  await Firebase.initializeApp();

  // Lock orientation to portrait on phones.
  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  // Set status bar style.
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
    ),
  );

  // Initialize storage service.
  final storageService = StorageService();
  await storageService.initialize();

  // Initialize API client.
  const baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://api.seva.app/v1',
  );
  final apiClient = ApiClient(baseUrl: baseUrl);

  // Initialize auth service.
  final authRepository = AuthRepository(api: apiClient);
  final authService = AuthService(authRepository: authRepository);
  await authService.initialize();

  // Initialize push notifications.
  final pushNotificationService = PushNotificationService(
    apiClient: apiClient,
  );
  pushNotificationService.onNavigate = (type, data) {
    // Navigation is handled via GoRouter after app is built.
    // Store the pending navigation for the router to pick up.
    debugPrint('Customer notification tap: $type');
  };
  if (authService.isAuthenticated) {
    await pushNotificationService.initialize();
  }

  // Initialize local storage and sync services.
  final localStorageService = LocalStorageService();
  await localStorageService.initialize();
  final syncService = SyncService(
    apiClient: apiClient,
    localStorage: localStorageService,
  );

  runApp(
    ProviderScope(
      overrides: [
        apiClientProvider.overrideWithValue(apiClient),
        authServiceProvider.overrideWithValue(authService),
        storageServiceProvider.overrideWithValue(storageService),
        pushNotificationServiceProvider
            .overrideWithValue(pushNotificationService),
        localStorageServiceProvider.overrideWithValue(localStorageService),
        syncServiceProvider.overrideWithValue(syncService),
      ],
      child: const SevaCustomerApp(),
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

final pushNotificationServiceProvider =
    Provider<PushNotificationService>((ref) {
  throw UnimplementedError('Must be overridden at app startup');
});

final localStorageServiceProvider = Provider<LocalStorageService>((ref) {
  throw UnimplementedError('Must be overridden at app startup');
});

final syncServiceProvider = Provider<SyncService>((ref) {
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
