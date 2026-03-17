import 'dart:async';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../api/api_client.dart';
import '../models/user.dart';
import '../repositories/auth_repository.dart';

/// Authentication state exposed by [AuthService].
enum AuthState {
  /// Initial state before any check has been performed.
  unknown,

  /// The user is authenticated and has a valid session.
  authenticated,

  /// The user is not authenticated (no token or token expired).
  unauthenticated,
}

/// Manages authentication lifecycle: OTP login, auto-login on cold start,
/// session monitoring, and sign-out.
class AuthService {
  final AuthRepository _authRepository;
  final FlutterSecureStorage _storage;

  static const String _lastPhoneKey = 'seva_last_phone';

  User? _currentUser;
  AuthState _state = AuthState.unknown;

  final StreamController<AuthState> _authStateController =
      StreamController<AuthState>.broadcast();

  final StreamController<User?> _userController =
      StreamController<User?>.broadcast();

  AuthService({
    required AuthRepository authRepository,
    FlutterSecureStorage? storage,
  })  : _authRepository = authRepository,
        _storage = storage ?? const FlutterSecureStorage();

  /// Stream of authentication state changes.
  Stream<AuthState> get authStateStream => _authStateController.stream;

  /// Stream of user changes.
  Stream<User?> get userStream => _userController.stream;

  /// The currently authenticated user, or null.
  User? get currentUser => _currentUser;

  /// The current authentication state.
  AuthState get state => _state;

  /// Whether the user is currently authenticated.
  bool get isAuthenticated => _state == AuthState.authenticated;

  /// Attempt to restore a session from stored tokens.
  /// Call this at app startup.
  Future<void> initialize() async {
    try {
      final user = await _authRepository.getCurrentUser();
      if (user != null) {
        _setAuthenticated(user);
      } else {
        _setUnauthenticated();
      }
    } catch (_) {
      _setUnauthenticated();
    }
  }

  /// Request an OTP for the given phone number.
  /// Stores the phone number for convenience on the verify screen.
  Future<bool> requestOtp(String phone) async {
    await _storage.write(key: _lastPhoneKey, value: phone);
    return _authRepository.requestOtp(phone);
  }

  /// Verify the OTP. On success, the user is signed in.
  Future<User?> verifyOtp(String phone, String code) async {
    final user = await _authRepository.verifyOtp(phone, code);
    if (user != null) {
      _setAuthenticated(user);
    }
    return user;
  }

  /// Register a new account. On success, the user is signed in.
  Future<User?> register({
    required String name,
    required String phone,
    required String role,
    String? email,
    String? postcode,
  }) async {
    final user = await _authRepository.register(
      name: name,
      phone: phone,
      role: role,
      email: email,
      postcode: postcode,
    );
    if (user != null) {
      _setAuthenticated(user);
    }
    return user;
  }

  /// Update the current user's profile.
  Future<User?> updateProfile(Map<String, dynamic> updates) async {
    if (_currentUser == null) return null;
    final user = await _authRepository.updateProfile(
      _currentUser!.id,
      updates,
    );
    if (user != null) {
      _currentUser = user;
      _userController.add(user);
    }
    return user;
  }

  /// Upload a new avatar.
  Future<String?> uploadAvatar(String filePath) async {
    if (_currentUser == null) return null;
    return _authRepository.uploadAvatar(_currentUser!.id, filePath);
  }

  /// Retrieve the last phone number used for OTP.
  Future<String?> getLastPhone() async {
    return _storage.read(key: _lastPhoneKey);
  }

  /// Sign out: clear tokens and reset state.
  Future<void> signOut() async {
    await _authRepository.signOut();
    _setUnauthenticated();
  }

  /// Dispose of stream controllers.
  void dispose() {
    _authStateController.close();
    _userController.close();
  }

  void _setAuthenticated(User user) {
    _currentUser = user;
    _state = AuthState.authenticated;
    _authStateController.add(AuthState.authenticated);
    _userController.add(user);
  }

  void _setUnauthenticated() {
    _currentUser = null;
    _state = AuthState.unauthenticated;
    _authStateController.add(AuthState.unauthenticated);
    _userController.add(null);
  }
}
