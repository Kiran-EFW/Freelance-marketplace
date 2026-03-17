import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class ProviderProfileScreen extends ConsumerWidget {
  const ProviderProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authService = ref.watch(authServiceProvider);
    final user = authService.currentUser;

    if (user == null) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('My Profile'),
        actions: [
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.settings_outlined),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            // Avatar
            Stack(
              children: [
                CircleAvatar(
                  radius: 50,
                  backgroundColor: SevaColors.primaryFaded,
                  backgroundImage: user.avatarUrl != null
                      ? CachedNetworkImageProvider(user.avatarUrl!)
                      : null,
                  child: user.avatarUrl == null
                      ? Text(
                          user.name[0].toUpperCase(),
                          style: const TextStyle(
                            fontSize: 36,
                            fontWeight: FontWeight.w700,
                            color: SevaColors.primary,
                          ),
                        )
                      : null,
                ),
                Positioned(
                  bottom: 0,
                  right: 0,
                  child: Container(
                    padding: const EdgeInsets.all(4),
                    decoration: const BoxDecoration(
                      color: SevaColors.primary,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.camera_alt,
                      color: Colors.white,
                      size: 16,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              user.name,
              style: Theme.of(context)
                  .textTheme
                  .headlineMedium
                  ?.copyWith(fontWeight: FontWeight.w700),
            ),
            Text(
              user.phone,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: SevaColors.textSecondary),
            ),
            const SizedBox(height: 20),

            // KYC status card
            SevaCard(
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: _kycColor(user.kycStatus).withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      Icons.verified_user,
                      color: _kycColor(user.kycStatus),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('KYC Verification',
                            style: Theme.of(context).textTheme.titleSmall),
                        Text(
                          _kycLabel(user.kycStatus),
                          style: Theme.of(context)
                              .textTheme
                              .bodySmall
                              ?.copyWith(color: SevaColors.textTertiary),
                        ),
                      ],
                    ),
                  ),
                  StatusBadge(
                    status: user.kycStatus.toJson(),
                    isCompact: true,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),

            // Menu items
            _MenuItem(
              icon: Icons.handyman_outlined,
              title: 'My Skills',
              subtitle: 'Manage your service categories',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.schedule_outlined,
              title: 'Availability',
              subtitle: 'Set your working hours',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.location_on_outlined,
              title: 'Service Area',
              subtitle: 'Define your service radius',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.account_balance_outlined,
              title: 'Bank Details',
              subtitle: 'Payout account settings',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.description_outlined,
              title: 'Documents',
              subtitle: 'Upload ID and certifications',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.star_outline,
              title: 'Reviews',
              subtitle: 'See what customers say',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.language_outlined,
              title: 'Language',
              subtitle: user.preferredLanguage ?? 'English',
              onTap: () {},
            ),
            _MenuItem(
              icon: Icons.help_outline,
              title: 'Help & Support',
              subtitle: 'FAQs and contact support',
              onTap: () {},
            ),
            const SizedBox(height: 20),

            // Sign out
            SevaButton(
              label: 'Sign Out',
              variant: SevaButtonVariant.outline,
              icon: Icons.logout,
              onPressed: () async {
                await authService.signOut();
                if (context.mounted) {
                  context.go('/auth/login');
                }
              },
            ),
            const SizedBox(height: 40),
          ],
        ),
      ),
    );
  }

  Color _kycColor(KycStatus status) {
    switch (status) {
      case KycStatus.verified:
        return SevaColors.success;
      case KycStatus.pending:
        return SevaColors.warning;
      case KycStatus.rejected:
        return SevaColors.error;
      case KycStatus.notStarted:
        return SevaColors.neutral400;
    }
  }

  String _kycLabel(KycStatus status) {
    switch (status) {
      case KycStatus.verified:
        return 'Your identity is verified';
      case KycStatus.pending:
        return 'Verification in progress';
      case KycStatus.rejected:
        return 'Verification rejected - resubmit';
      case KycStatus.notStarted:
        return 'Complete verification to get more jobs';
    }
  }
}

class _MenuItem extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  const _MenuItem({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: SevaColors.primaryFaded,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: SevaColors.primary, size: 20),
      ),
      title: Text(title, style: Theme.of(context).textTheme.titleSmall),
      subtitle: Text(
        subtitle,
        style: Theme.of(context)
            .textTheme
            .bodySmall
            ?.copyWith(color: SevaColors.textTertiary),
      ),
      trailing:
          const Icon(Icons.chevron_right, color: SevaColors.neutral400),
      onTap: onTap,
      contentPadding: EdgeInsets.zero,
    );
  }
}
