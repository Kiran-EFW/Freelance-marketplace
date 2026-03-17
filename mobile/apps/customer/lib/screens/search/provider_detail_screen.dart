import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class ProviderDetailScreen extends ConsumerStatefulWidget {
  final String providerId;

  const ProviderDetailScreen({super.key, required this.providerId});

  @override
  ConsumerState<ProviderDetailScreen> createState() =>
      _ProviderDetailScreenState();
}

class _ProviderDetailScreenState extends ConsumerState<ProviderDetailScreen>
    with SingleTickerProviderStateMixin {
  ServiceProvider? _provider;
  List<Review> _reviews = [];
  bool _isLoading = true;
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    final providerRepo = ref.read(providerRepositoryProvider);
    final results = await Future.wait([
      providerRepo.getProvider(widget.providerId),
      providerRepo.getProviderReviews(widget.providerId),
    ]);

    if (mounted) {
      setState(() {
        _provider = results[0] as ServiceProvider?;
        _reviews = (results[1] as PaginatedResult<Review>).items;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    if (_provider == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: Text('Provider not found')),
      );
    }

    final provider = _provider!;

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          // Profile header
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      SevaColors.primary,
                      SevaColors.primary.withValues(alpha: 0.8),
                    ],
                  ),
                ),
                child: SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(20, 60, 20, 20),
                    child: Row(
                      children: [
                        // Avatar
                        CircleAvatar(
                          radius: 40,
                          backgroundColor: Colors.white.withValues(alpha: 0.3),
                          backgroundImage: provider.avatarUrl != null
                              ? CachedNetworkImageProvider(provider.avatarUrl!)
                              : null,
                          child: provider.avatarUrl == null
                              ? Text(
                                  provider.name[0].toUpperCase(),
                                  style: const TextStyle(
                                    fontSize: 32,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                )
                              : null,
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Row(
                                children: [
                                  Flexible(
                                    child: Text(
                                      provider.name,
                                      style: const TextStyle(
                                        fontSize: 22,
                                        fontWeight: FontWeight.w700,
                                        color: Colors.white,
                                      ),
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                  if (provider.isVerified) ...[
                                    const SizedBox(width: 6),
                                    const Icon(
                                      Icons.verified,
                                      color: Colors.white,
                                      size: 20,
                                    ),
                                  ],
                                ],
                              ),
                              const SizedBox(height: 4),
                              Row(
                                children: [
                                  StarRating(
                                    rating: provider.rating,
                                    size: 16,
                                    filledColor: Colors.white,
                                    emptyColor:
                                        Colors.white.withValues(alpha: 0.4),
                                  ),
                                  const SizedBox(width: 6),
                                  Text(
                                    '${provider.rating.toStringAsFixed(1)} (${provider.reviewCount})',
                                    style: TextStyle(
                                      color:
                                          Colors.white.withValues(alpha: 0.9),
                                      fontSize: 13,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(
                                '${provider.completedJobs} jobs completed',
                                style: TextStyle(
                                  color: Colors.white.withValues(alpha: 0.8),
                                  fontSize: 13,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),

          // Stats row
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Row(
                children: [
                  _StatItem(
                    label: 'Trust Score',
                    value: '${provider.trustScore.round()}',
                    color: SevaColors.trustColor(provider.trustScore),
                  ),
                  _StatItem(
                    label: 'Response Time',
                    value: provider.responseTimeMinutes < 60
                        ? '${provider.responseTimeMinutes}m'
                        : '${provider.responseTimeMinutes ~/ 60}h',
                    color: SevaColors.info,
                  ),
                  if (provider.distanceKm != null)
                    _StatItem(
                      label: 'Distance',
                      value: provider.distanceDisplay,
                      color: SevaColors.secondary,
                    ),
                ],
              ),
            ),
          ),

          // Tabs
          SliverPersistentHeader(
            pinned: true,
            delegate: _TabBarDelegate(
              TabBar(
                controller: _tabController,
                labelColor: SevaColors.primary,
                unselectedLabelColor: SevaColors.textTertiary,
                indicatorColor: SevaColors.primary,
                tabs: const [
                  Tab(text: 'About'),
                  Tab(text: 'Reviews'),
                  Tab(text: 'Availability'),
                ],
              ),
            ),
          ),

          // Tab content
          SliverFillRemaining(
            child: TabBarView(
              controller: _tabController,
              children: [
                _AboutTab(provider: provider),
                _ReviewsTab(reviews: _reviews),
                _AvailabilityTab(provider: provider),
              ],
            ),
          ),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: SevaButton(
            label: 'Book This Provider',
            icon: Icons.calendar_today_outlined,
            onPressed: () {
              context.push(
                '/job/create?provider=${provider.id}',
              );
            },
          ),
        ),
      ),
    );
  }
}

class _StatItem extends StatelessWidget {
  final String label;
  final String value;
  final Color color;

  const _StatItem({
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Column(
        children: [
          Text(
            value,
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: color,
                ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: SevaColors.textTertiary,
                ),
          ),
        ],
      ),
    );
  }
}

class _TabBarDelegate extends SliverPersistentHeaderDelegate {
  final TabBar tabBar;

  _TabBarDelegate(this.tabBar);

  @override
  double get minExtent => tabBar.preferredSize.height;

  @override
  double get maxExtent => tabBar.preferredSize.height;

  @override
  Widget build(context, shrinkOffset, overlapsContent) {
    return Container(
      color: Theme.of(context).scaffoldBackgroundColor,
      child: tabBar,
    );
  }

  @override
  bool shouldRebuild(covariant _TabBarDelegate oldDelegate) => false;
}

class _AboutTab extends StatelessWidget {
  final ServiceProvider provider;

  const _AboutTab({required this.provider});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (provider.bio != null && provider.bio!.isNotEmpty) ...[
            Text('About',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Text(
              provider.bio!,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: SevaColors.textSecondary,
                    height: 1.5,
                  ),
            ),
            const SizedBox(height: 20),
          ],
          Text('Skills',
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: provider.skills.map((skill) {
              return Chip(
                label: Text(skill),
                backgroundColor: SevaColors.primaryFaded,
                labelStyle: const TextStyle(
                  color: SevaColors.primaryDark,
                  fontSize: 12,
                ),
              );
            }).toList(),
          ),
          if (provider.categories.isNotEmpty) ...[
            const SizedBox(height: 20),
            Text('Categories',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: provider.categories.map((cat) {
                return Chip(
                  label: Text(cat.name),
                  backgroundColor: SevaColors.secondaryFaded,
                  labelStyle: const TextStyle(
                    color: SevaColors.secondaryDark,
                    fontSize: 12,
                  ),
                );
              }).toList(),
            ),
          ],
          if (provider.hourlyRate != null) ...[
            const SizedBox(height: 20),
            Text('Pricing',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Text(
              '${provider.currency ?? "INR"} ${provider.hourlyRate!.toStringAsFixed(0)} / hour',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    color: SevaColors.primary,
                    fontWeight: FontWeight.w700,
                  ),
            ),
          ],
        ],
      ),
    );
  }
}

