import '../api/api_client.dart';
import '../models/job.dart';
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
}

/// Handles all job-related API interactions.
class JobRepository {
  final ApiClient _api;

  JobRepository({required ApiClient api}) : _api = api;

  /// Create a new job posting.
  Future<Job?> createJob({
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
      return Job.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Fetch a single job by ID.
  Future<Job?> getJob(String jobId) async {
    try {
      final response = await _api.getJob(jobId);
      return Job.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// List jobs for the current user, filtered by [status] and [role].
  Future<PaginatedResult<Job>> getJobs({
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

  /// Accept a job (provider action).
  Future<Job?> acceptJob(String jobId) async {
    try {
      final response = await _api.acceptJob(jobId);
      return Job.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Decline a job (provider action).
  Future<bool> declineJob(String jobId, {String? reason}) async {
    try {
      await _api.declineJob(jobId, reason: reason);
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Update job status (e.g., start work, complete).
  Future<Job?> updateStatus(String jobId, String status) async {
    try {
      final response = await _api.updateJobStatus(jobId, status);
      return Job.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Mark job as completed, optionally with completion photos.
  Future<Job?> completeJob(String jobId, {List<String>? photoUrls}) async {
    try {
      final response = await _api.completeJob(jobId, photoUrls: photoUrls);
      return Job.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Cancel a job.
  Future<bool> cancelJob(String jobId, {required String reason}) async {
    try {
      await _api.cancelJob(jobId, reason: reason);
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Submit a review for a completed job.
  Future<Review?> submitReview(
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
      return Review.fromJson(response.data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  /// Upload photos for a job.
  Future<List<String>> uploadPhotos(
    String jobId,
    List<String> filePaths,
  ) async {
    try {
      final response = await _api.uploadJobPhotos(jobId, filePaths);
      return (response.data['urls'] as List<dynamic>)
          .map((e) => e as String)
          .toList();
    } catch (_) {
      return [];
    }
  }
}
