import 'dart:async';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';

import '../../main.dart';

/// Map view showing nearby service providers as markers.
class MapScreen extends ConsumerStatefulWidget {
  const MapScreen({super.key});

  @override
  ConsumerState<MapScreen> createState() => _MapScreenState();
}

class _MapScreenState extends ConsumerState<MapScreen> {
  GoogleMapController? _mapController;
  final Completer<GoogleMapController> _controllerCompleter = Completer();

  // Default location (Bangalore, India) until device location is obtained.
  static const LatLng _defaultLocation = LatLng(12.9716, 77.5946);

  LatLng _currentCenter = _defaultLocation;
  LatLng? _userLocation;
  int _radiusKm = 10;
  String? _selectedCategoryId;
  List<Category> _categories = [];
  List<ServiceProvider> _providers = [];
  Set<Marker> _markers = {};
  ServiceProvider? _selectedProvider;
  bool _isLoading = false;
  bool _showSearchButton = false;

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  Future<void> _loadInitialData() async {
    setState(() => _isLoading = true);

    // Load categories for the filter.
    final providerRepo = ref.read(providerRepositoryProvider);
    final categoriesResult = await providerRepo.getCategories();
    if (categoriesResult.isSuccess) {
      _categories = categoriesResult.dataOrNull ?? [];
    }

    // Get user location.
    final locationService = ref.read(locationServiceProvider);
    final hasPermission = await locationService.checkPermission();
    if (hasPermission) {
      final position = await locationService.getCurrentPosition();
      if (position != null) {
        _userLocation = LatLng(position.latitude, position.longitude);
        _currentCenter = _userLocation!;
      }
    }

    await _searchProviders();

    if (mounted) {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _searchProviders() async {
    setState(() => _isLoading = true);

    final providerRepo = ref.read(providerRepositoryProvider);
    final result = await providerRepo.searchProviders(
      latitude: _currentCenter.latitude,
      longitude: _currentCenter.longitude,
      radiusKm: _radiusKm,
      categoryId: _selectedCategoryId,
      limit: 50,
    );

    if (mounted) {
      switch (result) {
        case Success(:final data):
          _providers = data.items;
          _updateMarkers();
        case Failure(:final message):
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(message)),
          );
      }
      setState(() {
        _isLoading = false;
        _showSearchButton = false;
      });
    }
  }

  void _updateMarkers() {
    _markers = _providers.map((provider) {
      return Marker(
        markerId: MarkerId(provider.id),
        position: LatLng(
          provider.latitude ?? _defaultLocation.latitude,
          provider.longitude ?? _defaultLocation.longitude,
        ),
        infoWindow: InfoWindow(
          title: provider.name,
          snippet: provider.categories.isNotEmpty
              ? provider.categories.first.name
              : null,
        ),
        onTap: () {
          setState(() => _selectedProvider = provider);
        },
      );
    }).toSet();
  }

  void _onMapMoved() {
    if (!_showSearchButton) {
      setState(() => _showSearchButton = true);
    }
  }

