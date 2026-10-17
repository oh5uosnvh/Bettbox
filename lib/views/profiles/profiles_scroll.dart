import 'dart:async';

import 'package:material_ui/material_ui.dart';

void scrollProfilesToBottom(ScrollController controller) {
  if (!controller.hasClients || !controller.position.hasContentDimensions)
    return;
  unawaited(
    controller.animateTo(
      controller.position.maxScrollExtent,
      duration: const Duration(milliseconds: 500),
      curve: Curves.easeOutCubic,
    ),
  );
}
