import '../api/api_client.dart';
import '../models/provider.dart';
import '../models/review.dart';
import '../models/category.dart';
import 'job_repository.dart';

/// Handles provider search, profile, and earnings API interactions.
class ProviderRepository {
  final ApiClient _api;

  ProviderRepository({required ApiClient api}) : _api = api;

  /// Search for providers with optional filters.
  Future<PaginatedResult<ServiceProvider>> searchProviders({
    String? query,
    String? categoryId,
    double? latitude,
    double? longitude,
    int? radiusKm,
    double? minRating,
    int page = 1,
    int limit = 20,
    String? sortBy,
  }) async {
    try {
      final response = await _api.searchProviders(
        query: query,
        categoryId: categoryId,
        latitude: latitude,
        longitude: longitude,
        radiusKm: radiusKm,
        minRating: minRating,
        page: page,
        limit: limit,
        sortBy: sortBy,
      );
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => ServiceProvider.fromJson(e as Map<String, dynamic>))
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

  /// Fetch a single provider's full profile.
  Future<ServiceProvider?> getProvider(String providerId) async {
    try {
      final response = await _api.getProvider(providerId);
      return ServiceProvider.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Fetch reviews for a provider.
  Future<PaginatedResult<Review>> getProviderReviews(
    String providerId, {
    int page = 1,
  }) async {
    try {
      final response = await _api.getProviderReviews(providerId, page: page);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Review.fromJson(e as Map<String, dynamic>))
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

  /// Update the current provider's profile.
  Future<ServiceProvider?> updateProfile(
    String providerId,
    Map<String, dynamic> updates,
  ) async {
    try {
      final response = await _api.updateProviderProfile(providerId, updates);
      return ServiceProvider.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Update availability schedule.
  Future<bool> updateAvailability(
    String providerId,
    List<AvailabilitySlot> slots,
  ) async {
    try {
      await _api.updateAvailability(providerId, {
        'slots': slots.map((s) => s.toJson()).toList(),
      });
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Fetch all service categories.
  Future<List<Category>> getCategories({String? parentId}) async {
    try {
      final response = await _api.getCategories(parentId: parentId);
      return (response.data as List<dynamic>)
          .map((e) => Category.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }

  /// Fetch earnings summary.
  Future<EarningsSummary?> getEarnings({
    String? period,
    String? from,
    String? to,
  }) async {
    try {
      final response = await _api.getEarnings(
        period: period,
        from: from,
        to: to,
      );
      return EarningsSummary.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Fetch payout history.
  Future<PaginatedResult<Payout>> getPayoutHistory({int page = 1}) async {
    try {
      final response = await _api.getPayoutHistory(page: page);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Payout.fromJson(e as Map<String, dynamic>))
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

  /// Request a payout.
  Future<Payout?> requestPayout(double amount) async {
    try {
      final response = await _api.requestPayout(amount: amount);
      return Payout.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }
}

/// Aggregated earnings data.
class EarningsSummary {
  final double totalEarnings;
  final double pendingPayout;
  final double availableBalance;
  final int jobsCompleted;
  final double averageJobValue;
  final List<EarningsDataPoint> chartData;

  const EarningsSummary({
    required this.totalEarnings,
    required this.pendingPayout,
    required this.availableBalance,
    required this.jobsCompleted,
    required this.averageJobValue,
    required this.chartData,
  });

  factory EarningsSummary.fromJson(Map<String, dynamic> json) {
    return EarningsSummary(
      totalEarnings: (json['total_earnings'] as num).toDouble(),
      pendingPayout: (json['pending_payout'] as num).toDouble(),
      availableBalance: (json['available_balance'] as num).toDouble(),
      jobsCompleted: json['jobs_completed'] as int,
      averageJobValue: (json['average_job_value'] as num).toDouble(),
      chartData: (json['chart_data'] as List<dynamic>?)
              ?.map(
                  (e) => EarningsDataPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }
}

/// A single data point for the earnings chart.
class EarningsDataPoint {
  final String label;
  final double amount;

  const EarningsDataPoint({required this.label, required this.amount});

  factory EarningsDataPoint.fromJson(Map<String, dynamic> json) {
    return EarningsDataPoint(
      label: json['label'] as String,
      amount: (json['amount'] as num).toDouble(),
    );
  }
}

/// A payout transaction record.
class Payout {
  final String id;
  final double amount;
  final String currency;
  final String status;
  final String? bankAccount;
  final DateTime createdAt;
  final DateTime? processedAt;

  const Payout({
    required this.id,
    required this.amount,
    required this.currency,
    required this.status,
    this.bankAccount,
    required this.createdAt,
    this.processedAt,
  });

  factory Payout.fromJson(Map<String, dynamic> json) {
    return Payout(
      id: json['id'] as String,
      amount: (json['amount'] as num).toDouble(),
      currency: json['currency'] as String,
      status: json['status'] as String,
      bankAccount: json['bank_account'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String)
          : null,
    );
  }
}
