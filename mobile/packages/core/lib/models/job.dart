import 'package:equatable/equatable.dart';

/// Status lifecycle for a job on the Seva platform.
enum JobStatus {
  draft,
  posted,
  matched,
  accepted,
  inProgress,
  completed,
  cancelled,
  disputed;

  factory JobStatus.fromString(String value) {
    switch (value.toLowerCase()) {
      case 'draft':
        return JobStatus.draft;
      case 'posted':
        return JobStatus.posted;
      case 'matched':
        return JobStatus.matched;
      case 'accepted':
        return JobStatus.accepted;
      case 'in_progress':
        return JobStatus.inProgress;
      case 'completed':
        return JobStatus.completed;
      case 'cancelled':
        return JobStatus.cancelled;
      case 'disputed':
        return JobStatus.disputed;
      default:
        return JobStatus.draft;
    }
  }

  String toJson() {
    switch (this) {
      case JobStatus.draft:
        return 'draft';
      case JobStatus.posted:
        return 'posted';
      case JobStatus.matched:
        return 'matched';
      case JobStatus.accepted:
        return 'accepted';
      case JobStatus.inProgress:
        return 'in_progress';
      case JobStatus.completed:
        return 'completed';
      case JobStatus.cancelled:
        return 'cancelled';
      case JobStatus.disputed:
        return 'disputed';
    }
  }
}

/// Urgency level for a job.
enum JobUrgency {
  low,
  normal,
  high,
  emergency;

  factory JobUrgency.fromString(String value) {
    return JobUrgency.values.firstWhere(
      (e) => e.name == value.toLowerCase(),
      orElse: () => JobUrgency.normal,
    );
  }
}

/// A service job posted by a customer and fulfilled by a provider.
class Job extends Equatable {
  final String id;
  final String customerId;
  final String? providerId;
  final String categoryId;
  final String title;
  final String description;
  final JobStatus status;
  final JobUrgency urgency;
  final double? budgetMin;
  final double? budgetMax;
  final double? agreedPrice;
  final String? currency;
  final String? postcode;
  final double? latitude;
  final double? longitude;
  final String? address;
  final DateTime? scheduledAt;
  final DateTime? startedAt;
  final DateTime? completedAt;
  final List<String> photoUrls;
  final List<String> completionPhotoUrls;
  final String? customerName;
  final String? providerName;
  final String? categoryName;
  final double? providerRating;
  final DateTime createdAt;
  final DateTime updatedAt;

