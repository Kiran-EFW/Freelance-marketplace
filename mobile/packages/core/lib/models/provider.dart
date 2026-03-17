import 'package:equatable/equatable.dart';
import 'category.dart';

/// Availability slot for a service provider.
class AvailabilitySlot extends Equatable {
  final int dayOfWeek; // 0 = Sunday, 6 = Saturday
  final String startTime; // "09:00"
  final String endTime; // "17:00"

  const AvailabilitySlot({
    required this.dayOfWeek,
    required this.startTime,
    required this.endTime,
  });

  factory AvailabilitySlot.fromJson(Map<String, dynamic> json) {
    return AvailabilitySlot(
      dayOfWeek: json['day_of_week'] as int,
      startTime: json['start_time'] as String,
      endTime: json['end_time'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
        'day_of_week': dayOfWeek,
        'start_time': startTime,
        'end_time': endTime,
      };

  @override
  List<Object?> get props => [dayOfWeek, startTime, endTime];
}

/// A service provider on the Seva platform.
class ServiceProvider extends Equatable {
  final String id;
  final String userId;
  final String name;
  final String phone;
  final String? email;
  final String? avatarUrl;
  final String? bio;
  final double rating;
  final int reviewCount;
  final int completedJobs;
  final double trustScore;
  final List<String> skills;
  final List<Category> categories;
  final String? postcode;
  final double? latitude;
  final double? longitude;
  final double? distanceKm;
  final bool isAvailable;
  final bool isVerified;
  final List<AvailabilitySlot> availability;
  final String? serviceRadius;
  final double? hourlyRate;
  final String? currency;
  final int responseTimeMinutes;
  final DateTime createdAt;
  final DateTime updatedAt;

  const ServiceProvider({
    required this.id,
    required this.userId,
    required this.name,
    required this.phone,
    this.email,
    this.avatarUrl,
    this.bio,
    this.rating = 0.0,
    this.reviewCount = 0,
    this.completedJobs = 0,
    this.trustScore = 0.0,
    this.skills = const [],
    this.categories = const [],
    this.postcode,
    this.latitude,
    this.longitude,
    this.distanceKm,
    this.isAvailable = true,
    this.isVerified = false,
    this.availability = const [],
    this.serviceRadius,
    this.hourlyRate,
    this.currency,
    this.responseTimeMinutes = 0,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ServiceProvider.fromJson(Map<String, dynamic> json) {
    return ServiceProvider(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      phone: json['phone'] as String,
      email: json['email'] as String?,
      avatarUrl: json['avatar_url'] as String?,
      bio: json['bio'] as String?,
      rating: (json['rating'] as num?)?.toDouble() ?? 0.0,
      reviewCount: json['review_count'] as int? ?? 0,
      completedJobs: json['completed_jobs'] as int? ?? 0,
      trustScore: (json['trust_score'] as num?)?.toDouble() ?? 0.0,
      skills: (json['skills'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
      categories: (json['categories'] as List<dynamic>?)
              ?.map((e) => Category.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      postcode: json['postcode'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      distanceKm: (json['distance_km'] as num?)?.toDouble(),
      isAvailable: json['is_available'] as bool? ?? true,
      isVerified: json['is_verified'] as bool? ?? false,
      availability: (json['availability'] as List<dynamic>?)
              ?.map(
                  (e) => AvailabilitySlot.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      serviceRadius: json['service_radius'] as String?,
      hourlyRate: (json['hourly_rate'] as num?)?.toDouble(),
      currency: json['currency'] as String?,
      responseTimeMinutes: json['response_time_minutes'] as int? ?? 0,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'name': name,
      'phone': phone,
      'email': email,
      'avatar_url': avatarUrl,
      'bio': bio,
      'rating': rating,
      'review_count': reviewCount,
      'completed_jobs': completedJobs,
      'trust_score': trustScore,
      'skills': skills,
      'categories': categories.map((c) => c.toJson()).toList(),
      'postcode': postcode,
      'latitude': latitude,
      'longitude': longitude,
      'distance_km': distanceKm,
      'is_available': isAvailable,
      'is_verified': isVerified,
      'availability': availability.map((a) => a.toJson()).toList(),
      'service_radius': serviceRadius,
      'hourly_rate': hourlyRate,
      'currency': currency,
      'response_time_minutes': responseTimeMinutes,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Human-readable trust score label.
  String get trustLabel {
    if (trustScore >= 90) return 'Excellent';
    if (trustScore >= 75) return 'Very Good';
    if (trustScore >= 60) return 'Good';
    if (trustScore >= 40) return 'Fair';
    return 'New';
  }

  /// Human-readable distance string.
  String get distanceDisplay {
    if (distanceKm == null) return '';
    if (distanceKm! < 1) {
      return '${(distanceKm! * 1000).round()} m away';
    }
    return '${distanceKm!.toStringAsFixed(1)} km away';
  }

  ServiceProvider copyWith({
    String? id,
    String? userId,
    String? name,
    String? phone,
    String? email,
    String? avatarUrl,
    String? bio,
    double? rating,
    int? reviewCount,
    int? completedJobs,
    double? trustScore,
    List<String>? skills,
    List<Category>? categories,
    String? postcode,
    double? latitude,
    double? longitude,
    double? distanceKm,
    bool? isAvailable,
    bool? isVerified,
    List<AvailabilitySlot>? availability,
    String? serviceRadius,
    double? hourlyRate,
    String? currency,
    int? responseTimeMinutes,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ServiceProvider(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      name: name ?? this.name,
      phone: phone ?? this.phone,
      email: email ?? this.email,
      avatarUrl: avatarUrl ?? this.avatarUrl,
      bio: bio ?? this.bio,
      rating: rating ?? this.rating,
      reviewCount: reviewCount ?? this.reviewCount,
      completedJobs: completedJobs ?? this.completedJobs,
      trustScore: trustScore ?? this.trustScore,
      skills: skills ?? this.skills,
      categories: categories ?? this.categories,
      postcode: postcode ?? this.postcode,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      distanceKm: distanceKm ?? this.distanceKm,
      isAvailable: isAvailable ?? this.isAvailable,
      isVerified: isVerified ?? this.isVerified,
      availability: availability ?? this.availability,
      serviceRadius: serviceRadius ?? this.serviceRadius,
      hourlyRate: hourlyRate ?? this.hourlyRate,
      currency: currency ?? this.currency,
      responseTimeMinutes: responseTimeMinutes ?? this.responseTimeMinutes,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  List<Object?> get props => [id, userId, name, rating, trustScore, isVerified];
}
