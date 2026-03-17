import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:seva_ui_kit/ui_kit.dart';

import 'messages_screen.dart';

/// A single chat message in a conversation.
class ChatMessage {
  final String id;
  final String senderId;
  final String content;
  final DateTime sentAt;
  final bool isFromCurrentUser;
  final ChatMessageStatus status;

  const ChatMessage({
    required this.id,
    required this.senderId,
    required this.content,
    required this.sentAt,
    required this.isFromCurrentUser,
    this.status = ChatMessageStatus.sent,
  });

  factory ChatMessage.fromJson(
    Map<String, dynamic> json, {
    required String currentUserId,
  }) {
    return ChatMessage(
      id: json['id'] as String,
      senderId: json['sender_id'] as String,
      content: json['content'] as String,
      sentAt: DateTime.parse(json['sent_at'] as String),
      isFromCurrentUser: json['sender_id'] == currentUserId,
      status: ChatMessageStatus.fromString(json['status'] as String? ?? 'sent'),
    );
  }
}

enum ChatMessageStatus {
  sending,
  sent,
  delivered,
  read;

  factory ChatMessageStatus.fromString(String value) {
    return ChatMessageStatus.values.firstWhere(
      (e) => e.name == value.toLowerCase(),
      orElse: () => ChatMessageStatus.sent,
    );
  }
}

/// Full-screen chat view for a single conversation.
class ConversationScreen extends ConsumerStatefulWidget {
  final String conversationId;
  final Conversation? conversation;

  const ConversationScreen({
    super.key,
    required this.conversationId,
    this.conversation,
  });

  @override
  ConsumerState<ConversationScreen> createState() => _ConversationScreenState();
}

class _ConversationScreenState extends ConsumerState<ConversationScreen> {
  final _messageController = TextEditingController();
  final _scrollController = ScrollController();
  final _focusNode = FocusNode();

  List<ChatMessage> _messages = [];
  bool _isLoading = true;
  bool _isSending = false;
  bool _hasMore = true;
  int _currentPage = 1;

  @override
  void initState() {
    super.initState();
    _loadMessages();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _messageController.dispose();
    _scrollController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _onScroll() {
    // Pull to load older messages when scrolled to top.
    if (_scrollController.position.pixels <=
            _scrollController.position.minScrollExtent + 50 &&
        _hasMore &&
        !_isLoading) {
      _loadOlderMessages();
    }
  }

  Future<void> _loadMessages() async {
    setState(() => _isLoading = true);

    try {
      // TODO: Replace with actual API call.
      // final response = await apiClient.getMessages(widget.conversationId);
      await Future.delayed(const Duration(milliseconds: 300));

      if (mounted) {
        setState(() {
          _messages = [];
          _isLoading = false;
        });
        _scrollToBottom();
      }
    } catch (_) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _loadOlderMessages() async {
    if (!_hasMore || _isLoading) return;

    _currentPage++;

    try {
      // TODO: Replace with actual API call.
      // final response = await apiClient.getMessages(
      //   widget.conversationId,
      //   page: _currentPage,
      // );
      await Future.delayed(const Duration(milliseconds: 300));

      if (mounted) {
        setState(() {
          // Prepend older messages.
          _hasMore = false; // Until API is wired up.
        });
      }
    } catch (_) {
      _currentPage--;
    }
  }

  Future<void> _sendMessage() async {
    final text = _messageController.text.trim();
    if (text.isEmpty || _isSending) return;

    setState(() => _isSending = true);
    _messageController.clear();

    // Optimistic insert.
    final tempMessage = ChatMessage(
      id: 'temp_${DateTime.now().millisecondsSinceEpoch}',
      senderId: 'current_user',
      content: text,
      sentAt: DateTime.now(),
      isFromCurrentUser: true,
      status: ChatMessageStatus.sending,
    );

    setState(() {
      _messages.add(tempMessage);
    });
    _scrollToBottom();

    try {
      // TODO: Replace with actual API call.
      // final response = await apiClient.sendMessage(
      //   widget.conversationId,
      //   content: text,
      // );
      await Future.delayed(const Duration(milliseconds: 500));

      if (mounted) {
        setState(() {
          // Replace temp message with server response.
          final index = _messages.indexWhere((m) => m.id == tempMessage.id);
          if (index != -1) {
            _messages[index] = ChatMessage(
              id: tempMessage.id,
              senderId: tempMessage.senderId,
              content: tempMessage.content,
              sentAt: tempMessage.sentAt,
              isFromCurrentUser: true,
              status: ChatMessageStatus.sent,
            );
          }
          _isSending = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => _isSending = false);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to send message')),
        );
      }
    }
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final providerName =
        widget.conversation?.providerName ?? 'Conversation';

    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            CircleAvatar(
              radius: 18,
              backgroundColor: SevaColors.primaryFaded,
              backgroundImage:
                  widget.conversation?.providerAvatarUrl != null
                      ? CachedNetworkImageProvider(
                          widget.conversation!.providerAvatarUrl!,
                        )
                      : null,
              child: widget.conversation?.providerAvatarUrl == null
                  ? Text(
                      providerName[0].toUpperCase(),
                      style: const TextStyle(
                        color: SevaColors.primary,
                        fontWeight: FontWeight.w700,
                        fontSize: 14,
                      ),
                    )
                  : null,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    providerName,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  if (widget.conversation?.jobTitle != null)
                    Text(
                      widget.conversation!.jobTitle!,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: SevaColors.textTertiary,
                          ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
      body: Column(
        children: [
          // Messages list
          Expanded(child: _buildMessageList()),

          // Input area
          _buildInputBar(),
        ],
      ),
    );
  }

  Widget _buildMessageList() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_messages.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.chat_bubble_outline,
              size: 64,
              color: SevaColors.neutral300,
            ),
            const SizedBox(height: 12),
            Text(
              'No messages yet',
              style: Theme.of(context)
                  .textTheme
                  .bodyLarge
                  ?.copyWith(color: SevaColors.textTertiary),
            ),
            const SizedBox(height: 4),
            Text(
              'Send a message to start the conversation.',
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: SevaColors.textTertiary),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      itemCount: _messages.length,
      itemBuilder: (context, index) {
        final message = _messages[index];
        final showDate = index == 0 ||
            !_isSameDay(
              _messages[index - 1].sentAt,
              message.sentAt,
            );

        return Column(
          children: [
            if (showDate) _DateSeparator(date: message.sentAt),
            _ChatBubble(message: message),
          ],
        );
      },
    );
  }

