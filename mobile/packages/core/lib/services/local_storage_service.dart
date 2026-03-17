import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as p;
import 'package:sqflite/sqflite.dart';

/// An operation that was queued while offline and needs to be synced
/// when connectivity is restored.
class PendingOperation {
  final String id;
  final String method; // POST, PUT, PATCH, DELETE
  final String endpoint;
  final Map<String, dynamic> body;
  final DateTime createdAt;

  const PendingOperation({
    required this.id,
    required this.method,
    required this.endpoint,
    required this.body,
    required this.createdAt,
  });

  Map<String, dynamic> toMap() => {
        'id': id,
        'method': method,
        'endpoint': endpoint,
        'body': jsonEncode(body),
        'created_at': createdAt.toIso8601String(),
      };

  factory PendingOperation.fromMap(Map<String, dynamic> map) {
    return PendingOperation(
      id: map['id'] as String,
      method: map['method'] as String,
      endpoint: map['endpoint'] as String,
      body: jsonDecode(map['body'] as String) as Map<String, dynamic>,
      createdAt: DateTime.parse(map['created_at'] as String),
    );
  }
}

/// Local SQLite-backed storage service for offline-first caching and
/// operation queuing.
///
/// Provides two main capabilities:
/// 1. **API response caching** -- cache JSON data by key for fast
///    offline reads and reduced network calls.
/// 2. **Operation queue** -- persist write operations (POST/PUT/DELETE)
///    that were attempted while offline so they can be replayed when
///    connectivity returns.
class LocalStorageService {
  Database? _db;

  static const String _cacheTable = 'cache';
  static const String _pendingOpsTable = 'pending_operations';

  /// Open (or create) the local database. Call once at app startup.
  Future<void> initialize() async {
    final dbPath = await getDatabasesPath();
    final path = p.join(dbPath, 'seva_local.db');

    _db = await openDatabase(
      path,
      version: 1,
      onCreate: (db, version) async {
        await db.execute('''
          CREATE TABLE $_cacheTable (
            key TEXT PRIMARY KEY,
            data TEXT NOT NULL,
            cached_at TEXT NOT NULL
          )
        ''');
        await db.execute('''
          CREATE TABLE $_pendingOpsTable (
            id TEXT PRIMARY KEY,
            method TEXT NOT NULL,
            endpoint TEXT NOT NULL,
            body TEXT NOT NULL,
            created_at TEXT NOT NULL
          )
        ''');
      },
    );
  }

  Database get _database {
    if (_db == null) {
      throw StateError(
        'LocalStorageService.initialize() must be called before use.',
      );
    }
    return _db!;
  }

  // ---------------------------------------------------------------------------
  // Cache API responses
  // ---------------------------------------------------------------------------

  /// Cache a JSON map under [key], replacing any previous value.
  Future<void> cacheData(String key, Map<String, dynamic> data) async {
    await _database.insert(
      _cacheTable,
      {
        'key': key,
        'data': jsonEncode(data),
        'cached_at': DateTime.now().toIso8601String(),
      },
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  /// Retrieve cached data for [key], or null if not cached.
  Future<Map<String, dynamic>?> getCachedData(String key) async {
    final results = await _database.query(
      _cacheTable,
      where: 'key = ?',
      whereArgs: [key],
      limit: 1,
    );

    if (results.isEmpty) return null;
    return jsonDecode(results.first['data'] as String) as Map<String, dynamic>;
  }

  /// Retrieve cached data only if it was cached within [maxAge].
  Future<Map<String, dynamic>?> getCachedDataIfFresh(
    String key, {
    Duration maxAge = const Duration(hours: 1),
  }) async {
    final results = await _database.query(
      _cacheTable,
      where: 'key = ?',
      whereArgs: [key],
      limit: 1,
    );

    if (results.isEmpty) return null;

    final cachedAt = DateTime.parse(results.first['cached_at'] as String);
    if (DateTime.now().difference(cachedAt) > maxAge) return null;

    return jsonDecode(results.first['data'] as String) as Map<String, dynamic>;
  }

  /// Remove a single cache entry.
  Future<void> removeCachedData(String key) async {
    await _database.delete(
      _cacheTable,
      where: 'key = ?',
      whereArgs: [key],
    );
  }

  /// Clear all cached data.
  Future<void> clearCache() async {
    await _database.delete(_cacheTable);
  }

  // ---------------------------------------------------------------------------
  // Operation queue (offline writes)
  // ---------------------------------------------------------------------------

  /// Enqueue a write operation to be replayed when online.
  Future<void> queueOperation(PendingOperation op) async {
    await _database.insert(
      _pendingOpsTable,
      op.toMap(),
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
    debugPrint('LocalStorage: queued ${op.method} ${op.endpoint}');
  }

  /// Get all pending operations ordered by creation time (FIFO).
  Future<List<PendingOperation>> getPendingOperations() async {
    final results = await _database.query(
      _pendingOpsTable,
      orderBy: 'created_at ASC',
    );
    return results.map((row) => PendingOperation.fromMap(row)).toList();
  }

  /// Remove a successfully synced operation from the queue.
  Future<void> markOperationSynced(String operationId) async {
    await _database.delete(
      _pendingOpsTable,
      where: 'id = ?',
      whereArgs: [operationId],
    );
  }

  /// Get the number of pending operations.
  Future<int> getPendingOperationCount() async {
    final result = await _database.rawQuery(
      'SELECT COUNT(*) as count FROM $_pendingOpsTable',
    );
    return Sqflite.firstIntValue(result) ?? 0;
  }

  /// Clear all pending operations.
  Future<void> clearPendingOperations() async {
    await _database.delete(_pendingOpsTable);
  }

  /// Close the database. Call on app teardown if needed.
  Future<void> close() async {
    await _db?.close();
    _db = null;
  }
}
