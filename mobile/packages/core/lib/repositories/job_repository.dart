import 'package:dio/dio.dart';

import '../api/api_client.dart';
import '../models/job.dart';
import '../models/result.dart';
import '../models/review.dart';

/// Paginated response container.
class PaginatedResult<T> {
  final List<T> items;
  final int total;
  final int page;
  final int totalPages;

  const PaginatedResult({
    required this.items,
    required this.total,
    required this.page,
    required this.totalPages,
  });

  bool get hasMore => page < totalPages;

  Map<String, dynamic> toJson(
    Map<String, dynamic> Function(T) itemToJson,
  ) {
    return {
      'items': items.map(itemToJson).toList(),
      'total': total,
      'page': page,
      'total_pages': totalPages,
    };
  }
}

/// Handles all job-related API interactions.
class JobRepository {
  final ApiClient _api;

  JobRepository({required ApiClient api}) : _api = api;

  /// Create a new job posting.
  Future<Result<Job>> createJob({
    required String categoryId,
    required String title,
    required String description,
    double? budgetMin,
    double? budgetMax,
    String? postcode,
    double? latitude,
    double? longitude,
    String? address,
    DateTime? scheduledAt,
    String? urgency,
    List<String>? photoUrls,
  }) async {
    try {
      final response = await _api.createJob({
        'category_id': categoryId,
        'title': title,
        'description': description,
        if (budgetMin != null) 'budget_min': budgetMin,
        if (budgetMax != null) 'budget_max': budgetMax,
        if (postcode != null) 'postcode': postcode,
        if (latitude != null) 'latitude': latitude,
        if (longitude != null) 'longitude': longitude,
        if (address != null) 'address': address,
        if (scheduledAt != null) 'scheduled_at': scheduledAt.toIso8601String(),
        if (urgency != null) 'urgency': urgency,
        if (photoUrls != null) 'photo_urls': photoUrls,
      });
      return Success(Job.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to create job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to create job: $e');
    }
  }

  /// Fetch a single job by ID.
  Future<Result<Job>> getJob(String jobId) async {
    try {
      final response = await _api.getJob(jobId);
      return Success(Job.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load job: $e');
    }
  }

  /// List jobs for the current user, filtered by [status] and [role].
  Future<Result<PaginatedResult<Job>>> getJobs({
    String? status,
    String? role,
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final response = await _api.getJobs(
        status: status,
        role: role,
        page: page,
        limit: limit,
      );
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Job.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load jobs'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load jobs: $e');
    }
  }

  /// Accept a job (provider action).
  Future<Result<Job>> acceptJob(String jobId) async {
    try {
      final response = await _api.acceptJob(jobId);
      return Success(Job.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to accept job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to accept job: $e');
    }
  }

  /// Decline a job (provider action).
  Future<Result<bool>> declineJob(String jobId, {String? reason}) async {
    try {
      await _api.declineJob(jobId, reason: reason);
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to decline job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to decline job: $e');
    }
  }

  /// Update job status (e.g., start work, complete).
  Future<Result<Job>> updateStatus(String jobId, String status) async {
    try {
      final response = await _api.updateJobStatus(jobId, status);
      return Success(Job.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to update job status'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to update job status: $e');
    }
  }

  /// Mark job as completed, optionally with completion photos.
  Future<Result<Job>> completeJob(String jobId,
      {List<String>? photoUrls}) async {
    try {
      final response = await _api.completeJob(jobId, photoUrls: photoUrls);
      return Success(Job.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to complete job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to complete job: $e');
    }
  }

  /// Cancel a job.
  Future<Result<bool>> cancelJob(String jobId,
      {required String reason}) async {
    try {
      await _api.cancelJob(jobId, reason: reason);
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to cancel job'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to cancel job: $e');
    }
  }

  /// Submit a review for a completed job.
  Future<Result<Review>> submitReview(
    String jobId, {
    required int rating,
    String? comment,
  }) async {
    try {
      final response = await _api.submitReview(
        jobId,
        rating: rating,
        comment: comment,
      );
      return Success(Review.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to submit review'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to submit review: $e');
    }
  }

  /// Upload photos for a job.
  Future<Result<List<String>>> uploadPhotos(
    String jobId,
    List<String> filePaths,
  ) async {
    try {
      final response = await _api.uploadJobPhotos(jobId, filePaths);
      final urls = (response.data['urls'] as List<dynamic>)
          .map((e) => e as String)
          .toList();
      return Success(urls);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to upload photos'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to upload photos: $e');
    }
  }

  /// Extract a human-readable error message from a [DioException].
  String _extractErrorMessage(DioException e, String fallback) {
    // Try to get the server-provided error message.
    final data = e.response?.data;
    if (data is Map<String, dynamic>) {
      final message = data['message'] ?? data['error'];
      if (message is String && message.isNotEmpty) return message;
    }

    // Fall back to Dio error type descriptions.
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
