import 'package:flutter/gestures.dart';
import 'package:material_ui/material_ui.dart';

class ProfilesTitle extends StatelessWidget {
  const ProfilesTitle({
    super.key,
    required this.title,
    required this.onTap,
    required this.onLongPress,
  });

  final String title;
  final VoidCallback onTap;
  final VoidCallback onLongPress;

  @override
  Widget build(BuildContext context) => Align(
    alignment: Alignment.centerLeft,
    child: RawGestureDetector(
      behavior: HitTestBehavior.opaque,
      gestures: {
        TapGestureRecognizer:
            GestureRecognizerFactoryWithHandlers<TapGestureRecognizer>(
              TapGestureRecognizer.new,
              (recognizer) => recognizer.onTap = onTap,
            ),
        LongPressGestureRecognizer:
            GestureRecognizerFactoryWithHandlers<LongPressGestureRecognizer>(
              () => LongPressGestureRecognizer(
                duration: const Duration(milliseconds: 300),
              ),
              (recognizer) => recognizer.onLongPress = onLongPress,
            ),
      },
      child: Text(title),
    ),
  );
}
