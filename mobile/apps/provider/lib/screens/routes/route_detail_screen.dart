import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class RouteDetailScreen extends ConsumerStatefulWidget {
  final String routeId;

  const RouteDetailScreen({super.key, required this.routeId});

  @override
  ConsumerState<RouteDetailScreen> createState() => _RouteDetailScreenState();
}

class _RouteDetailScreenState extends ConsumerState<RouteDetailScreen> {
  Map<String, dynamic>? _route;
  bool _isLoading = true;
  bool _isOptimizing = false;

  @override
  void initState() {
    super.initState();
    _loadRoute();
  }

  Future<void> _loadRoute() async {
    setState(() => _isLoading = true);

    try {
      final apiClient = ref.read(apiClientProvider);
      final response = await apiClient.getRoute(widget.routeId);
      if (mounted) {
        setState(() {
          _route = response.data as Map<String, dynamic>;
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _optimizeRoute() async {
    setState(() => _isOptimizing = true);

    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.optimizeRoute(widget.routeId);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Route optimized!')),
        );
        _loadRoute();
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to optimize route')),
        );
      }
    } finally {
      if (mounted) setState(() => _isOptimizing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    if (_route == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: Text('Route not found')),
      );
    }

    final stops = (_route!['stops'] as List<dynamic>?) ?? [];

    return Scaffold(
      appBar: AppBar(
        title: Text(_route!['name'] as String? ?? 'Route'),
        actions: [
          TextButton.icon(
            onPressed: _isOptimizing ? null : _optimizeRoute,
            icon: _isOptimizing
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.auto_fix_high),
            label: const Text('Optimize'),
          ),
        ],
      ),
      body: Column(
        children: [
          // Route summary
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              children: [
                Expanded(
                  child: SevaStatCard(
                    label: 'Stops',
                    value: '${stops.length}',
                    icon: Icons.pin_drop,
                    iconColor: SevaColors.primary,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: SevaStatCard(
                    label: 'Distance',
                    value:
                        '${_route!['estimated_distance_km'] ?? "?"} km',
                    icon: Icons.straighten,
                    iconColor: SevaColors.info,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: SevaStatCard(
                    label: 'Duration',
                    value:
                        '${_route!['estimated_duration'] ?? "?"} min',
                    icon: Icons.access_time,
                    iconColor: SevaColors.secondary,
                  ),
                ),
              ],
            ),
          ),

          // Stops list
          Expanded(
            child: stops.isEmpty
                ? Center(
                    child: Text(
                      'No stops in this route',
                      style: Theme.of(context)
                          .textTheme
                          .bodyLarge
                          ?.copyWith(color: SevaColors.textTertiary),
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    itemCount: stops.length,
                    itemBuilder: (context, index) {
                      final stop = stops[index] as Map<String, dynamic>;
                      final isFirst = index == 0;
                      final isLast = index == stops.length - 1;

                      return IntrinsicHeight(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Timeline indicator
                            SizedBox(
                              width: 40,
                              child: Column(
                                children: [
                                  if (!isFirst)
                                    Container(
                                      width: 2,
                                      height: 12,
                                      color: SevaColors.neutral300,
                                    ),
                                  Container(
                                    width: 32,
                                    height: 32,
                                    decoration: BoxDecoration(
                                      color: SevaColors.primary
                                          .withValues(alpha: 0.1),
                                      shape: BoxShape.circle,
                                      border: Border.all(
                                          color: SevaColors.primary, width: 2),
                                    ),
                                    child: Center(
                                      child: Text(
                                        '${index + 1}',
                                        style: const TextStyle(
                                          color: SevaColors.primary,
                                          fontWeight: FontWeight.w700,
                                          fontSize: 14,
                                        ),
                                      ),
                                    ),
                                  ),
                                  if (!isLast)
                                    Expanded(
                                      child: Container(
                                        width: 2,
                                        color: SevaColors.neutral300,
                                      ),
                                    ),
                                ],
                              ),
                            ),
                            const SizedBox(width: 12),

                            // Stop details
                            Expanded(
                              child: Padding(
                                padding:
                                    const EdgeInsets.only(bottom: 16),
                                child: SevaCard(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        stop['title'] as String? ??
                                            'Stop ${index + 1}',
                                        style: Theme.of(context)
                                            .textTheme
                                            .titleSmall,
                                      ),
                                      const SizedBox(height: 4),
                                      if (stop['address'] != null)
                                        Row(
                                          children: [
                                            const Icon(
                                              Icons.location_on_outlined,
                                              size: 14,
                                              color:
                                                  SevaColors.textTertiary,
                                            ),
                                            const SizedBox(width: 4),
                                            Expanded(
                                              child: Text(
                                                stop['address'] as String,
                                                style: Theme.of(context)
                                                    .textTheme
                                                    .bodySmall
                                                    ?.copyWith(
                                                      color: SevaColors
                                                          .textTertiary,
                                                    ),
                                                maxLines: 1,
                                                overflow:
                                                    TextOverflow.ellipsis,
                                              ),
                                            ),
                                          ],
                                        ),
                                      if (stop['estimated_time'] !=
                                          null) ...[
                                        const SizedBox(height: 2),
                                        Row(
                                          children: [
                                            const Icon(
                                              Icons.access_time,
                                              size: 14,
                                              color:
                                                  SevaColors.textTertiary,
                                            ),
                                            const SizedBox(width: 4),
                                            Text(
                                              stop['estimated_time']
                                                  as String,
                                              style: Theme.of(context)
                                                  .textTheme
                                                  .bodySmall
                                                  ?.copyWith(
                                                    color: SevaColors
                                                        .textTertiary,
                                                  ),
                                            ),
                                          ],
                                        ),
                                      ],
                                      if (stop['status'] != null) ...[
                                        const SizedBox(height: 8),
                                        StatusBadge(
                                          status: stop['status'] as String,
                                          isCompact: true,
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
