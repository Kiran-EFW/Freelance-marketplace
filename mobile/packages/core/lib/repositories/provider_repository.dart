import 'package:dio/dio.dart';

import '../api/api_client.dart';
import '../models/provider.dart';
import '../models/result.dart';
import '../models/review.dart';
import '../models/category.dart';
import 'job_repository.dart';

/// Handles provider search, profile, and earnings API interactions.
class ProviderRepository {
  final ApiClient _api;

  ProviderRepository({required ApiClient api}) : _api = api;

  /// Search for providers with optional filters.
  Future<Result<PaginatedResult<ServiceProvider>>> searchProviders({
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

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to search providers'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to search providers: $e');
    }
  }

  /// Fetch a single provider's full profile.
  Future<Result<ServiceProvider>> getProvider(String providerId) async {
    try {
      final response = await _api.getProvider(providerId);
      return Success(
        ServiceProvider.fromJson(response.data as Map<String, dynamic>),
      );
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load provider'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load provider: $e');
    }
  }

  /// Fetch reviews for a provider.
  Future<Result<PaginatedResult<Review>>> getProviderReviews(
    String providerId, {
    int page = 1,
  }) async {
    try {
      final response = await _api.getProviderReviews(providerId, page: page);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Review.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load reviews'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load reviews: $e');
    }
  }

  /// Update the current provider's profile.
  Future<Result<ServiceProvider>> updateProfile(
    String providerId,
    Map<String, dynamic> updates,
  ) async {
    try {
      final response = await _api.updateProviderProfile(providerId, updates);
      return Success(
        ServiceProvider.fromJson(response.data as Map<String, dynamic>),
      );
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to update profile'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to update profile: $e');
    }
  }

  /// Update availability schedule.
  Future<Result<bool>> updateAvailability(
    String providerId,
    List<AvailabilitySlot> slots,
  ) async {
    try {
      await _api.updateAvailability(providerId, {
        'slots': slots.map((s) => s.toJson()).toList(),
      });
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to update availability'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to update availability: $e');
    }
  }

  /// Fetch all service categories.
  Future<Result<List<Category>>> getCategories({String? parentId}) async {
    try {
      final response = await _api.getCategories(parentId: parentId);
      final categories = (response.data as List<dynamic>)
          .map((e) => Category.fromJson(e as Map<String, dynamic>))
          .toList();
      return Success(categories);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load categories'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load categories: $e');
    }
  }

  /// Fetch earnings summary.
  Future<Result<EarningsSummary>> getEarnings({
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
      return Success(
        EarningsSummary.fromJson(response.data as Map<String, dynamic>),
      );
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load earnings'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load earnings: $e');
    }
  }

  /// Fetch payout history.
  Future<Result<PaginatedResult<Payout>>> getPayoutHistory({
    int page = 1,
  }) async {
    try {
      final response = await _api.getPayoutHistory(page: page);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Payout.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load payout history'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load payout history: $e');
    }
  }

  /// Request a payout.
  Future<Result<Payout>> requestPayout(double amount) async {
    try {
      final response = await _api.requestPayout(amount: amount);
      return Success(Payout.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to request payout'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to request payout: $e');
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
