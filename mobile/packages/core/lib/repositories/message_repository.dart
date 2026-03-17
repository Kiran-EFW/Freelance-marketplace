import 'package:dio/dio.dart';

import '../api/api_client.dart';
import '../models/message.dart';
import '../models/result.dart';
import 'job_repository.dart';

/// Handles all messaging-related API interactions.
class MessageRepository {
  final ApiClient _api;

  MessageRepository({required ApiClient api}) : _api = api;

  /// Fetch paginated conversations for the current user.
  Future<Result<PaginatedResult<Conversation>>> getConversations({
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final response = await _api.getConversations(page: page, limit: limit);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Conversation.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load conversations'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load conversations: $e');
    }
  }

  /// Fetch paginated messages for a single conversation.
  Future<Result<PaginatedResult<Message>>> getMessages(
    String conversationId, {
    int page = 1,
  }) async {
    try {
      final response = await _api.getMessages(conversationId, page: page);
      final data = response.data as Map<String, dynamic>;
      final items = (data['items'] as List<dynamic>)
          .map((e) => Message.fromJson(e as Map<String, dynamic>))
          .toList();

      return Success(PaginatedResult(
        items: items,
        total: data['total'] as int,
        page: data['page'] as int,
        totalPages: data['total_pages'] as int,
      ));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to load messages'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to load messages: $e');
    }
  }

  /// Send a message in a conversation and return the created message.
  Future<Result<Message>> sendMessage(
    String conversationId,
    String content,
  ) async {
    try {
      final response = await _api.sendMessage(conversationId, content);
      return Success(Message.fromJson(response.data as Map<String, dynamic>));
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to send message'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to send message: $e');
    }
  }

  /// Create a new conversation with a provider, optionally linked to a job.
  Future<Result<Conversation>> createConversation(
    String providerId, {
    String? jobId,
  }) async {
    try {
      final response = await _api.createConversation(providerId, jobId: jobId);
      return Success(
        Conversation.fromJson(response.data as Map<String, dynamic>),
      );
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to create conversation'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to create conversation: $e');
    }
  }

  /// Mark all messages in a conversation as read.
  Future<Result<bool>> markAsRead(String conversationId) async {
    try {
      await _api.markMessagesRead(conversationId);
      return const Success(true);
    } on DioException catch (e) {
      return Failure(
        _extractErrorMessage(e, 'Failed to mark messages as read'),
        statusCode: e.response?.statusCode,
      );
    } catch (e) {
      return Failure('Failed to mark messages as read: $e');
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
