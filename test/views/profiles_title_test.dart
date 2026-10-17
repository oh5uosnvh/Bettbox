import 'package:bett_box/views/profiles/profiles_title.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('title triggers at 300ms and does not also tap', (tester) async {
    var taps = 0;
    var holds = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          appBar: AppBar(
            title: ProfilesTitle(
              title: '配置',
              onTap: () => taps++,
              onLongPress: () => holds++,
            ),
          ),
        ),
      ),
    );
    final gesture = await tester.startGesture(
      tester.getCenter(find.text('配置')),
    );
    await tester.pump(const Duration(milliseconds: 299));
    expect(holds, 0);
    await tester.pump(const Duration(milliseconds: 1));
    expect(holds, 1);
    await gesture.up();
    await tester.pump();
    expect(taps, 0);
    await tester.tap(find.text('配置'));
    expect(taps, 1);
    expect(find.byType(PopupMenuButton<String>), findsNothing);
  });
}
