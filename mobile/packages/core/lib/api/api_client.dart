import 'dart:async';
import 'dart:ui';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Central HTTP client for all Seva API communication.
///
/// Wraps Dio with JWT token management, automatic token refresh on 401,
/// and typed endpoint methods for every backend route.
class ApiClient {
  late final Dio _dio;
  final FlutterSecureStorage _storage;

  static const String _accessTokenKey = 'seva_access_token';
  static const String _refreshTokenKey = 'seva_refresh_token';

  /// Whether a token refresh is currently in flight, used to queue
  /// concurrent requests that hit 401 simultaneously.
  bool _isRefreshing = false;
  final List<Function(String)> _pendingRequests = [];

  /// Called when token refresh fails and user must re-authenticate.
  /// Set this to [AuthService.forceSignOut] after construction.
  VoidCallback? onTokenExpired;

  ApiClient({
    required String baseUrl,
    FlutterSecureStorage? storage,
  }) : _storage = storage ?? const FlutterSecureStorage() {
    _dio = Dio(
      BaseOptions(
        baseUrl: baseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 30),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: _onRequest,
        onError: _onError,
      ),
    );
  }

  /// Attach JWT token to every outgoing request.
  Future<void> _onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _storage.read(key: _accessTokenKey);
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  /// Intercept 401 errors and attempt a single token refresh.
  /// All concurrent 401s are queued so we only refresh once.
  Future<void> _onError(
    DioException error,
    ErrorInterceptorHandler handler,
  ) async {
    if (error.response?.statusCode != 401) {
      return handler.next(error);
    }

    final refreshToken = await _storage.read(key: _refreshTokenKey);
    if (refreshToken == null) {
      await clearTokens();
      return handler.next(error);
    }

    if (_isRefreshing) {
      // Queue this request; it will be retried once the refresh completes.
      final completer = Completer<Response>();
      _pendingRequests.add((newToken) async {
        error.requestOptions.headers['Authorization'] = 'Bearer $newToken';
        try {
          final response = await _dio.fetch(error.requestOptions);
          completer.complete(response);
        } catch (e) {
          completer.completeError(e);
        }
      });
      try {
        final response = await completer.future;
        return handler.resolve(response);
      } catch (e) {
        return handler.next(error);
      }
    }

    _isRefreshing = true;

    try {
      final response = await _dio.post(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );

      final newAccessToken = response.data['access_token'] as String;
      final newRefreshToken = response.data['refresh_token'] as String;

      await _storage.write(key: _accessTokenKey, value: newAccessToken);
      await _storage.write(key: _refreshTokenKey, value: newRefreshToken);

      // Retry the original request with the new token.
      error.requestOptions.headers['Authorization'] = 'Bearer $newAccessToken';
      final retryResponse = await _dio.fetch(error.requestOptions);

      // Flush queued requests.
      for (final pending in _pendingRequests) {
        pending(newAccessToken);
      }
      _pendingRequests.clear();

      return handler.resolve(retryResponse);
    } on DioException {
      await clearTokens();
      _pendingRequests.clear();
      onTokenExpired?.call();
      return handler.next(error);
    } finally {
      _isRefreshing = false;
    }
  }

  // ---------------------------------------------------------------------------
  // Token helpers
  // ---------------------------------------------------------------------------

  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    await _storage.write(key: _accessTokenKey, value: accessToken);
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  Future<void> clearTokens() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }

  Future<bool> get hasValidToken async {
    final token = await _storage.read(key: _accessTokenKey);
    return token != null;
  }

  // ---------------------------------------------------------------------------
  // Auth endpoints
  // ---------------------------------------------------------------------------

  /// Request a one-time password sent via SMS to [phone].
  Future<Response> requestOtp({required String phone}) {
    return _dio.post('/auth/otp/request', data: {'phone': phone});
  }

  /// Verify the OTP for the given [phone].
  Future<Response> verifyOtp({
    required String phone,
    required String code,
  }) {
    return _dio.post('/auth/otp/verify', data: {
      'phone': phone,
      'code': code,
    });
  }

  /// Register a new user account.
  Future<Response> register({
    required String name,
    required String phone,
    required String role,
    String? email,
    String? postcode,
  }) {
    return _dio.post('/auth/register', data: {
      'name': name,
      'phone': phone,
      'role': role,
      if (email != null) 'email': email,
      if (postcode != null) 'postcode': postcode,
    });
  }

  /// Fetch the currently authenticated user's profile.
  Future<Response> getMe() {
    return _dio.get('/auth/me');
  }

  /// Refresh the access token using a refresh token.
  Future<Response> refreshToken({required String refreshToken}) {
    return _dio.post('/auth/refresh', data: {'refresh_token': refreshToken});
  }

  // ---------------------------------------------------------------------------
  // User endpoints
  // ---------------------------------------------------------------------------

  Future<Response> getUser(String userId) {
    return _dio.get('/users/$userId');
  }

  Future<Response> updateUser(String userId, Map<String, dynamic> data) {
    return _dio.patch('/users/$userId', data: data);
  }

  Future<Response> uploadAvatar(String userId, String filePath) {
    return _dio.post(
      '/users/$userId/avatar',
      data: FormData.fromMap({
        'file': MultipartFile.fromFileSync(filePath),
      }),
    );
  }

  // ---------------------------------------------------------------------------
  // Category endpoints
  // ---------------------------------------------------------------------------

  Future<Response> getCategories({String? parentId}) {
    return _dio.get('/categories', queryParameters: {
      if (parentId != null) 'parent_id': parentId,
    });
  }

  Future<Response> getCategory(String categoryId) {
    return _dio.get('/categories/$categoryId');
  }

  // ---------------------------------------------------------------------------
  // Provider endpoints
  // ---------------------------------------------------------------------------

  Future<Response> searchProviders({
    String? query,
    String? categoryId,
    double? latitude,
    double? longitude,
    int? radiusKm,
    double? minRating,
    int page = 1,
    int limit = 20,
    String? sortBy,
  }) {
    return _dio.get('/providers', queryParameters: {
      if (query != null) 'q': query,
      if (categoryId != null) 'category_id': categoryId,
      if (latitude != null) 'lat': latitude,
      if (longitude != null) 'lng': longitude,
      if (radiusKm != null) 'radius_km': radiusKm,
      if (minRating != null) 'min_rating': minRating,
      'page': page,
      'limit': limit,
      if (sortBy != null) 'sort_by': sortBy,
    });
  }

  Future<Response> getProvider(String providerId) {
    return _dio.get('/providers/$providerId');
  }

  Future<Response> getProviderReviews(String providerId, {int page = 1}) {
    return _dio.get('/providers/$providerId/reviews', queryParameters: {
      'page': page,
    });
  }

  Future<Response> updateProviderProfile(
    String providerId,
    Map<String, dynamic> data,
  ) {
    return _dio.patch('/providers/$providerId', data: data);
  }

  Future<Response> updateAvailability(
    String providerId,
    Map<String, dynamic> availability,
  ) {
    return _dio.put('/providers/$providerId/availability', data: availability);
  }

  // ---------------------------------------------------------------------------
  // Job endpoints
  // ---------------------------------------------------------------------------

  Future<Response> createJob(Map<String, dynamic> data) {
    return _dio.post('/jobs', data: data);
  }

  Future<Response> getJob(String jobId) {
    return _dio.get('/jobs/$jobId');
  }

  Future<Response> getJobs({
    String? status,
    String? role,
    int page = 1,
    int limit = 20,
  }) {
    return _dio.get('/jobs', queryParameters: {
      if (status != null) 'status': status,
      if (role != null) 'role': role,
      'page': page,
      'limit': limit,
    });
  }

  Future<Response> updateJobStatus(String jobId, String status) {
    return _dio.patch('/jobs/$jobId/status', data: {'status': status});
  }

  Future<Response> acceptJob(String jobId) {
    return _dio.post('/jobs/$jobId/accept');
  }

  Future<Response> declineJob(String jobId, {String? reason}) {
    return _dio.post('/jobs/$jobId/decline', data: {
      if (reason != null) 'reason': reason,
    });
  }

  Future<Response> completeJob(String jobId, {List<String>? photoUrls}) {
    return _dio.post('/jobs/$jobId/complete', data: {
      if (photoUrls != null) 'photo_urls': photoUrls,
    });
  }

  Future<Response> cancelJob(String jobId, {required String reason}) {
    return _dio.post('/jobs/$jobId/cancel', data: {'reason': reason});
  }

  Future<Response> submitReview(
    String jobId, {
    required int rating,
    String? comment,
  }) {
    return _dio.post('/jobs/$jobId/review', data: {
      'rating': rating,
      if (comment != null) 'comment': comment,
    });
  }

  Future<Response> uploadJobPhotos(String jobId, List<String> filePaths) {
    return _dio.post(
      '/jobs/$jobId/photos',
      data: FormData.fromMap({
        'files': filePaths
            .map((path) => MultipartFile.fromFileSync(path))
            .toList(),
      }),
    );
  }

  // ---------------------------------------------------------------------------
  // Route endpoints (provider-only)
  // ---------------------------------------------------------------------------

  Future<Response> getRoutes({int page = 1}) {
    return _dio.get('/routes', queryParameters: {'page': page});
  }

  Future<Response> getRoute(String routeId) {
    return _dio.get('/routes/$routeId');
  }

  Future<Response> optimizeRoute(String routeId) {
    return _dio.post('/routes/$routeId/optimize');
  }

  // ---------------------------------------------------------------------------
  // Notification endpoints
  // ---------------------------------------------------------------------------

  Future<Response> getNotifications({int page = 1, int limit = 50}) {
    return _dio.get('/notifications', queryParameters: {
      'page': page,
      'limit': limit,
    });
  }

  Future<Response> markNotificationRead(String notificationId) {
    return _dio.patch('/notifications/$notificationId/read');
  }

  Future<Response> markAllNotificationsRead() {
    return _dio.post('/notifications/read-all');
  }

  Future<Response> registerPushToken(String token, String platform) {
    return _dio.post('/notifications/push-token', data: {
      'token': token,
      'platform': platform,
    });
  }

  // ---------------------------------------------------------------------------
  // Earnings endpoints (provider-only)
  // ---------------------------------------------------------------------------

  Future<Response> getEarnings({String? period, String? from, String? to}) {
    return _dio.get('/earnings', queryParameters: {
      if (period != null) 'period': period,
      if (from != null) 'from': from,
      if (to != null) 'to': to,
    });
  }

  Future<Response> getPayoutHistory({int page = 1}) {
    return _dio.get('/earnings/payouts', queryParameters: {'page': page});
  }

  Future<Response> requestPayout({required double amount}) {
    return _dio.post('/earnings/payout', data: {'amount': amount});
  }

  // ---------------------------------------------------------------------------
  // Messaging endpoints
  // ---------------------------------------------------------------------------

  /// List conversations for the current user.
  Future<Response> getConversations({int page = 1, int limit = 20}) {
    return _dio.get('/messages/conversations', queryParameters: {
      'page': page,
      'limit': limit,
    });
  }

  /// Get messages in a conversation.
  Future<Response> getMessages(String conversationId, {int page = 1}) {
    return _dio.get('/messages/conversations/$conversationId', queryParameters: {
      'page': page,
    });
  }

  /// Send a message in a conversation.
  Future<Response> sendMessage(String conversationId, String content) {
    return _dio.post('/messages/conversations/$conversationId', data: {
      'content': content,
    });
  }

  /// Create a new conversation with a provider.
  Future<Response> createConversation(String providerId, {String? jobId}) {
    return _dio.post('/messages/conversations', data: {
      'provider_id': providerId,
      if (jobId != null) 'job_id': jobId,
    });
  }

  /// Mark all messages in a conversation as read.
  Future<Response> markMessagesRead(String conversationId) {
    return _dio.put('/messages/conversations/$conversationId/read');
  }

  // ---------------------------------------------------------------------------
  // Photo analysis
  // ---------------------------------------------------------------------------

  Future<Response> analyzePhoto(String filePath) {
    return _dio.post(
      '/ai/analyze-photo',
      data: FormData.fromMap({
        'file': MultipartFile.fromFileSync(filePath),
      }),
    );
  }
}
