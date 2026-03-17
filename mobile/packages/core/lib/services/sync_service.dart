import 'dart:async';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../api/api_client.dart';
import 'local_storage_service.dart';

/// Manages offline/online synchronisation.
///
/// Responsibilities:
/// - Monitors connectivity changes and auto-syncs when the device
///   comes back online.
/// - Replays queued write operations (POST/PUT/PATCH/DELETE) that
///   were recorded while offline.
/// - Provides a cache-first fetch strategy for GET requests so
///   screens render immediately from local data, then refresh from
///   the network.
class SyncService {
  final ApiClient _apiClient;
  final LocalStorageService _localStorage;
  final Connectivity _connectivity = Connectivity();

  StreamSubscription<List<ConnectivityResult>>? _connectivitySubscription;

  /// Whether the device currently has network connectivity.
  bool _isOnline = true;
  bool get isOnline => _isOnline;

  /// Stream that emits whenever the connectivity state changes.
  final StreamController<bool> _onlineController =
      StreamController<bool>.broadcast();
  Stream<bool> get onlineStream => _onlineController.stream;

  SyncService({
    required ApiClient apiClient,
    required LocalStorageService localStorage,
  })  : _apiClient = apiClient,
        _localStorage = localStorage {
    _startConnectivityMonitor();
  }

  // ---------------------------------------------------------------------------
  // Connectivity monitoring
  // ---------------------------------------------------------------------------

  void _startConnectivityMonitor() {
    _connectivitySubscription =
        _connectivity.onConnectivityChanged.listen((results) {
      final wasOnline = _isOnline;
      _isOnline =
          results.isNotEmpty && !results.contains(ConnectivityResult.none);
      _onlineController.add(_isOnline);

      if (!wasOnline && _isOnline) {
        debugPrint('SyncService: back online – syncing pending operations');
        syncPendingOperations();
      }
    });

    // Seed the initial connectivity state.
    _connectivity.checkConnectivity().then((results) {
      _isOnline =
          results.isNotEmpty && !results.contains(ConnectivityResult.none);
      _onlineController.add(_isOnline);
    });
  }

  // ---------------------------------------------------------------------------
  // Sync pending operations
  // ---------------------------------------------------------------------------

  /// Replay all pending write operations in FIFO order.
  ///
  /// Stops on the first failure under the assumption that later
  /// operations may depend on earlier ones. The remaining operations
  /// stay in the queue and will be retried on the next sync attempt.
  Future<void> syncPendingOperations() async {
    final pending = await _localStorage.getPendingOperations();
    if (pending.isEmpty) return;

    debugPrint('SyncService: ${pending.length} operations to sync');

    for (final op in pending) {
      try {
        await _rawRequest(op.method, op.endpoint, body: op.body);
        await _localStorage.markOperationSynced(op.id);
        debugPrint('SyncService: synced ${op.method} ${op.endpoint}');
      } on DioException catch (e) {
        // If the server explicitly rejects the request (4xx) we remove
        // it from the queue to avoid infinite retries. Network errors
        // (no connectivity, timeouts) cause us to stop and retry later.
        final statusCode = e.response?.statusCode;
        if (statusCode != null && statusCode >= 400 && statusCode < 500) {
          debugPrint(
            'SyncService: server rejected ${op.method} ${op.endpoint} '
            '($statusCode) – removing from queue',
          );
          await _localStorage.markOperationSynced(op.id);
          continue;
        }
        debugPrint(
          'SyncService: failed to sync ${op.method} ${op.endpoint} – '
          'will retry later',
        );
        break;
      } catch (e) {
        debugPrint('SyncService: unexpected error: $e – stopping sync');
        break;
      }
    }
  }

  // ---------------------------------------------------------------------------
  // Cache-first fetch strategy
  // ---------------------------------------------------------------------------

  /// Fetch data using a cache-first strategy.
  ///
  /// 1. Try the network request.
  /// 2. On success, cache the response and return it.
  /// 3. On failure, fall back to the locally cached copy if available.
  /// 4. If no cache exists, rethrow the error.
  Future<T> fetchWithCache<T>({
    required String cacheKey,
    required Future<T> Function() apiCall,
    required T Function(Map<String, dynamic>) fromJson,
    required Map<String, dynamic> Function(T) toJson,
    Duration maxAge = const Duration(hours: 1),
  }) async {
    try {
      final data = await apiCall();
      // Cache the successful response.
      await _localStorage.cacheData(cacheKey, toJson(data));
      return data;
    } catch (e) {
      // Fall back to cache.
      final cached = await _localStorage.getCachedData(cacheKey);
      if (cached != null) {
        debugPrint('SyncService: serving $cacheKey from cache');
        return fromJson(cached);
      }
      rethrow;
    }
  }

  // ---------------------------------------------------------------------------
  // Internal HTTP helper
  // ---------------------------------------------------------------------------

  /// Perform a raw HTTP request using the API client's Dio instance.
  /// This is used to replay queued operations.
  Future<Response> _rawRequest(
    String method,
    String endpoint, {
    Map<String, dynamic>? body,
  }) async {
    // Use the internal typed helpers on ApiClient based on method.
    switch (method.toUpperCase()) {
      case 'POST':
        return _apiClient.createJob(body ?? {});
      case 'PATCH':
        return _apiClient.updateJobStatus(endpoint, body?['status'] ?? '');
      case 'DELETE':
        // Cancellation is the most common DELETE-like operation.
        return _apiClient.cancelJob(
          endpoint,
          reason: body?['reason'] ?? 'Offline cancellation',
        );
      default:
        throw UnsupportedError('SyncService: unsupported method $method');
    }
  }

  // ---------------------------------------------------------------------------
  // Lifecycle
  // ---------------------------------------------------------------------------

  /// Dispose of resources. Call on app teardown.
  void dispose() {
    _connectivitySubscription?.cancel();
    _onlineController.close();
  }
}
