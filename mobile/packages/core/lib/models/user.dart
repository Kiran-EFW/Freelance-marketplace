import 'package:equatable/equatable.dart';

/// Roles a user can hold within the Seva platform.
enum UserRole {
  customer,
  provider,
  admin;

  factory UserRole.fromString(String value) {
    return UserRole.values.firstWhere(
      (e) => e.name == value.toLowerCase(),
      orElse: () => UserRole.customer,
    );
  }
}

/// KYC verification status for a user.
enum KycStatus {
  notStarted,
  pending,
  verified,
  rejected;

  factory KycStatus.fromString(String value) {
    switch (value.toLowerCase()) {
      case 'not_started':
        return KycStatus.notStarted;
      case 'pending':
        return KycStatus.pending;
      case 'verified':
        return KycStatus.verified;
      case 'rejected':
        return KycStatus.rejected;
      default:
        return KycStatus.notStarted;
    }
  }

  String toJson() {
    switch (this) {
      case KycStatus.notStarted:
        return 'not_started';
      case KycStatus.pending:
        return 'pending';
      case KycStatus.verified:
        return 'verified';
      case KycStatus.rejected:
        return 'rejected';
    }
  }
}

/// Core user model shared across the platform.
class User extends Equatable {
  final String id;
  final String name;
  final String phone;
  final String? email;
  final UserRole role;
  final String? avatarUrl;
  final String? postcode;
  final double? latitude;
  final double? longitude;
  final KycStatus kycStatus;
  final int loyaltyPoints;
  final String? preferredLanguage;
  final DateTime createdAt;
  final DateTime updatedAt;

  const User({
    required this.id,
    required this.name,
    required this.phone,
    this.email,
    required this.role,
    this.avatarUrl,
    this.postcode,
    this.latitude,
    this.longitude,
    this.kycStatus = KycStatus.notStarted,
    this.loyaltyPoints = 0,
    this.preferredLanguage,
    required this.createdAt,
    required this.updatedAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      name: json['name'] as String,
      phone: json['phone'] as String,
      email: json['email'] as String?,
      role: UserRole.fromString(json['role'] as String),
      avatarUrl: json['avatar_url'] as String?,
      postcode: json['postcode'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      kycStatus: KycStatus.fromString(json['kyc_status'] as String? ?? 'not_started'),
      loyaltyPoints: json['loyalty_points'] as int? ?? 0,
      preferredLanguage: json['preferred_language'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'phone': phone,
      'email': email,
      'role': role.name,
      'avatar_url': avatarUrl,
      'postcode': postcode,
      'latitude': latitude,
      'longitude': longitude,
      'kyc_status': kycStatus.toJson(),
      'loyalty_points': loyaltyPoints,
      'preferred_language': preferredLanguage,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  User copyWith({
    String? id,
    String? name,
    String? phone,
    String? email,
    UserRole? role,
    String? avatarUrl,
    String? postcode,
    double? latitude,
    double? longitude,
    KycStatus? kycStatus,
    int? loyaltyPoints,
    String? preferredLanguage,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return User(
      id: id ?? this.id,
      name: name ?? this.name,
      phone: phone ?? this.phone,
      email: email ?? this.email,
      role: role ?? this.role,
      avatarUrl: avatarUrl ?? this.avatarUrl,
      postcode: postcode ?? this.postcode,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      kycStatus: kycStatus ?? this.kycStatus,
      loyaltyPoints: loyaltyPoints ?? this.loyaltyPoints,
      preferredLanguage: preferredLanguage ?? this.preferredLanguage,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  List<Object?> get props => [
        id,
        name,
        phone,
        email,
        role,
        avatarUrl,
        postcode,
        latitude,
        longitude,
        kycStatus,
        loyaltyPoints,
        preferredLanguage,
        createdAt,
        updatedAt,
      ];
}
