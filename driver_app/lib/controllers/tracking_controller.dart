import 'dart:async';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';

import '../config/config.dart';
import '../services/local_buffer.dart';
import '../models/ping.dart';
import '../services/location_service.dart';
import '../services/sync_service.dart';

enum TrackingStatus { stopped, starting, tracking, error }

/// Orchestrates: GPS stream -> SQLite buffer -> upload. The UI just watches this.
class TrackingController extends ChangeNotifier {
  TrackingController({
    required this.location,
    required this.buffer,
    required this.sync,
  });

  final LocationService location;
  final LocalBuffer buffer;
  final SyncService sync;

  TrackingStatus status = TrackingStatus.stopped;
  String? error;
  int busId = AppConfig.defaultBusId;
  Position? last;
  int pending = 0;
  bool online = true;
  DateTime? lastSyncAt;

  StreamSubscription<Position>? _posSub;
  StreamSubscription<List<ConnectivityResult>>? _connSub;
  Timer? _retryTimer;
  DateTime? _lastStored;

  Future<void> start() async {
    if (status == TrackingStatus.tracking ||
        status == TrackingStatus.starting) {
      return;
    }
    status = TrackingStatus.starting;
    error = null;
    notifyListeners();

    try {
      await location.ensurePermission();
      _posSub = location.positions().listen(_onPosition, onError: _onError);
      // Connectivity regained -> flush immediately instead of waiting for the timer.
      _connSub = Connectivity().onConnectivityChanged.listen((results) {
        if (results.any((r) => r != ConnectivityResult.none)) {
          _flush();
        }
      });
      _retryTimer = Timer.periodic(AppConfig.retryInterval, (_) => _flush());
      status = TrackingStatus.tracking;
      notifyListeners();
      unawaited(_flush()); // upload anything left over from a previous session
    } catch (e) {
      await _teardown();
      error = e.toString();
      status = TrackingStatus.error;
      notifyListeners();
    }
  }

  Future<void> stop() async {
    await _teardown();
    status = TrackingStatus.stopped;
    notifyListeners();
    await _flush(); // one last attempt to upload what's buffered
  }

  Future<void> _onPosition(Position pos) async {
    final now = DateTime.now();
    if (_lastStored != null &&
        now.difference(_lastStored!) < AppConfig.minPingGap) {
      return;
    }
    _lastStored = now;

    last = pos;
    // 1) Always persist first, so nothing is lost if the network or app dies.
    await buffer.add(Ping.fromPosition(busId, pos));
    // 2) Then try to upload (immediately, when online).
    await _flush();
  }

  Future<void> _flush() async {
    online = await sync.flush();
    if (online) lastSyncAt = DateTime.now();
    pending = await buffer.count();
    notifyListeners();
  }

  void _onError(Object e) async {
    await _teardown();
    error = e.toString();
    status = TrackingStatus.error;
    notifyListeners();
  }

  Future<void> _teardown() async {
    await _posSub
        ?.cancel(); // cancelling the stream also stops the foreground service
    _posSub = null;
    await _connSub?.cancel();
    _connSub = null;
    _retryTimer?.cancel();
    _retryTimer = null;
  }

  @override
  void dispose() {
    _posSub?.cancel();
    _connSub?.cancel();
    _retryTimer?.cancel();
    super.dispose();
  }
}