class _ReviewsTab extends StatelessWidget {
  final List<Review> reviews;

  const _ReviewsTab({required this.reviews});

  @override
  Widget build(BuildContext context) {
    if (reviews.isEmpty) {
      return Center(
        child: Text(
          'No reviews yet',
          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                color: SevaColors.textTertiary,
              ),
        ),
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.all(20),
      itemCount: reviews.length,
      separatorBuilder: (_, __) => const Divider(height: 24),
      itemBuilder: (context, index) {
        final review = reviews[index];
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 16,
                  backgroundColor: SevaColors.neutral200,
                  backgroundImage: review.reviewerAvatarUrl != null
                      ? CachedNetworkImageProvider(review.reviewerAvatarUrl!)
                      : null,
                  child: review.reviewerAvatarUrl == null
                      ? Text(
                          (review.reviewerName ?? '?')[0].toUpperCase(),
                          style: const TextStyle(fontSize: 12),
                        )
                      : null,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        review.reviewerName ?? 'Anonymous',
                        style: Theme.of(context).textTheme.titleSmall,
                      ),
                      Text(
                        review.timeAgo,
                        style: Theme.of(context)
                            .textTheme
                            .bodySmall
                            ?.copyWith(color: SevaColors.textTertiary),
                      ),
                    ],
                  ),
                ),
                StarRating(rating: review.rating.toDouble(), size: 14),
              ],
            ),
            if (review.comment != null && review.comment!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                review.comment!,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: SevaColors.textSecondary,
                      height: 1.4,
                    ),
              ),
            ],
          ],
        );
      },
    );
  }
}

class _AvailabilityTab extends StatelessWidget {
  final ServiceProvider provider;

  const _AvailabilityTab({required this.provider});

  static const _dayNames = [
    'Sunday',
    'Monday',
    'Tuesday',
    'Wednesday',
    'Thursday',
    'Friday',
    'Saturday',
  ];

  @override
  Widget build(BuildContext context) {
    if (provider.availability.isEmpty) {
      return Center(
        child: Text(
          'Availability not set',
          style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                color: SevaColors.textTertiary,
              ),
        ),
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.all(20),
      itemCount: 7,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final slots = provider.availability
            .where((s) => s.dayOfWeek == index)
            .toList();

        return ListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(
            _dayNames[index],
            style: Theme.of(context).textTheme.titleSmall,
          ),
          trailing: slots.isEmpty
              ? Text(
                  'Unavailable',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: SevaColors.textTertiary,
                      ),
                )
              : Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: slots
                      .map((s) => Text(
                            '${s.startTime} - ${s.endTime}',
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(
                                  color: SevaColors.success,
                                  fontWeight: FontWeight.w500,
                                ),
                          ))
                      .toList(),
                ),
        );
      },
    );
  }
}
