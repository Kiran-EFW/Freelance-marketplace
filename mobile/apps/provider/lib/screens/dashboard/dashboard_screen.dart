import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class DashboardScreen extends ConsumerStatefulWidget {
  const DashboardScreen({super.key});

  @override
  ConsumerState<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends ConsumerState<DashboardScreen> {
  EarningsSummary? _earnings;
  List<Job> _activeJobs = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    final providerRepo = ref.read(providerRepositoryProvider);
    final jobRepo = ref.read(jobRepositoryProvider);

    final results = await Future.wait([
      providerRepo.getEarnings(period: 'month'),
      jobRepo.getJobs(status: 'in_progress', role: 'provider', limit: 5),
    ]);

    if (mounted) {
      setState(() {
        _earnings = results[0] as EarningsSummary?;
        _activeJobs = (results[1] as PaginatedResult<Job>).items;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final authService = ref.watch(authServiceProvider);
    final user = authService.currentUser;

    return Scaffold(
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: _loadData,
          child: CustomScrollView(
            slivers: [
              // Header
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Dashboard',
                              style: Theme.of(context)
                                  .textTheme
                                  .headlineMedium
                                  ?.copyWith(fontWeight: FontWeight.w700),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Welcome back, ${user?.name ?? "Provider"}',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodyMedium
                                  ?.copyWith(color: SevaColors.textSecondary),
                            ),
                          ],
                        ),
                      ),
                      // Online toggle
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: SevaColors.successLight,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 8,
                              height: 8,
                              decoration: const BoxDecoration(
                                color: SevaColors.success,
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            const Text(
                              'Online',
                              style: TextStyle(
                                color: SevaColors.success,
                                fontWeight: FontWeight.w600,
                                fontSize: 12,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              const SliverToBoxAdapter(child: SizedBox(height: 20)),

              // Earnings cards
              SliverToBoxAdapter(
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        child: Column(
                          children: [
                            Row(
                              children: [
                                Expanded(
                                  child: SevaStatCard(
                                    label: 'This Month',
                                    value:
                                        'INR ${_earnings?.totalEarnings.toStringAsFixed(0) ?? "0"}',
                                    icon: Icons.account_balance_wallet,
                                    iconColor: SevaColors.primary,
                                    trend: '+12%',
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: SevaStatCard(
                                    label: 'Available',
                                    value:
                                        'INR ${_earnings?.availableBalance.toStringAsFixed(0) ?? "0"}',
                                    icon: Icons.payments,
                                    iconColor: SevaColors.success,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 12),
                            Row(
                              children: [
                                Expanded(
                                  child: SevaStatCard(
                                    label: 'Jobs Done',
                                    value:
                                        '${_earnings?.jobsCompleted ?? 0}',
                                    icon: Icons.check_circle,
                                    iconColor: SevaColors.secondary,
                                    trend: '+5',
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: SevaStatCard(
                                    label: 'Trust Score',
                                    value: '87',
                                    icon: Icons.verified_user,
                                    iconColor: SevaColors.trustVeryGood,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
              ),

              const SliverToBoxAdapter(child: SizedBox(height: 24)),

              // Quick actions
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Text(
                    'Quick Actions',
                    style: Theme.of(context)
                        .textTheme
                        .headlineSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 12)),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    children: [
                      _DashAction(
                        icon: Icons.search,
                        label: 'Find Jobs',
                        onTap: () => context.go('/jobs'),
                      ),
                      const SizedBox(width: 12),
                      _DashAction(
                        icon: Icons.route,
                        label: 'My Routes',
                        onTap: () => context.go('/routes'),
                      ),
                      const SizedBox(width: 12),
                      _DashAction(
                        icon: Icons.payment,
                        label: 'Request Payout',
                        onTap: () => context.go('/earnings'),
                      ),
                      const SizedBox(width: 12),
                      _DashAction(
                        icon: Icons.schedule,
                        label: 'Availability',
                        onTap: () {},
                      ),
                    ],
                  ),
                ),
              ),

              const SliverToBoxAdapter(child: SizedBox(height: 24)),

              // Active jobs
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Active Jobs',
                        style: Theme.of(context)
                            .textTheme
                            .headlineSmall
                            ?.copyWith(fontWeight: FontWeight.w700),
                      ),
                      TextButton(
                        onPressed: () => context.go('/jobs'),
                        child: const Text('View All'),
                      ),
                    ],
                  ),
                ),
              ),

              if (_activeJobs.isEmpty && !_isLoading)
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.all(40),
                    child: Column(
                      children: [
                        Icon(Icons.work_off_outlined,
                            size: 64, color: SevaColors.neutral300),
                        const SizedBox(height: 12),
                        Text(
                          'No active jobs',
                          style:
                              Theme.of(context).textTheme.bodyLarge?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Browse available jobs to get started',
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                      ],
                    ),
                  ),
                )
              else
                SliverPadding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) {
                        final job = _activeJobs[index];
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: JobCard(
                            title: job.title,
                            status: job.status.toJson(),
                            categoryName: job.categoryName,
                            createdAt: job.createdAt,
                            scheduledAt: job.scheduledAt,
                            priceDisplay: job.budgetDisplay,
                            customerName: job.customerName,
                            address: job.address,
                            urgency: job.urgency.name,
                            onTap: () => context.push('/job/${job.id}'),
                          ),
                        );
                      },
                      childCount: _activeJobs.length,
                    ),
                  ),
                ),

              const SliverToBoxAdapter(child: SizedBox(height: 24)),
            ],
          ),
        ),
      ),
    );
  }
}

class _DashAction extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _DashAction({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 14),
          decoration: BoxDecoration(
            color: SevaColors.primaryFaded,
            borderRadius: BorderRadius.circular(12),
            border:
                Border.all(color: SevaColors.primary.withValues(alpha: 0.15)),
          ),
          child: Column(
            children: [
              Icon(icon, color: SevaColors.primary, size: 24),
              const SizedBox(height: 6),
              Text(
                label,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  color: SevaColors.primaryDark,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
