import 'dart:convert';
import 'dart:io';

import 'package:archive/archive_io.dart';
import 'package:bett_box/models/models.dart';
import 'package:sqlite3/sqlite3.dart';

/// Reads metadata from a private temporary copy; never opens the live database.
Future<List<Profile>?> readFlClashBackupProfiles(Archive archive) async {
  final files = archive.files.where((f) => f.name == 'database.sqlite');
  if (files.isEmpty) return null;
  final temp = await Directory.systemTemp.createTemp('bettbox-flclash-');
  try {
    final file = File('${temp.path}/database.sqlite');
    await file.writeAsBytes(files.single.content, flush: true);
    final db = sqlite3.open(file.path, mode: OpenMode.readOnly);
    try {
      final columns = db
          .select('PRAGMA table_info(profiles)')
          .map((r) => r['name'])
          .toSet();
      if (!columns.containsAll(['id', 'label', 'url'])) {
        throw const FormatException('Unsupported FlClash profiles schema');
      }
      final rows = db.select('SELECT * FROM profiles ORDER BY "order", id');
      dynamic decoded(dynamic value, dynamic fallback) =>
          value == null ? fallback : jsonDecode(value as String);
      return rows.map((row) {
        final r = Map<String, dynamic>.from(row);
        final stamp = r['last_update_date'] as int?;
        return Profile(
          id: r['id'].toString(),
          label: r['label'] as String?,
          url: r['url'] as String? ?? '',
          currentGroupName: r['current_group_name'] as String?,
          ageSecretKey: r['age_secret_key'] as String?,
          lastUpdateDate: stamp == null
              ? null
              : DateTime.fromMillisecondsSinceEpoch(
                  stamp.abs() < 100000000000 ? stamp * 1000 : stamp,
                ),
          autoUpdateDuration: Duration(
            milliseconds:
                (r['auto_update_duration_millis'] as int?) ?? 86400000,
          ),
          autoUpdate: r['auto_update'] != 0,
          subscriptionInfo: r['subscription_info'] == null
              ? null
              : SubscriptionInfo.fromJson(
                  Map<String, dynamic>.from(
                    decoded(r['subscription_info'], {}),
                  ),
                ),
          selectedMap: Map<String, String>.from(decoded(r['selected_map'], {})),
          unfoldSet: Set<String>.from(decoded(r['unfold_set'], [])),
        );
      }).toList();
    } finally {
      db.dispose();
    }
  } finally {
    await temp.delete(recursive: true);
  }
}
