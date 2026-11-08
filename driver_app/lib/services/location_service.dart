import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';

class LocationException implements Exception {
  LocationException(this.message);
  final String message;
  @override
  String toString() => message;
}

class LocationService {
  Future<void> ensurePermission() async {
    if (!await Geolocator.isLocationServiceEnabled()) {
      throw LocationException('Location services are turned off. Enable GPS and try again.');
    }
    var perm = await Geolocator.checkPermission();
    if (perm == LocationPermission.denied) {
      perm = await Geolocator.requestPermission();
    }
    if (perm == LocationPermission.denied) {
      throw LocationException('Location permission denied.');
    }
    if (perm == LocationPermission.deniedForever) {
      throw LocationException('Location permission is blocked. Enable it in system settings.');
    }
  }

  /// Continuous GPS stream. On Android the foreground notification turns this into a
  /// foreground service, so it keeps running with the screen locked or app in background.
  Stream<Position> positions() =>
      Geolocator.getPositionStream(locationSettings: _settings());

  LocationSettings _settings() {
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return AndroidSettings(
          accuracy: LocationAccuracy.high,
          distanceFilter: 0,
          intervalDuration: const Duration(seconds: 3),
          foregroundNotificationConfig: const ForegroundNotificationConfig(
            notificationTitle: 'School bus tracking is active',
            notificationText: 'Sharing live location with parents and school',
            enableWakeLock: true,
            setOngoing: true,
          ),
        );
      case TargetPlatform.iOS:
        return AppleSettings(
          accuracy: LocationAccuracy.high,
          activityType: ActivityType.automotiveNavigation,
          distanceFilter: 0,
          pauseLocationUpdatesAutomatically: false,
          showBackgroundLocationIndicator: true,
          allowBackgroundLocationUpdates: true,
        );
      default:
        return const LocationSettings(accuracy: LocationAccuracy.high);
    }
  }
}