  const Job({
    required this.id,
    required this.customerId,
    this.providerId,
    required this.categoryId,
    required this.title,
    required this.description,
    required this.status,
    this.urgency = JobUrgency.normal,
    this.budgetMin,
    this.budgetMax,
    this.agreedPrice,
    this.currency,
    this.postcode,
    this.latitude,
    this.longitude,
    this.address,
    this.scheduledAt,
    this.startedAt,
    this.completedAt,
    this.photoUrls = const [],
    this.completionPhotoUrls = const [],
    this.customerName,
    this.providerName,
    this.categoryName,
    this.providerRating,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Job.fromJson(Map<String, dynamic> json) {
    return Job(
      id: json['id'] as String,
      customerId: json['customer_id'] as String,
      providerId: json['provider_id'] as String?,
      categoryId: json['category_id'] as String,
      title: json['title'] as String,
      description: json['description'] as String,
      status: JobStatus.fromString(json['status'] as String),
      urgency: JobUrgency.fromString(json['urgency'] as String? ?? 'normal'),
      budgetMin: (json['budget_min'] as num?)?.toDouble(),
      budgetMax: (json['budget_max'] as num?)?.toDouble(),
      agreedPrice: (json['agreed_price'] as num?)?.toDouble(),
      currency: json['currency'] as String?,
      postcode: json['postcode'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      address: json['address'] as String?,
      scheduledAt: json['scheduled_at'] != null
          ? DateTime.parse(json['scheduled_at'] as String)
          : null,
      startedAt: json['started_at'] != null
          ? DateTime.parse(json['started_at'] as String)
          : null,
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
      photoUrls: (json['photo_urls'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
      completionPhotoUrls: (json['completion_photo_urls'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
      customerName: json['customer_name'] as String?,
      providerName: json['provider_name'] as String?,
      categoryName: json['category_name'] as String?,
      providerRating: (json['provider_rating'] as num?)?.toDouble(),
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'customer_id': customerId,
      'provider_id': providerId,
      'category_id': categoryId,
      'title': title,
      'description': description,
      'status': status.toJson(),
      'urgency': urgency.name,
      'budget_min': budgetMin,
      'budget_max': budgetMax,
      'agreed_price': agreedPrice,
      'currency': currency,
      'postcode': postcode,
      'latitude': latitude,
      'longitude': longitude,
      'address': address,
      'scheduled_at': scheduledAt?.toIso8601String(),
      'started_at': startedAt?.toIso8601String(),
      'completed_at': completedAt?.toIso8601String(),
      'photo_urls': photoUrls,
      'completion_photo_urls': completionPhotoUrls,
      'customer_name': customerName,
      'provider_name': providerName,
      'category_name': categoryName,
      'provider_rating': providerRating,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Whether the job is in a terminal state.
  bool get isTerminal =>
      status == JobStatus.completed ||
      status == JobStatus.cancelled ||
      status == JobStatus.disputed;

  /// Whether the job can be cancelled.
  bool get isCancellable =>
      status == JobStatus.posted ||
      status == JobStatus.matched ||
      status == JobStatus.accepted;

  /// Formatted budget range string.
  String get budgetDisplay {
    if (agreedPrice != null) {
      return '${currency ?? "INR"} ${agreedPrice!.toStringAsFixed(0)}';
    }
    if (budgetMin != null && budgetMax != null) {
      return '${currency ?? "INR"} ${budgetMin!.toStringAsFixed(0)} - ${budgetMax!.toStringAsFixed(0)}';
    }
    if (budgetMin != null) {
      return 'From ${currency ?? "INR"} ${budgetMin!.toStringAsFixed(0)}';
    }
    return 'Price TBD';
  }

  Job copyWith({
    String? id,
    String? customerId,
    String? providerId,
    String? categoryId,
    String? title,
    String? description,
    JobStatus? status,
    JobUrgency? urgency,
    double? budgetMin,
    double? budgetMax,
    double? agreedPrice,
    String? currency,
    String? postcode,
    double? latitude,
    double? longitude,
    String? address,
    DateTime? scheduledAt,
    DateTime? startedAt,
    DateTime? completedAt,
    List<String>? photoUrls,
    List<String>? completionPhotoUrls,
    String? customerName,
    String? providerName,
    String? categoryName,
    double? providerRating,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return Job(
      id: id ?? this.id,
      customerId: customerId ?? this.customerId,
      providerId: providerId ?? this.providerId,
      categoryId: categoryId ?? this.categoryId,
      title: title ?? this.title,
      description: description ?? this.description,
      status: status ?? this.status,
      urgency: urgency ?? this.urgency,
      budgetMin: budgetMin ?? this.budgetMin,
      budgetMax: budgetMax ?? this.budgetMax,
      agreedPrice: agreedPrice ?? this.agreedPrice,
      currency: currency ?? this.currency,
      postcode: postcode ?? this.postcode,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      address: address ?? this.address,
      scheduledAt: scheduledAt ?? this.scheduledAt,
      startedAt: startedAt ?? this.startedAt,
      completedAt: completedAt ?? this.completedAt,
      photoUrls: photoUrls ?? this.photoUrls,
      completionPhotoUrls: completionPhotoUrls ?? this.completionPhotoUrls,
      customerName: customerName ?? this.customerName,
      providerName: providerName ?? this.providerName,
      categoryName: categoryName ?? this.categoryName,
      providerRating: providerRating ?? this.providerRating,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  List<Object?> get props => [
        id,
        customerId,
        providerId,
        categoryId,
        title,
        description,
        status,
        urgency,
        budgetMin,
        budgetMax,
        agreedPrice,
        scheduledAt,
        createdAt,
      ];
}
