import 'package:equatable/equatable.dart';

/// Types of push/in-app notifications the Seva platform sends.
enum NotificationType {
  jobPosted,
  jobAccepted,
  jobDeclined,
  jobStarted,
  jobCompleted,
  jobCancelled,
  newReview,
  paymentReceived,
  payoutProcessed,
  kycUpdate,
  promotional,
  system;

  factory NotificationType.fromString(String value) {
    switch (value.toLowerCase()) {
      case 'job_posted':
        return NotificationType.jobPosted;
      case 'job_accepted':
        return NotificationType.jobAccepted;
      case 'job_declined':
        return NotificationType.jobDeclined;
      case 'job_started':
        return NotificationType.jobStarted;
      case 'job_completed':
        return NotificationType.jobCompleted;
      case 'job_cancelled':
        return NotificationType.jobCancelled;
      case 'new_review':
        return NotificationType.newReview;
      case 'payment_received':
        return NotificationType.paymentReceived;
      case 'payout_processed':
        return NotificationType.payoutProcessed;
      case 'kyc_update':
        return NotificationType.kycUpdate;
      case 'promotional':
        return NotificationType.promotional;
      case 'system':
        return NotificationType.system;
      default:
        return NotificationType.system;
    }
  }

  String toJson() {
    switch (this) {
      case NotificationType.jobPosted:
        return 'job_posted';
      case NotificationType.jobAccepted:
        return 'job_accepted';
      case NotificationType.jobDeclined:
        return 'job_declined';
      case NotificationType.jobStarted:
        return 'job_started';
      case NotificationType.jobCompleted:
        return 'job_completed';
      case NotificationType.jobCancelled:
        return 'job_cancelled';
      case NotificationType.newReview:
        return 'new_review';
      case NotificationType.paymentReceived:
        return 'payment_received';
      case NotificationType.payoutProcessed:
        return 'payout_processed';
      case NotificationType.kycUpdate:
        return 'kyc_update';
      case NotificationType.promotional:
        return 'promotional';
      case NotificationType.system:
        return 'system';
    }
  }
}

/// An in-app or push notification for a Seva user.
class AppNotification extends Equatable {
  final String id;
  final String userId;
  final NotificationType type;
  final String title;
  final String body;
  final Map<String, dynamic>? data;
  final String? actionUrl;
  final String? imageUrl;
  final bool isRead;
  final DateTime createdAt;

  const AppNotification({
    required this.id,
    required this.userId,
    required this.type,
    required this.title,
    required this.body,
    this.data,
    this.actionUrl,
    this.imageUrl,
    this.isRead = false,
    required this.createdAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) {
    return AppNotification(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      type: NotificationType.fromString(json['type'] as String),
      title: json['title'] as String,
      body: json['body'] as String,
      data: json['data'] as Map<String, dynamic>?,
      actionUrl: json['action_url'] as String?,
      imageUrl: json['image_url'] as String?,
      isRead: json['is_read'] as bool? ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'type': type.toJson(),
      'title': title,
      'body': body,
      'data': data,
      'action_url': actionUrl,
      'image_url': imageUrl,
      'is_read': isRead,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Time elapsed since the notification was created.
  String get timeAgo {
    final diff = DateTime.now().difference(createdAt);
    if (diff.inDays > 365) return '${diff.inDays ~/ 365}y ago';
    if (diff.inDays > 30) return '${diff.inDays ~/ 30}mo ago';
    if (diff.inDays > 0) return '${diff.inDays}d ago';
    if (diff.inHours > 0) return '${diff.inHours}h ago';
    if (diff.inMinutes > 0) return '${diff.inMinutes}m ago';
    return 'Just now';
  }

  AppNotification copyWith({
    String? id,
    String? userId,
    NotificationType? type,
    String? title,
    String? body,
    Map<String, dynamic>? data,
    String? actionUrl,
    String? imageUrl,
    bool? isRead,
    DateTime? createdAt,
  }) {
    return AppNotification(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      type: type ?? this.type,
      title: title ?? this.title,
      body: body ?? this.body,
      data: data ?? this.data,
      actionUrl: actionUrl ?? this.actionUrl,
      imageUrl: imageUrl ?? this.imageUrl,
      isRead: isRead ?? this.isRead,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  @override
  List<Object?> get props => [id, userId, type, isRead, createdAt];
}
