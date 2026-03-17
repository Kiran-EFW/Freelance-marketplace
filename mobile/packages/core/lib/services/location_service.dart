import 'dart:async';
import 'package:geolocator/geolocator.dart';
import 'package:geocoding/geocoding.dart';

/// Result of a location lookup.
class LocationResult {
  final double latitude;
  final double longitude;
  final String? postcode;
  final String? locality;
  final String? administrativeArea;
  final String? country;
  final String? formattedAddress;

  const LocationResult({
    required this.latitude,
    required this.longitude,
    this.postcode,
    this.locality,
    this.administrativeArea,
    this.country,
    this.formattedAddress,
  });
}

/// Provides device location and reverse geocoding services.
///
/// Handles permission requests, GPS availability checks, and converts
/// coordinates to human-readable addresses and postcodes.
class LocationService {
  /// Check if location services are enabled and we have permission.
  Future<bool> checkPermission() async {
    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }

    return permission == LocationPermission.whileInUse ||
        permission == LocationPermission.always;
  }

  /// Request location permission from the user.
  Future<LocationPermission> requestPermission() async {
    return Geolocator.requestPermission();
  }

  /// Get the device's current position.
  Future<Position?> getCurrentPosition() async {
    final hasPermission = await checkPermission();
    if (!hasPermission) return null;

    try {
      return await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(
          accuracy: LocationAccuracy.high,
          timeLimit: Duration(seconds: 10),
        ),
      );
    } catch (_) {
      return null;
    }
  }

  /// Get the current location with reverse geocoding.
  Future<LocationResult?> getCurrentLocation() async {
    final position = await getCurrentPosition();
    if (position == null) return null;

    return _reverseGeocode(position.latitude, position.longitude);
  }

  /// Reverse geocode a set of coordinates to get address details.
  Future<LocationResult?> reverseGeocode(
    double latitude,
    double longitude,
  ) async {
    return _reverseGeocode(latitude, longitude);
  }

  /// Forward geocode a postcode or address string to coordinates.
  Future<LocationResult?> geocodeAddress(String address) async {
    try {
      final locations = await locationFromAddress(address);
      if (locations.isEmpty) return null;

      final location = locations.first;
      return _reverseGeocode(location.latitude, location.longitude);
    } catch (_) {
      return null;
    }
  }

  /// Calculate the distance in kilometers between two points.
  double distanceBetween(
    double startLat,
    double startLng,
    double endLat,
    double endLng,
  ) {
    return Geolocator.distanceBetween(startLat, startLng, endLat, endLng) /
        1000.0;
  }

  /// Stream position updates for real-time location tracking.
  Stream<Position> getPositionStream({
    int distanceFilterMeters = 50,
  }) {
    return Geolocator.getPositionStream(
      locationSettings: LocationSettings(
        accuracy: LocationAccuracy.high,
        distanceFilter: distanceFilterMeters,
      ),
    );
  }

  Future<LocationResult?> _reverseGeocode(
    double latitude,
    double longitude,
  ) async {
    try {
      final placemarks = await placemarkFromCoordinates(latitude, longitude);
      if (placemarks.isEmpty) {
        return LocationResult(latitude: latitude, longitude: longitude);
      }

      final place = placemarks.first;
      final parts = <String>[];
      if (place.street != null && place.street!.isNotEmpty) {
        parts.add(place.street!);
      }
      if (place.locality != null && place.locality!.isNotEmpty) {
        parts.add(place.locality!);
      }
      if (place.administrativeArea != null &&
          place.administrativeArea!.isNotEmpty) {
        parts.add(place.administrativeArea!);
      }
      if (place.postalCode != null && place.postalCode!.isNotEmpty) {
        parts.add(place.postalCode!);
      }

      return LocationResult(
        latitude: latitude,
        longitude: longitude,
        postcode: place.postalCode,
        locality: place.locality,
        administrativeArea: place.administrativeArea,
        country: place.country,
        formattedAddress: parts.join(', '),
      );
    } catch (_) {
      return LocationResult(latitude: latitude, longitude: longitude);
    }
  }
}
