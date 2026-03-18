import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

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
        title: const Text('Profile'),
        actions: [
          IconButton(
            onPressed: () {
              // Navigate to settings
            },
            icon: const Icon(Icons.settings_outlined),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            // Avatar and name
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
            const SizedBox(height: 12),
            Text(
              user.name,
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
            ),
            const SizedBox(height: 4),
            Text(
              user.phone,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: SevaColors.textSecondary,
                  ),
            ),
            const SizedBox(height: 20),

            // Loyalty points card
            SevaCard(
              backgroundColor: SevaColors.primary,
              child: Row(
                children: [
                  const Icon(Icons.stars, color: Colors.white, size: 32),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Loyalty Points',
                          style: TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                          ),
                        ),
                        Text(
                          '${user.loyaltyPoints}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 28,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.2),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: const Text(
                      'Redeem',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w600,
                        fontSize: 12,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),

            // Menu items
            _ProfileMenuItem(
              icon: Icons.work_outline,
              title: 'My Jobs',
              subtitle: 'View your job history',
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.location_on_outlined,
              title: 'Saved Addresses',
              subtitle: 'Manage your addresses',
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.payment_outlined,
              title: 'Payment Methods',
              subtitle: 'Manage payment options',
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.verified_user_outlined,
              title: 'KYC Verification',
              subtitle: user.kycStatus.toJson().replaceAll('_', ' '),
              trailing: StatusBadge(
                status: user.kycStatus.toJson(),
                isCompact: true,
              ),
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.language_outlined,
              title: 'Language',
              subtitle: user.preferredLanguage ?? 'English',
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.help_outline,
              title: 'Help & Support',
              subtitle: 'FAQs and contact support',
              onTap: () {},
            ),
            _ProfileMenuItem(
              icon: Icons.info_outline,
              title: 'About Seva',
              subtitle: 'Version 1.0.0',
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
}

class _ProfileMenuItem extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final Widget? trailing;
  final VoidCallback onTap;

  const _ProfileMenuItem({
    required this.icon,
    required this.title,
    required this.subtitle,
    this.trailing,
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
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: SevaColors.textTertiary,
            ),
      ),
      trailing:
          trailing ?? const Icon(Icons.chevron_right, color: SevaColors.neutral400),
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 0, vertical: 2),
    );
  }
}
