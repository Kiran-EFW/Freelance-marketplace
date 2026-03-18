import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class RouteListScreen extends ConsumerStatefulWidget {
  const RouteListScreen({super.key});

  @override
  ConsumerState<RouteListScreen> createState() => _RouteListScreenState();
}

class _RouteListScreenState extends ConsumerState<RouteListScreen> {
  List<Map<String, dynamic>> _routes = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadRoutes();
  }

  Future<void> _loadRoutes() async {
    setState(() => _isLoading = true);

    try {
      final apiClient = ref.read(apiClientProvider);
      final response = await apiClient.getRoutes();
      final data = response.data as Map<String, dynamic>;
      final items = data['items'] as List<dynamic>;

      if (mounted) {
        setState(() {
          _routes = items.cast<Map<String, dynamic>>();
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _routes = [];
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('My Routes'),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _routes.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.route, size: 64, color: SevaColors.neutral300),
                      const SizedBox(height: 12),
                      Text(
                        'No routes yet',
                        style: Theme.of(context)
                            .textTheme
                            .bodyLarge
                            ?.copyWith(color: SevaColors.textTertiary),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Routes are created when you have multiple jobs in a day',
                        textAlign: TextAlign.center,
                        style: Theme.of(context)
                            .textTheme
                            .bodySmall
                            ?.copyWith(color: SevaColors.textTertiary),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadRoutes,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _routes.length,
                    itemBuilder: (context, index) {
                      final route = _routes[index];
                      final stops =
                          (route['stops'] as List<dynamic>?)?.length ?? 0;

                      return Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: SevaCard(
                          onTap: () {
                            context.push('/route/${route['id']}');
                          },
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: SevaColors.primary
                                          .withValues(alpha: 0.1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: const Icon(
                                      Icons.route,
                                      color: SevaColors.primary,
                                      size: 20,
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          route['name'] as String? ??
                                              'Route ${index + 1}',
                                          style: Theme.of(context)
                                              .textTheme
                                              .titleMedium,
                                        ),
                                        Text(
                                          '$stops stops',
                                          style: Theme.of(context)
                                              .textTheme
                                              .bodySmall
                                              ?.copyWith(
                                                color:
                                                    SevaColors.textTertiary,
                                              ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  StatusBadge(
                                    status:
                                        route['status'] as String? ?? 'pending',
                                    isCompact: true,
                                  ),
                                ],
                              ),
                              if (route['estimated_duration'] != null) ...[
                                const SizedBox(height: 8),
                                Row(
                                  children: [
                                    const Icon(Icons.access_time,
                                        size: 14,
                                        color: SevaColors.textTertiary),
                                    const SizedBox(width: 4),
                                    Text(
                                      'Est. ${route['estimated_duration']} min',
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodySmall
                                          ?.copyWith(
                                            color: SevaColors.textTertiary,
                                          ),
                                    ),
                                    const SizedBox(width: 16),
                                    const Icon(Icons.straighten,
                                        size: 14,
                                        color: SevaColors.textTertiary),
                                    const SizedBox(width: 4),
                                    Text(
                                      '${route['estimated_distance_km'] ?? "?"} km',
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodySmall
                                          ?.copyWith(
                                            color: SevaColors.textTertiary,
                                          ),
                                    ),
                                  ],
                                ),
                              ],
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}