  bool _isSameDay(DateTime a, DateTime b) {
    return a.year == b.year && a.month == b.month && a.day == b.day;
  }

  Widget _buildInputBar() {
    return Container(
      padding: EdgeInsets.only(
        left: 12,
        right: 8,
        top: 8,
        bottom: MediaQuery.of(context).padding.bottom + 8,
      ),
      decoration: BoxDecoration(
        color: Theme.of(context).scaffoldBackgroundColor,
        border: Border(
          top: BorderSide(color: SevaColors.neutral200),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: TextField(
              controller: _messageController,
              focusNode: _focusNode,
              maxLines: 4,
              minLines: 1,
              textInputAction: TextInputAction.newline,
              textCapitalization: TextCapitalization.sentences,
              decoration: InputDecoration(
                hintText: 'Type a message...',
                hintStyle: TextStyle(color: SevaColors.textTertiary),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: SevaColors.neutral200),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: BorderSide(color: SevaColors.neutral200),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: const BorderSide(color: SevaColors.primary),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                isDense: true,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Material(
            color: SevaColors.primary,
            shape: const CircleBorder(),
            child: InkWell(
              customBorder: const CircleBorder(),
              onTap: _sendMessage,
              child: const Padding(
                padding: EdgeInsets.all(10),
                child: Icon(
                  Icons.send,
                  color: Colors.white,
                  size: 22,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _DateSeparator extends StatelessWidget {
  final DateTime date;

  const _DateSeparator({required this.date});

  String _formatDate() {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inDays == 0) return 'Today';
    if (diff.inDays == 1) return 'Yesterday';
    if (diff.inDays < 7) return DateFormat.EEEE().format(date);
    return DateFormat.yMMMd().format(date);
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 16),
      child: Center(
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          decoration: BoxDecoration(
            color: SevaColors.neutral200,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            _formatDate(),
            style: const TextStyle(
              fontSize: 12,
              color: SevaColors.textTertiary,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
      ),
    );
  }
}

class _ChatBubble extends StatelessWidget {
  final ChatMessage message;

  const _ChatBubble({required this.message});

  @override
  Widget build(BuildContext context) {
    final isMe = message.isFromCurrentUser;

    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        mainAxisAlignment:
            isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (!isMe) const SizedBox(width: 4),
          Flexible(
            child: Container(
              constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.75,
              ),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: isMe ? SevaColors.primary : SevaColors.neutral100,
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(16),
                  topRight: const Radius.circular(16),
                  bottomLeft: Radius.circular(isMe ? 16 : 4),
                  bottomRight: Radius.circular(isMe ? 4 : 16),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    message.content,
                    style: TextStyle(
                      color: isMe ? Colors.white : SevaColors.textPrimary,
                      fontSize: 15,
                      height: 1.4,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        DateFormat.jm().format(message.sentAt),
                        style: TextStyle(
                          fontSize: 10,
                          color: isMe
                              ? Colors.white.withValues(alpha: 0.7)
                              : SevaColors.textTertiary,
                        ),
                      ),
                      if (isMe) ...[
                        const SizedBox(width: 3),
                        Icon(
                          message.status == ChatMessageStatus.sending
                              ? Icons.access_time
                              : message.status == ChatMessageStatus.read
                                  ? Icons.done_all
                                  : Icons.done,
                          size: 14,
                          color: Colors.white.withValues(alpha: 0.7),
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ),
          if (isMe) const SizedBox(width: 4),
        ],
      ),
    );
  }
}
