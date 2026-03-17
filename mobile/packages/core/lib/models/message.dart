import 'package:equatable/equatable.dart';

/// Status of a chat message.
enum MessageStatus {
  sending,
  sent,
  delivered,
  read;

  factory MessageStatus.fromString(String value) {
    return MessageStatus.values.firstWhere(
      (e) => e.name == value.toLowerCase(),
      orElse: () => MessageStatus.sent,
    );
  }
}

/// A single message within a conversation.
class Message extends Equatable {
  final String id;
  final String conversationId;
  final String senderId;
  final String content;
  final MessageStatus status;
  final DateTime createdAt;

  const Message({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.content,
    this.status = MessageStatus.sent,
    required this.createdAt,
  });

  factory Message.fromJson(Map<String, dynamic> json) {
    return Message(
      id: json['id'] as String,
      conversationId: json['conversation_id'] as String,
      senderId: json['sender_id'] as String,
      content: json['content'] as String,
      status: MessageStatus.fromString(json['status'] as String? ?? 'sent'),
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'conversation_id': conversationId,
      'sender_id': senderId,
      'content': content,
      'status': status.name,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Whether this message was sent by the given [userId].
  bool isSentBy(String userId) => senderId == userId;

  Message copyWith({
    String? id,
    String? conversationId,
    String? senderId,
    String? content,
    MessageStatus? status,
    DateTime? createdAt,
  }) {
    return Message(
      id: id ?? this.id,
      conversationId: conversationId ?? this.conversationId,
      senderId: senderId ?? this.senderId,
      content: content ?? this.content,
      status: status ?? this.status,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  @override
  List<Object?> get props => [id, conversationId, senderId, content, status, createdAt];
}

/// A conversation summary used in the conversation list.
///
/// This is the shared model; the screen-level [Conversation] class in
/// `messages_screen.dart` is replaced by this canonical model.
class Conversation extends Equatable {
  final String id;
  final String providerId;
  final String providerName;
  final String? providerAvatarUrl;
  final String lastMessage;
  final DateTime lastMessageAt;
  final int unreadCount;
  final String? jobId;
  final String? jobTitle;

  const Conversation({
    required this.id,
    required this.providerId,
    required this.providerName,
    this.providerAvatarUrl,
    required this.lastMessage,
    required this.lastMessageAt,
    this.unreadCount = 0,
    this.jobId,
    this.jobTitle,
  });

  factory Conversation.fromJson(Map<String, dynamic> json) {
    return Conversation(
      id: json['id'] as String,
      providerId: json['provider_id'] as String,
      providerName: json['provider_name'] as String,
      providerAvatarUrl: json['provider_avatar_url'] as String?,
      lastMessage: json['last_message'] as String,
      lastMessageAt: DateTime.parse(json['last_message_at'] as String),
      unreadCount: json['unread_count'] as int? ?? 0,
      jobId: json['job_id'] as String?,
      jobTitle: json['job_title'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'provider_id': providerId,
      'provider_name': providerName,
      'provider_avatar_url': providerAvatarUrl,
      'last_message': lastMessage,
      'last_message_at': lastMessageAt.toIso8601String(),
      'unread_count': unreadCount,
      'job_id': jobId,
      'job_title': jobTitle,
    };
  }

  @override
  List<Object?> get props => [
        id,
        providerId,
        providerName,
        lastMessage,
        lastMessageAt,
        unreadCount,
        jobId,
      ];
}
