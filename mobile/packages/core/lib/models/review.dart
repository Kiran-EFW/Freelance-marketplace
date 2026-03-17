import 'package:equatable/equatable.dart';

/// A review left by a customer for a completed job.
class Review extends Equatable {
  final String id;
  final String jobId;
  final String reviewerId;
  final String revieweeId;
  final int rating;
  final String? comment;
  final String? reviewerName;
  final String? reviewerAvatarUrl;
  final String? jobTitle;
  final String? categoryName;
  final List<String> photoUrls;
  final bool isFlagged;
  final String? flagReason;
  final DateTime createdAt;

  const Review({
    required this.id,
    required this.jobId,
    required this.reviewerId,
    required this.revieweeId,
    required this.rating,
    this.comment,
    this.reviewerName,
    this.reviewerAvatarUrl,
    this.jobTitle,
    this.categoryName,
    this.photoUrls = const [],
    this.isFlagged = false,
    this.flagReason,
    required this.createdAt,
  });

  factory Review.fromJson(Map<String, dynamic> json) {
    return Review(
      id: json['id'] as String,
      jobId: json['job_id'] as String,
      reviewerId: json['reviewer_id'] as String,
      revieweeId: json['reviewee_id'] as String,
      rating: json['rating'] as int,
      comment: json['comment'] as String?,
      reviewerName: json['reviewer_name'] as String?,
      reviewerAvatarUrl: json['reviewer_avatar_url'] as String?,
      jobTitle: json['job_title'] as String?,
      categoryName: json['category_name'] as String?,
      photoUrls: (json['photo_urls'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
      isFlagged: json['is_flagged'] as bool? ?? false,
      flagReason: json['flag_reason'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'job_id': jobId,
      'reviewer_id': reviewerId,
      'reviewee_id': revieweeId,
      'rating': rating,
      'comment': comment,
      'reviewer_name': reviewerName,
      'reviewer_avatar_url': reviewerAvatarUrl,
      'job_title': jobTitle,
      'category_name': categoryName,
      'photo_urls': photoUrls,
      'is_flagged': isFlagged,
      'flag_reason': flagReason,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Time elapsed since the review was posted.
  String get timeAgo {
    final diff = DateTime.now().difference(createdAt);
    if (diff.inDays > 365) return '${diff.inDays ~/ 365}y ago';
    if (diff.inDays > 30) return '${diff.inDays ~/ 30}mo ago';
    if (diff.inDays > 0) return '${diff.inDays}d ago';
    if (diff.inHours > 0) return '${diff.inHours}h ago';
    if (diff.inMinutes > 0) return '${diff.inMinutes}m ago';
    return 'Just now';
  }

  @override
  List<Object?> get props => [id, jobId, reviewerId, revieweeId, rating];
}
