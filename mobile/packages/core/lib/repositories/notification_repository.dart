import 'package:dio/dio.dart';

import '../api/api_client.dart';
import '../models/notification.dart';
import '../models/result.dart';
import 'job_repository.dart';

/// Handles notification API interactions and push token registration.
class NotificationRepository {
  final ApiClient _api;

  NotificationRepository({required ApiClient api}) : _api = api;

  /// Fetch paginated notifications for the current user.
  Future<Result<PaginatedResult<AppNotification>>> getNotifications({
    int page = 1,
    int limit = 50,
  }) async {
    try {
      final response = await _api.getNotifications(page: page, limit: limit);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load notifications'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load notifications: $e');
    }
  }

  /// Mark a single notification as read.
  Future<Result<bool>> markAsRead(String notificationId) async {
    try {
      await _api.markNotificationRead(notificationId);
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to mark notification as read'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to mark notification as read: $e');
    }
  }

  /// Mark all notifications as read.
  Future<Result<bool>> markAllAsRead() async {
    try {
      await _api.markAllNotificationsRead();
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to mark all notifications as read'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to mark all notifications as read: $e');
    }
  }

  /// Register a push notification token for the current device.
  Future<Result<bool>> registerPushToken(String token, String platform) async {
    try {
      await _api.registerPushToken(token, platform);
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to register push token'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to register push token: $e');
    }
  }

  /// Count unread notifications.
  Future<Result<int>> getUnreadCount() async {
    try {
      final result = await getNotifications(page: 1, limit: 1);
      switch (result) {
        case Success(:final data):
          final count = data.items.where((n) => !n.isRead).length;
          return Success(count);
        case Failure(:final message, :final statusCode):
          return Failure(message, statusCode: statusCode);
      }
    } catch (e) {
      return Failure('Failed to get unread count: $e');
    }
  }

  /// Extract a human-readable error message from a [DioException].
  String _extractErrorMessage(DioException e, String fallback) {
    final data = e.response?.data;
    if (data is Map<String, dynamic>) {
      final message = data['message'] ?? data['error'];
      if (message is String && message.isNotEmpty) return message;
    }

    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return 'Connection timed out. Please check your network.';
      case DioExceptionType.connectionError:
        return 'No internet connection. Please try again later.';
      default:
        return fallback;
    }
  }
}