  void _onCameraIdle() async {
    final controller = await _controllerCompleter.future;
    final bounds = await controller.getVisibleRegion();
    _currentCenter = LatLng(
      (bounds.northeast.latitude + bounds.southwest.latitude) / 2,
      (bounds.northeast.longitude + bounds.southwest.longitude) / 2,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          // Google Map
          GoogleMap(
            initialCameraPosition: CameraPosition(
              target: _currentCenter,
              zoom: 13,
            ),
            markers: _markers,
            myLocationEnabled: _userLocation != null,
            myLocationButtonEnabled: false,
            zoomControlsEnabled: false,
            mapToolbarEnabled: false,
            onMapCreated: (controller) {
              _mapController = controller;
              _controllerCompleter.complete(controller);
            },
            onCameraMove: (_) => _onMapMoved(),
            onCameraIdle: _onCameraIdle,
            onTap: (_) {
              setState(() => _selectedProvider = null);
            },
          ),

          // Top bar with category filter and radius
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Column(
                children: [
                  // Category filter chips
                  SizedBox(
                    height: 40,
                    child: ListView(
                      scrollDirection: Axis.horizontal,
                      children: [
                        _FilterChip(
                          label: 'All',
                          isSelected: _selectedCategoryId == null,
                          onTap: () {
                            setState(() => _selectedCategoryId = null);
                            _searchProviders();
                          },
                        ),
                        const SizedBox(width: 8),
                        ..._categories.take(8).map((category) {
                          return Padding(
                            padding: const EdgeInsets.only(right: 8),
                            child: _FilterChip(
                              label: category.name,
                              isSelected:
                                  _selectedCategoryId == category.id,
                              onTap: () {
                                setState(
                                  () => _selectedCategoryId = category.id,
                                );
                                _searchProviders();
                              },
                            ),
                          );
                        }),
                      ],
                    ),
                  ),

                  const SizedBox(height: 8),

                  // Radius selector
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withValues(alpha: 0.1),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.tune,
                          size: 16,
                          color: SevaColors.textSecondary,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Radius:',
                          style: Theme.of(context)
                              .textTheme
                              .bodySmall
                              ?.copyWith(color: SevaColors.textSecondary),
                        ),
                        const SizedBox(width: 4),
                        DropdownButton<int>(
                          value: _radiusKm,
                          underline: const SizedBox(),
                          isDense: true,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    fontWeight: FontWeight.w600,
                                    color: SevaColors.primary,
                                  ),
                          items: [5, 10, 15, 25, 50].map((km) {
                            return DropdownMenuItem(
                              value: km,
                              child: Text('$km km'),
                            );
                          }).toList(),
                          onChanged: (value) {
                            if (value != null) {
                              setState(() => _radiusKm = value);
                              _searchProviders();
                            }
                          },
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),

          // "Search This Area" button
          if (_showSearchButton)
            Positioned(
              top: MediaQuery.of(context).padding.top + 110,
              left: 0,
              right: 0,
              child: Center(
                child: Material(
                  elevation: 4,
                  borderRadius: BorderRadius.circular(20),
                  color: SevaColors.primary,
                  child: InkWell(
                    borderRadius: BorderRadius.circular(20),
                    onTap: _searchProviders,
                    child: const Padding(
                      padding:
                          EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.search, color: Colors.white, size: 18),
                          SizedBox(width: 6),
                          Text(
                            'Search This Area',
                            style: TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.w600,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),

          // Loading indicator
          if (_isLoading)
            const Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: LinearProgressIndicator(),
            ),

          // My location FAB
          Positioned(
            right: 16,
            bottom: _selectedProvider != null ? 200 : 24,
            child: FloatingActionButton.small(
              heroTag: 'my_location',
              backgroundColor: Colors.white,
              foregroundColor: SevaColors.primary,
              onPressed: () async {
                if (_userLocation != null && _mapController != null) {
                  _mapController!.animateCamera(
                    CameraUpdate.newLatLngZoom(_userLocation!, 14),
                  );
                }
              },
              child: const Icon(Icons.my_location),
            ),
          ),

          // Selected provider card
          if (_selectedProvider != null)
            Positioned(
              left: 16,
              right: 16,
              bottom: MediaQuery.of(context).padding.bottom + 16,
              child: _ProviderMapCard(
                provider: _selectedProvider!,
                onTap: () {
                  context.push('/provider/${_selectedProvider!.id}');
                },
                onClose: () {
                  setState(() => _selectedProvider = null);
                },
              ),
            ),
        ],
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _FilterChip({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? SevaColors.primary : Colors.white,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.08),
              blurRadius: 6,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: isSelected ? Colors.white : SevaColors.textSecondary,
          ),
        ),
      ),
    );
  }
}

class _ProviderMapCard extends StatelessWidget {
  final ServiceProvider provider;
  final VoidCallback onTap;
  final VoidCallback onClose;

  const _ProviderMapCard({
    required this.provider,
    required this.onTap,
    required this.onClose,
  });

  @override
  Widget build(BuildContext context) {
    return SevaCard(
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              CircleAvatar(
                radius: 24,
                backgroundColor: SevaColors.primaryFaded,
                backgroundImage: provider.avatarUrl != null
                    ? CachedNetworkImageProvider(provider.avatarUrl!)
                    : null,
                child: provider.avatarUrl == null
                    ? Text(
                        provider.name[0].toUpperCase(),
                        style: const TextStyle(
                          color: SevaColors.primary,
                          fontWeight: FontWeight.w700,
                          fontSize: 18,
                        ),
                      )
                    : null,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            provider.name,
                            style: Theme.of(context)
                                .textTheme
                                .titleMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        if (provider.isVerified)
                          const Icon(
                            Icons.verified,
                            size: 18,
                            color: SevaColors.info,
                          ),
                      ],
                    ),
                    const SizedBox(height: 2),
                    if (provider.categories.isNotEmpty)
                      Text(
                        provider.categories.map((c) => c.name).join(', '),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: SevaColors.textTertiary,
                            ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                  ],
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close, size: 20),
                onPressed: onClose,
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              StarRating(rating: provider.rating, size: 16),
              const SizedBox(width: 6),
              Text(
                '${provider.rating.toStringAsFixed(1)} (${provider.reviewCount})',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: SevaColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(width: 12),
              if (provider.distanceKm != null) ...[
                Icon(Icons.location_on_outlined,
                    size: 14, color: SevaColors.textTertiary),
                const SizedBox(width: 2),
                Text(
                  provider.distanceDisplay,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: SevaColors.textTertiary,
                      ),
                ),
              ],
              const Spacer(),
              if (provider.hourlyRate != null)
                Text(
                  '${provider.currency ?? "INR"} ${provider.hourlyRate!.toStringAsFixed(0)}/hr',
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        color: SevaColors.primary,
                        fontWeight: FontWeight.w700,
                      ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
