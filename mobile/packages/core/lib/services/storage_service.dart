import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Unified local storage service.
///
/// Uses [FlutterSecureStorage] for sensitive data (tokens, credentials)
/// and [SharedPreferences] for non-sensitive preferences (theme, locale,
/// onboarding status).
class StorageService {
  final FlutterSecureStorage _secure;
  SharedPreferences? _prefs;

  StorageService({FlutterSecureStorage? secure})
      : _secure = secure ?? const FlutterSecureStorage();

  /// Initialize SharedPreferences. Call once at app startup.
  Future<void> initialize() async {
    _prefs = await SharedPreferences.getInstance();
  }

  SharedPreferences get _preferences {
    if (_prefs == null) {
      throw StateError(
        'StorageService.initialize() must be called before accessing preferences.',
      );
    }
    return _prefs!;
  }

  // ---------------------------------------------------------------------------
  // Secure storage (tokens, secrets)
  // ---------------------------------------------------------------------------

  Future<void> writeSecure(String key, String value) async {
    await _secure.write(key: key, value: value);
  }

  Future<String?> readSecure(String key) async {
    return _secure.read(key: key);
  }

  Future<void> deleteSecure(String key) async {
    await _secure.delete(key: key);
  }

  Future<void> clearSecure() async {
    await _secure.deleteAll();
  }

  // ---------------------------------------------------------------------------
  // Preferences (non-sensitive)
  // ---------------------------------------------------------------------------

  /// Save a string preference.
  Future<bool> setString(String key, String value) {
    return _preferences.setString(key, value);
  }

  /// Read a string preference.
  String? getString(String key) {
    return _preferences.getString(key);
  }

  /// Save a boolean preference.
  Future<bool> setBool(String key, bool value) {
    return _preferences.setBool(key, value);
  }

  /// Read a boolean preference.
  bool? getBool(String key) {
    return _preferences.getBool(key);
  }

  /// Save an integer preference.
  Future<bool> setInt(String key, int value) {
    return _preferences.setInt(key, value);
  }

  /// Read an integer preference.
  int? getInt(String key) {
    return _preferences.getInt(key);
  }

  /// Remove a preference.
  Future<bool> remove(String key) {
    return _preferences.remove(key);
  }

  // ---------------------------------------------------------------------------
  // Convenience keys
  // ---------------------------------------------------------------------------

  static const String keyThemeMode = 'seva_theme_mode';
  static const String keyLocale = 'seva_locale';
  static const String keyOnboardingComplete = 'seva_onboarding_complete';
  static const String keyLastSyncTimestamp = 'seva_last_sync';

  /// Get the stored theme mode preference.
  String get themeMode => getString(keyThemeMode) ?? 'system';

  /// Set the theme mode preference.
  Future<void> setThemeMode(String mode) => setString(keyThemeMode, mode);

  /// Get the stored locale preference.
  String get locale => getString(keyLocale) ?? 'en';

  /// Set the locale preference.
  Future<void> setLocale(String locale) => setString(keyLocale, locale);

  /// Whether the user has completed onboarding.
  bool get isOnboardingComplete =>
      getBool(keyOnboardingComplete) ?? false;

  /// Mark onboarding as complete.
  Future<void> completeOnboarding() =>
      setBool(keyOnboardingComplete, true);
}
