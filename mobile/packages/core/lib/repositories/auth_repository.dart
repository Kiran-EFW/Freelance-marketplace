import '../api/api_client.dart';
import '../models/user.dart';

/// Handles all authentication-related API interactions.
///
/// Manages OTP-based login, registration, token persistence,
/// and session lifecycle.
class AuthRepository {
  final ApiClient _api;

  AuthRepository({required ApiClient api}) : _api = api;

  /// Request an OTP code sent to the given phone number.
  /// Returns true if the OTP was sent successfully.
  Future<bool> requestOtp(String phone) async {
    try {
      final response = await _api.requestOtp(phone: phone);
      return response.statusCode == 200;
    } catch (_) {
      return false;
    }
  }

  /// Verify the OTP code. On success, persists the JWT tokens
  /// and returns the authenticated [User]. Returns null on failure.
  Future<User?> verifyOtp(String phone, String code) async {
    try {
      final response = await _api.verifyOtp(phone: phone, code: code);
      final data = response.data as Map<String, dynamic>;

      await _api.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );

      return User.fromJson(data['user'] as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Register a new user account.
  Future<User?> register({
    required String name,
    required String phone,
    required String role,
    String? email,
    String? postcode,
  }) async {
    try {
      final response = await _api.register(
        name: name,
        phone: phone,
        role: role,
        email: email,
        postcode: postcode,
      );
      final data = response.data as Map<String, dynamic>;

      await _api.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );

      return User.fromJson(data['user'] as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Attempt to restore the session using a stored token.
  /// Returns the current user if the token is still valid, null otherwise.
  Future<User?> getCurrentUser() async {
    final hasToken = await _api.hasValidToken;
    if (!hasToken) return null;

    try {
      final response = await _api.getMe();
      return User.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Update the current user's profile.
  Future<User?> updateProfile(
    String userId,
    Map<String, dynamic> updates,
  ) async {
    try {
      final response = await _api.updateUser(userId, updates);
      return User.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Upload a new avatar image for the user.
  Future<String?> uploadAvatar(String userId, String filePath) async {
    try {
      final response = await _api.uploadAvatar(userId, filePath);
      return response.data['avatar_url'] as String?;
    } catch (_) {
      return null;
    }
  }

  /// Sign out: clear stored tokens.
  Future<void> signOut() async {
    await _api.clearTokens();
  }
}
