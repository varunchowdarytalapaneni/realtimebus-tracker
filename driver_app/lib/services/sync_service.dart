import '../config/config.dart';
import '../services/api_client.dart';
import '../services/local_buffer.dart';

/// Drains the local SQLite buffer to the backend, oldest first, in batches.
class SyncService {
  SyncService(this._buffer, this._api);
  final LocalBuffer _buffer;
  final ApiClient _api;

  bool _flushing = false;
  bool _lastOk = true;

  /// Returns true if the buffer was fully drained (or last attempt succeeded),
  /// false if the server could not be reached.
  Future<bool> flush() async {
    if (_flushing) return _lastOk; // one flush at a time
    _flushing = true;
    try {
      while (true) {
        final batch = await _buffer.oldest(AppConfig.batchSize);
        if (batch.isEmpty) return _lastOk = true;

        final result = await _api.sendBatch(batch);
        if (result == SendResult.retry) return _lastOk = false;

        // ok -> delivered. rejected -> bad data, drop it so it can't block the queue.
        await _buffer.deleteIds(batch.map((p) => p.id!).toList());
      }
    } finally {
      _flushing = false;
    }
  }
}
