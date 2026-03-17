import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';

import '../../main.dart';

/// Screen listing all conversations for the customer.
class MessagesScreen extends ConsumerStatefulWidget {
  const MessagesScreen({super.key});

  @override
  ConsumerState<MessagesScreen> createState() => _MessagesScreenState();
}

class _MessagesScreenState extends ConsumerState<MessagesScreen> {
  List<Conversation> _conversations = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadConversations();
  }

  Future<void> _loadConversations() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final messageRepo = ref.read(messageRepositoryProvider);
      final result = await messageRepo.getConversations();

      if (!mounted) return;

      switch (result) {
        case Success(:final data):
          setState(() {
            _conversations = data.items;
            _isLoading = false;
          });
        case Failure(:final message):
          setState(() {
            _error = message;
            _isLoading = false;
          });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = 'Failed to load messages. Pull to retry.';
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Messages'),
        centerTitle: false,
        actions: [
          IconButton(
            onPressed: _loadConversations,
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _loadConversations,
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline, size: 64, color: SevaColors.neutral300),
              const SizedBox(height: 16),
              Text(
                _error!,
                textAlign: TextAlign.center,
                style: Theme.of(context)
                    .textTheme
                    .bodyLarge
                    ?.copyWith(color: SevaColors.textTertiary),
              ),
              const SizedBox(height: 16),
              SevaButton(
                label: 'Retry',
                isFullWidth: false,
                size: SevaButtonSize.small,
                onPressed: _loadConversations,
              ),
            ],
          ),
        ),
      );
    }

    if (_conversations.isEmpty) {
      return ListView(
        children: [
          const SizedBox(height: 120),
          Center(
            child: Column(
              children: [
                Icon(
                  Icons.chat_bubble_outline,
                  size: 80,
                  color: SevaColors.neutral300,
                ),
                const SizedBox(height: 16),
                Text(
                  'No conversations yet',
                  style: Theme.of(context)
                      .textTheme
                      .headlineSmall
                      ?.copyWith(color: SevaColors.textSecondary),
                ),
                const SizedBox(height: 8),
                Text(
                  'Start a conversation by booking a service\nor contacting a provider.',
                  textAlign: TextAlign.center,
                  style: Theme.of(context)
                      .textTheme
                      .bodyMedium
                      ?.copyWith(color: SevaColors.textTertiary),
                ),
                const SizedBox(height: 24),
                SevaButton(
                  label: 'Find a Provider',
                  isFullWidth: false,
                  size: SevaButtonSize.small,
                  icon: Icons.search,
                  onPressed: () => context.go('/search'),
                ),
              ],
            ),
          ),
        ],
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemCount: _conversations.length,
      separatorBuilder: (_, __) => const Divider(height: 1, indent: 76),
      itemBuilder: (context, index) {
        final conversation = _conversations[index];
        return _ConversationTile(
          conversation: conversation,
          onTap: () {
            context.push(
              '/conversation/${conversation.id}',
              extra: conversation,
            );
          },
        );
      },
    );
  }
}

class _ConversationTile extends StatelessWidget {
  final Conversation conversation;
  final VoidCallback onTap;

  const _ConversationTile({
    required this.conversation,
    required this.onTap,
  });

  String _formatTime(DateTime dateTime) {
    final now = DateTime.now();
    final diff = now.difference(dateTime);

    if (diff.inDays == 0) {
      return DateFormat.jm().format(dateTime);
    }
    if (diff.inDays == 1) {
      return 'Yesterday';
    }
    if (diff.inDays < 7) {
      return DateFormat.E().format(dateTime);
    }
    return DateFormat.MMMd().format(dateTime);
  }

  @override
  Widget build(BuildContext context) {
    final hasUnread = conversation.unreadCount > 0;

    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 4),
      onTap: onTap,
      leading: CircleAvatar(
        radius: 26,
        backgroundColor: SevaColors.primaryFaded,
        backgroundImage: conversation.providerAvatarUrl != null
            ? CachedNetworkImageProvider(conversation.providerAvatarUrl!)
            : null,
        child: conversation.providerAvatarUrl == null
            ? Text(
                conversation.providerName[0].toUpperCase(),
                style: const TextStyle(
                  color: SevaColors.primary,
                  fontWeight: FontWeight.w700,
                  fontSize: 18,
                ),
              )
            : null,
      ),
      title: Row(
        children: [
          Expanded(
            child: Text(
              conversation.providerName,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: hasUnread ? FontWeight.w700 : FontWeight.w500,
                  ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Text(
            _formatTime(conversation.lastMessageAt),
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color:
                      hasUnread ? SevaColors.primary : SevaColors.textTertiary,
                  fontWeight: hasUnread ? FontWeight.w600 : FontWeight.w400,
                ),
          ),
        ],
      ),
      subtitle: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (conversation.jobTitle != null)
                  Text(
                    conversation.jobTitle!,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: SevaColors.primary,
                          fontWeight: FontWeight.w500,
                        ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                Text(
                  conversation.lastMessage,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: hasUnread
                            ? SevaColors.textPrimary
                            : SevaColors.textTertiary,
                        fontWeight:
                            hasUnread ? FontWeight.w500 : FontWeight.w400,
                      ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          if (hasUnread)
            Container(
              margin: const EdgeInsets.only(left: 8),
              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
              decoration: BoxDecoration(
                color: SevaColors.primary,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                conversation.unreadCount > 99
                    ? '99+'
                    : '${conversation.unreadCount}',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
        ],
      ),
    );
  }
}
