import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  final _searchController = TextEditingController();
  List<Category> _categories = [];
  List<Job> _recentJobs = [];
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
      providerRepo.getCategories(),
      jobRepo.getJobs(role: 'customer', page: 1, limit: 5),
    ]);

    if (mounted) {
      setState(() {
        _categories = results[0] as List<Category>;
        _recentJobs = (results[1] as PaginatedResult<Job>).items;
        _isLoading = false;
      });
    }
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
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
              // Greeting header
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  _greeting(),
                                  style: Theme.of(context)
                                      .textTheme
                                      .bodyMedium
                                      ?.copyWith(
                                        color: SevaColors.textSecondary,
                                      ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  user?.name ?? 'Welcome',
                                  style: Theme.of(context)
                                      .textTheme
                                      .headlineMedium
                                      ?.copyWith(
                                        fontWeight: FontWeight.w700,
                                      ),
                                ),
                              ],
                            ),
                          ),
                          // Notification bell
                          IconButton(
                            onPressed: () => context.go('/notifications'),
                            icon: const Icon(Icons.notifications_outlined),
                          ),
                        ],
                      ),
                      const SizedBox(height: 20),
                    ],
                  ),
                ),
              ),

              // Search bar
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: GestureDetector(
                    onTap: () => context.go('/search'),
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      decoration: BoxDecoration(
                        color: SevaColors.neutral100,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: SevaColors.neutral200),
                      ),
                      child: Row(
                        children: [
                          const Icon(
                            Icons.search,
                            color: SevaColors.neutral400,
                          ),
                          const SizedBox(width: 12),
                          Text(
                            'What service do you need?',
                            style: Theme.of(context)
                                .textTheme
                                .bodyMedium
                                ?.copyWith(
                                  color: SevaColors.textTertiary,
                                ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),

              const SliverToBoxAdapter(child: SizedBox(height: 28)),

              // Quick actions
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    children: [
                      _QuickAction(
                        icon: Icons.camera_alt_outlined,
                        label: 'Photo\nDiagnose',
                        color: SevaColors.secondary,
                        onTap: () {
                          // Navigate to photo analysis
                        },
                      ),
                      const SizedBox(width: 12),
                      _QuickAction(
                        icon: Icons.bolt_outlined,
                        label: 'Emergency\nService',
                        color: SevaColors.error,
                        onTap: () {
                          context.go('/search?urgency=emergency');
                        },
                      ),
                      const SizedBox(width: 12),
                      _QuickAction(
                        icon: Icons.star_outline,
                        label: 'Top\nRated',
                        color: SevaColors.starFilled,
                        onTap: () {
                          context.go('/search?sort=rating');
                        },
                      ),
                      const SizedBox(width: 12),
                      _QuickAction(
                        icon: Icons.near_me_outlined,
                        label: 'Near\nMe',
                        color: SevaColors.info,
                        onTap: () {
                          context.go('/search?sort=distance');
                        },
                      ),
                    ],
                  ),
                ),
              ),

              const SliverToBoxAdapter(child: SizedBox(height: 28)),

              // Popular categories
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Text(
                    'Popular Categories',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                  ),
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 12)),

              if (_isLoading)
                const SliverToBoxAdapter(
                  child: Center(child: CircularProgressIndicator()),
                )
              else
                SliverPadding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  sliver: SliverGrid(
                    gridDelegate:
                        const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 3,
                      crossAxisSpacing: 12,
                      mainAxisSpacing: 12,
                      childAspectRatio: 1.0,
                    ),
                    delegate: SliverChildBuilderDelegate(
                      (context, index) {
                        if (index >= _categories.length) return null;
                        final category = _categories[index];
                        return _CategoryTile(
                          name: category.name,
                          providerCount: category.providerCount,
                          onTap: () {
                            context.go('/search?category=${category.id}');
                          },
                        );
                      },
                      childCount: _categories.length.clamp(0, 9),
                    ),
                  ),
                ),

              const SliverToBoxAdapter(child: SizedBox(height: 28)),

              // Recent activity
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Recent Activity',
                        style: Theme.of(context)
                            .textTheme
                            .headlineSmall
                            ?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                      ),
                      if (_recentJobs.isNotEmpty)
                        TextButton(
                          onPressed: () {
                            // Navigate to all jobs
                          },
                          child: const Text('See All'),
                        ),
                    ],
                  ),
                ),
              ),

              if (_recentJobs.isEmpty && !_isLoading)
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.all(40),
                    child: Column(
                      children: [
                        Icon(
                          Icons.work_outline,
                          size: 64,
                          color: SevaColors.neutral300,
                        ),
                        const SizedBox(height: 12),
                        Text(
                          'No recent jobs',
                          style:
                              Theme.of(context).textTheme.bodyLarge?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                        const SizedBox(height: 8),
                        SevaButton(
                          label: 'Find a Provider',
                          isFullWidth: false,
                          size: SevaButtonSize.small,
                          onPressed: () => context.go('/search'),
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
                        final job = _recentJobs[index];
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: JobCard(
                            title: job.title,
                            status: job.status.toJson(),
                            categoryName: job.categoryName,
                            createdAt: job.createdAt,
                            scheduledAt: job.scheduledAt,
                            priceDisplay: job.budgetDisplay,
                            providerName: job.providerName,
                            address: job.address,
                            onTap: () => context.push('/job/${job.id}'),
                          ),
                        );
                      },
                      childCount: _recentJobs.length,
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

  String _greeting() {
    final hour = DateTime.now().hour;
    if (hour < 12) return 'Good Morning';
    if (hour < 17) return 'Good Afternoon';
    return 'Good Evening';
  }
}

class _QuickAction extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _QuickAction({
    required this.icon,
    required this.label,
    required this.color,
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
            color: color.withValues(alpha: 0.08),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: color.withValues(alpha: 0.2)),
          ),
          child: Column(
            children: [
              Icon(icon, color: color, size: 28),
              const SizedBox(height: 6),
              Text(
                label,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  color: color,
                  height: 1.2,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  final String name;
  final int providerCount;
  final VoidCallback onTap;

  const _CategoryTile({
    required this.name,
    required this.providerCount,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: SevaColors.primaryFaded,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: SevaColors.primary.withValues(alpha: 0.15)),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.build_outlined,
              color: SevaColors.primary,
              size: 28,
            ),
            const SizedBox(height: 6),
            Text(
              name,
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: SevaColors.primaryDark,
                  ),
            ),
            if (providerCount > 0)
              Text(
                '$providerCount providers',
                style: TextStyle(
                  fontSize: 9,
                  color: SevaColors.textTertiary,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
