import 'package:path/path.dart' as p;
import 'package:sqflite/sqflite.dart';

import '../config/config.dart';
import '../models/ping.dart';

/// Durable on-device queue of pings waiting to be uploaded (SQLite).
class LocalBuffer {
  Database? _db;

  Future<Database> get _database async => _db ??= await _open();

  Future<Database> _open() async {
    final path = p.join(await getDatabasesPath(), 'pings.db');
    return openDatabase(
      path,
      version: 1,
      onCreate: (db, _) => db.execute('''
        CREATE TABLE pings (
          id          INTEGER PRIMARY KEY AUTOINCREMENT,
          bus_id      INTEGER NOT NULL,
          latitude    REAL    NOT NULL,
          longitude   REAL    NOT NULL,
          speed       REAL    NOT NULL,
          heading     REAL    NOT NULL,
          recorded_at INTEGER NOT NULL
        )'''),
    );
  }

  Future<void> add(Ping ping) async {
    final db = await _database;
    await db.insert('pings', ping.toRow());
    // Cap storage: delete everything older than the newest N rows.
    await db.rawDelete(
      'DELETE FROM pings WHERE id <= '
      '(SELECT id FROM pings ORDER BY id DESC LIMIT 1 OFFSET ?)',
      [AppConfig.maxBufferedPings],
    );
  }

  Future<List<Ping>> oldest(int limit) async {
    final db = await _database;
    final rows = await db.query(
      'pings',
      orderBy: 'recorded_at ASC',
      limit: limit,
    );
    return rows.map(Ping.fromRow).toList();
  }

  Future<void> deleteIds(List<int> ids) async {
    if (ids.isEmpty) return;
    final db = await _database;
    final marks = List.filled(ids.length, '?').join(',');
    await db.rawDelete('DELETE FROM pings WHERE id IN ($marks)', ids);
  }

  Future<int> count() async {
    final db = await _database;
    return Sqflite.firstIntValue(
          await db.rawQuery('SELECT COUNT(*) FROM pings'),
        ) ??
        0;
  }
}
