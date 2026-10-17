import 'dart:io';

import 'package:archive/archive_io.dart';
import 'package:bett_box/common/flclash_backup.dart';
import 'package:sqlite3/sqlite3.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'database metadata preserves URL labels order selections and dates',
    () async {
      final temp = await Directory.systemTemp.createTemp('backup-test-');
      try {
        final path = '${temp.path}/metadata.sqlite';
        final db = sqlite3.open(path);
        db.execute(
          'CREATE TABLE profiles (id INTEGER PRIMARY KEY, label TEXT, url TEXT, "order" INTEGER, selected_map TEXT, unfold_set TEXT, last_update_date INTEGER, auto_update_duration_millis INTEGER, auto_update INTEGER)',
        );
        db.execute('INSERT INTO profiles VALUES (?,?,?,?,?,?,?,?,?)', [
          320739999513317376,
          'Original subscription 香港',
          'https://example.invalid/sub?token=test-only',
          0,
          '{"group":"node"}',
          '["group"]',
          1792221601,
          86400000,
          1,
        ]);
        db.execute('INSERT INTO profiles VALUES (?,?,?,?,?,?,?,?,?)', [
          2,
          'Original local name',
          '',
          1,
          '{}',
          '[]',
          null,
          86400000,
          0,
        ]);
        db.dispose();
        final bytes = await File(path).readAsBytes();
        final archive = Archive()
          ..addFile(ArchiveFile('database.sqlite', bytes.length, bytes));
        final result = (await readFlClashBackupProfiles(archive))!;
        expect(result.map((p) => p.label), [
          'Original subscription 香港',
          'Original local name',
        ]);
        expect(result.first.id, '320739999513317376');
        expect(result.first.url, 'https://example.invalid/sub?token=test-only');
        expect(result.first.selectedMap, {'group': 'node'});
        expect(
          result.first.lastUpdateDate!.millisecondsSinceEpoch,
          1792221601000,
        );
        expect(result.last.url, '');
        expect(result.last.autoUpdate, false);
      } finally {
        await temp.delete(recursive: true);
      }
    },
  );
}
