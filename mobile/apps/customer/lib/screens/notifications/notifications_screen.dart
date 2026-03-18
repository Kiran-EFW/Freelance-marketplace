import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class NotificationsScreen extends ConsumerStatefulWidget {
  const NotificationsScreen({super.key});

  @override
  ConsumerState<NotificationsScreen> createState() =>
      _NotificationsScreenState();
}

class _NotificationsScreenState extends ConsumerState<NotificationsScreen> {
  List<AppNotification> _notifications = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadNotifications();
  }

  Future<void> _loadNotifications() async {
    setState(() => _isLoading = true);

    final notifRepo = ref.read(notificationRepositoryProvider);
    final result = await notifRepo.getNotifications();

    if (mounted) {
      setState(() {
        final paginated = result.dataOrNull;
        _notifications = paginated?.items ?? [];
        _isLoading = false;
      });
    }
  }

  Future<void> _markAllRead() async {
    await ref.read(notificationRepositoryProvider).markAllAsRead();
    if (mounted) {
      setState(() {
        _notifications = _notifications
            .map((n) => n.copyWith(isRead: true))
            .toList();
      });
    }
  }

  void _onNotificationTap(AppNotification notification) async {
    // Mark as read
    if (!notification.isRead) {
      await ref
          .read(notificationRepositoryProvider)
          .markAsRead(notification.id);
      if (mounted) {
        setState(() {
          final idx = _notifications.indexWhere((n) => n.id == notification.id);
          if (idx >= 0) {
            _notifications[idx] = notification.copyWith(isRead: true);
          }
        });
      }
    }

    // Navigate to relevant screen
    if (!mounted) return;

    final data = notification.data;
    if (data != null) {
      final jobId = data['job_id'] as String?;
      if (jobId != null) {
        context.push('/job/$jobId');
        return;
      }
      final providerId = data['provider_id'] as String?;
      if (providerId != null) {
        context.push('/provider/$providerId');
        return;
      }
    }
  }

  IconData _notificationIcon(NotificationType type) {
    switch (type) {
      case NotificationType.jobPosted:
        return Icons.work_outline;
      case NotificationType.jobAccepted:
        return Icons.check_circle_outline;
      case NotificationType.jobDeclined:
        return Icons.cancel_outlined;
      case NotificationType.jobStarted:
        return Icons.play_circle_outline;
      case NotificationType.jobCompleted:
        return Icons.task_alt;
      case NotificationType.jobCancelled:
        return Icons.cancel;
      case NotificationType.newReview:
        return Icons.star_outline;
      case NotificationType.paymentReceived:
        return Icons.payment;
      case NotificationType.payoutProcessed:
        return Icons.account_balance;
      case NotificationType.kycUpdate:
        return Icons.verified_user_outlined;
      case NotificationType.promotional:
        return Icons.local_offer_outlined;
      case NotificationType.system:
        return Icons.info_outline;
    }
  }

  Color _notificationColor(NotificationType type) {
    switch (type) {
      case NotificationType.jobCompleted:
      case NotificationType.paymentReceived:
      case NotificationType.payoutProcessed:
        return SevaColors.success;
      case NotificationType.jobCancelled:
      case NotificationType.jobDeclined:
        return SevaColors.error;
      case NotificationType.jobAccepted:
      case NotificationType.jobStarted:
        return SevaColors.secondary;
      case NotificationType.newReview:
        return SevaColors.starFilled;
      case NotificationType.promotional:
        return SevaColors.primary;
      default:
        return SevaColors.info;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
        actions: [
          if (_notifications.any((n) => !n.isRead))
            TextButton(
              onPressed: _markAllRead,
              child: const Text('Mark All Read'),
            ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _notifications.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.notifications_none,
                        size: 64,
                        color: SevaColors.neutral300,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'No notifications yet',
                        style:
                            Theme.of(context).textTheme.bodyLarge?.copyWith(
                                  color: SevaColors.textTertiary,
                                ),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadNotifications,
                  child: ListView.separated(
                    itemCount: _notifications.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final notification = _notifications[index];
                      return ListTile(
                        onTap: () => _onNotificationTap(notification),
                        tileColor:
                            notification.isRead ? null : SevaColors.primaryFaded,
                        leading: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: _notificationColor(notification.type)
                                .withValues(alpha: 0.1),
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            _notificationIcon(notification.type),
                            color: _notificationColor(notification.type),
                            size: 20,
                          ),
                        ),
                        title: Text(
                          notification.title,
                          style:
                              Theme.of(context).textTheme.titleSmall?.copyWith(
                                    fontWeight: notification.isRead
                                        ? FontWeight.w400
                                        : FontWeight.w600,
                                  ),
                        ),
                        subtitle: Text(
                          notification.body,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                        trailing: Text(
                          notification.timeAgo,
                          style:
                              Theme.of(context).textTheme.labelSmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}
