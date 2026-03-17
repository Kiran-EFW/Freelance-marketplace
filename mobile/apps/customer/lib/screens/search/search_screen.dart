import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class SearchScreen extends ConsumerStatefulWidget {
  final String? initialQuery;
  final String? categoryId;

  const SearchScreen({
    super.key,
    this.initialQuery,
    this.categoryId,
  });

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final _searchController = TextEditingController();
  final _scrollController = ScrollController();

  List<ServiceProvider> _providers = [];
  bool _isLoading = false;
  bool _hasMore = true;
  int _page = 1;

  // Filter state
  double? _minRating;
  int? _radiusKm;
  String? _sortBy;
  String? _selectedCategoryId;

  @override
  void initState() {
    super.initState();
    _searchController.text = widget.initialQuery ?? '';
    _selectedCategoryId = widget.categoryId;
    _scrollController.addListener(_onScroll);
    _search();
  }

  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 200 &&
        !_isLoading &&
        _hasMore) {
      _loadMore();
    }
  }

  Future<void> _search() async {
    setState(() {
      _isLoading = true;
      _page = 1;
      _providers = [];
    });
    await _fetchProviders();
  }

  Future<void> _loadMore() async {
    _page++;
    await _fetchProviders();
  }

  Future<void> _fetchProviders() async {
    setState(() => _isLoading = true);

    final providerRepo = ref.read(providerRepositoryProvider);
    final locationService = ref.read(locationServiceProvider);

    double? lat, lng;
    final position = await locationService.getCurrentPosition();
    if (position != null) {
      lat = position.latitude;
      lng = position.longitude;
    }

    final result = await providerRepo.searchProviders(
      query: _searchController.text.isNotEmpty
          ? _searchController.text
          : null,
      categoryId: _selectedCategoryId,
      latitude: lat,
      longitude: lng,
      radiusKm: _radiusKm,
      minRating: _minRating,
      page: _page,
      sortBy: _sortBy,
    );

    if (mounted) {
      setState(() {
        if (_page == 1) {
          _providers = result.items;
        } else {
          _providers.addAll(result.items);
        }
        _hasMore = result.hasMore;
        _isLoading = false;
      });
    }
  }

  void _showFilters() {
    showModalBottomSheet(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return SafeArea(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Filters',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 20),

                    // Minimum rating
                    Text('Minimum Rating',
                        style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      children: [3.0, 3.5, 4.0, 4.5].map((rating) {
                        final isSelected = _minRating == rating;
                        return FilterChip(
                          label: Text('${rating}+'),
                          selected: isSelected,
                          onSelected: (selected) {
                            setModalState(() {
                              _minRating = selected ? rating : null;
                            });
                          },
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 16),

                    // Distance
                    Text('Distance',
                        style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      children: [5, 10, 25, 50].map((km) {
                        final isSelected = _radiusKm == km;
                        return FilterChip(
                          label: Text('${km} km'),
                          selected: isSelected,
                          onSelected: (selected) {
                            setModalState(() {
                              _radiusKm = selected ? km : null;
                            });
                          },
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 16),

                    // Sort by
                    Text('Sort By',
                        style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      children: {
                        'rating': 'Top Rated',
                        'distance': 'Nearest',
                        'trust_score': 'Most Trusted',
                        'response_time': 'Fastest Response',
                      }.entries.map((entry) {
                        final isSelected = _sortBy == entry.key;
                        return FilterChip(
                          label: Text(entry.value),
                          selected: isSelected,
                          onSelected: (selected) {
                            setModalState(() {
                              _sortBy = selected ? entry.key : null;
                            });
                          },
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 24),

                    Row(
                      children: [
                        Expanded(
                          child: SevaButton(
                            label: 'Clear All',
                            variant: SevaButtonVariant.outline,
                            onPressed: () {
                              setModalState(() {
                                _minRating = null;
                                _radiusKm = null;
                                _sortBy = null;
                              });
                            },
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: SevaButton(
                            label: 'Apply',
                            onPressed: () {
                              Navigator.pop(context);
                              _search();
                            },
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Find Providers'),
      ),
      body: Column(
        children: [
          // Search bar
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 20, 12),
            child: Row(
              children: [
                Expanded(
                  child: SevaInput(
                    hint: 'Search services or providers...',
                    controller: _searchController,
                    prefixIcon: Icons.search,
                    onChanged: (_) {
                      // Debounce search
                      Future.delayed(const Duration(milliseconds: 500), () {
                        if (mounted) _search();
                      });
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  decoration: BoxDecoration(
                    border: Border.all(color: SevaColors.neutral300),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: IconButton(
                    onPressed: _showFilters,
                    icon: const Icon(Icons.tune),
                    color: _hasActiveFilters
                        ? SevaColors.primary
                        : SevaColors.neutral500,
                  ),
                ),
              ],
            ),
          ),

          // Results count
          if (!_isLoading || _providers.isNotEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Row(
                children: [
                  Text(
                    '${_providers.length} providers found',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: SevaColors.textTertiary,
                        ),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 8),

          // Results list
          Expanded(
            child: _isLoading && _providers.isEmpty
                ? const Center(child: CircularProgressIndicator())
                : _providers.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.search_off,
                              size: 64,
                              color: SevaColors.neutral300,
                            ),
                            const SizedBox(height: 12),
                            Text(
                              'No providers found',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodyLarge
                                  ?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              'Try adjusting your search or filters',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        controller: _scrollController,
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        itemCount: _providers.length + (_hasMore ? 1 : 0),
                        itemBuilder: (context, index) {
                          if (index == _providers.length) {
                            return const Padding(
                              padding: EdgeInsets.all(20),
                              child: Center(
                                child: CircularProgressIndicator(),
                              ),
                            );
                          }
                          final provider = _providers[index];
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 12),
                            child: ProviderCard(
                              name: provider.name,
                              avatarUrl: provider.avatarUrl,
                              rating: provider.rating,
                              reviewCount: provider.reviewCount,
                              distanceKm: provider.distanceKm,
                              skills: provider.skills,
                              trustScore: provider.trustScore,
                              isVerified: provider.isVerified,
                              hourlyRate: provider.hourlyRate?.toStringAsFixed(0),
                              currency: provider.currency,
                              responseTimeMinutes: provider.responseTimeMinutes,
                              onTap: () {
                                context.push('/provider/${provider.id}');
                              },
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  bool get _hasActiveFilters =>
      _minRating != null || _radiusKm != null || _sortBy != null;
}
