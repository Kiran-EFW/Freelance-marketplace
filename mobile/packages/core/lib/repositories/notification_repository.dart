import '../api/api_client.dart';
import '../models/notification.dart';
import 'job_repository.dart';

/// Handles notification API interactions and push token registration.
class NotificationRepository {
  final ApiClient _api;

  NotificationRepository({required ApiClient api}) : _api = api;

  /// Fetch paginated notifications for the current user.
  Future<PaginatedResult<AppNotification>> getNotifications({
    int page = 1,
    int limit = 50,
  }) async {
    try {
      final response = await _api.getNotifications(page: page, limit: limit);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
          .toList();

      return PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      );
    } catch (_) {
      return const PaginatedResult(
        items: [],
        total: 0,
        page: 1,
        totalPages: 1,
      );
    }
  }

  /// Mark a single notification as read.
  Future<bool> markAsRead(String notificationId) async {
    try {
      await _api.markNotificationRead(notificationId);
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Mark all notifications as read.
  Future<bool> markAllAsRead() async {
    try {
      await _api.markAllNotificationsRead();
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Register a push notification token for the current device.
  Future<bool> registerPushToken(String token, String platform) async {
    try {
      await _api.registerPushToken(token, platform);
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Count unread notifications.
  Future<int> getUnreadCount() async {
    try {
      final result = await getNotifications(page: 1, limit: 1);
      // The total from the API represents total unread when filtered
      return result.items.where((n) => !n.isRead).length;
    } catch (_) {
      return 0;
    }
  }
}
