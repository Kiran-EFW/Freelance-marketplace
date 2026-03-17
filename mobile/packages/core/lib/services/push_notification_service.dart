import 'dart:io';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import '../api/api_client.dart';

/// Callback type for handling navigation when a notification is tapped.
typedef NotificationNavigationCallback = void Function(
  String type,
  Map<String, dynamic> data,
);

/// Callback type for handling foreground message display.
typedef ForegroundMessageCallback = void Function(RemoteMessage message);

/// Handles FCM push notification setup, token registration, and
/// foreground/background message routing.
///
/// Shared across both customer and provider apps. Each app configures
/// the [onNavigate] and [onForegroundMessage] callbacks to define
/// app-specific behaviour.
class PushNotificationService {
  final FirebaseMessaging _messaging = FirebaseMessaging.instance;
  final ApiClient _apiClient;

  /// Called when a notification is tapped and the app should navigate
  /// to a specific screen.
  NotificationNavigationCallback? onNavigate;

  /// Called when a message arrives while the app is in the foreground.
  /// The app can use this to show a local notification banner.
  ForegroundMessageCallback? onForegroundMessage;

  PushNotificationService({required ApiClient apiClient})
      : _apiClient = apiClient;

  /// Initialize push notifications.
  ///
  /// Requests permission, registers the device token with the backend,
  /// and sets up message listeners. Call once after Firebase is initialized.
  Future<void> initialize() async {
    // Request permission (iOS will show the system prompt; Android auto-grants).
    final settings = await _messaging.requestPermission(
      alert: true,
      badge: true,
      sound: true,
      provisional: false,
    );

    if (settings.authorizationStatus == AuthorizationStatus.denied) {
      debugPrint('PushNotificationService: permission denied by user');
      return;
    }

    // Retrieve the current FCM token and register it.
    final token = await _messaging.getToken();
    if (token != null) {
      await _registerDeviceToken(token);
    }

    // Listen for token refresh events (e.g. app restore, token rotation).
    _messaging.onTokenRefresh.listen(_registerDeviceToken);

    // Handle messages received while the app is in the foreground.
    FirebaseMessaging.onMessage.listen(_handleForegroundMessage);

    // Handle notification tap when the app was in the background.
    FirebaseMessaging.onMessageOpenedApp.listen(_handleMessageOpenedApp);

    // Handle the case where the app was launched from a terminated state
    // by tapping a notification.
    final initialMessage = await _messaging.getInitialMessage();
    if (initialMessage != null) {
      _handleMessageOpenedApp(initialMessage);
    }
  }

  /// Register the device token with the Seva backend so the server
  /// can target this device for push notifications.
  Future<void> _registerDeviceToken(String token) async {
    try {
      final platform = Platform.isIOS ? 'ios' : 'android';
      await _apiClient.registerPushToken(token, platform);
      debugPrint('PushNotificationService: token registered ($platform)');
    } catch (e) {
      debugPrint('PushNotificationService: failed to register token: $e');
    }
  }

  /// Handle a message that arrived while the app is in the foreground.
  ///
  /// Delegates to [onForegroundMessage] so the app can show a local
  /// notification banner or update badge counts.
  void _handleForegroundMessage(RemoteMessage message) {
    debugPrint(
      'PushNotificationService: foreground message ${message.messageId}',
    );
    onForegroundMessage?.call(message);
  }

  /// Handle a notification tap that opened the app from the background
  /// or terminated state.
  ///
  /// Extracts the notification type and associated data, then delegates
  /// to [onNavigate] so the app can route to the correct screen.
  void _handleMessageOpenedApp(RemoteMessage message) {
    debugPrint(
      'PushNotificationService: opened from notification ${message.messageId}',
    );

    final data = message.data;
    final type = data['type'] as String? ?? 'unknown';

    onNavigate?.call(type, data);
  }

  /// Subscribe to a topic for broadcast notifications (e.g. 'promotions').
  Future<void> subscribeToTopic(String topic) async {
    await _messaging.subscribeToTopic(topic);
  }

  /// Unsubscribe from a topic.
  Future<void> unsubscribeFromTopic(String topic) async {
    await _messaging.unsubscribeFromTopic(topic);
  }
}
