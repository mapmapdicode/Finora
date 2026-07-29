import 'package:flutter/material.dart';

/// Layout, shape, and elevation tokens shared by all Finora mobile features.
abstract final class FinoraSpace {
  static const double xxs = 4;
  static const double xs = 8;
  static const double sm = 12;
  static const double md = 16;
  static const double lg = 20;
  static const double xl = 24;
  static const double xxl = 32;
}

abstract final class FinoraRadius {
  static const BorderRadius xs = BorderRadius.all(Radius.circular(8));
  static const BorderRadius sm = BorderRadius.all(Radius.circular(12));
  static const BorderRadius md = BorderRadius.all(Radius.circular(16));
  static const BorderRadius lg = BorderRadius.all(Radius.circular(16));
  static const BorderRadius xl = BorderRadius.all(Radius.circular(24));
  static const BorderRadius full = BorderRadius.all(Radius.circular(999));
}

abstract final class FinoraElevation {
  static const List<BoxShadow> card = [
    BoxShadow(color: Color(0x0D6D5DF6), blurRadius: 18, offset: Offset(0, 5)),
  ];
  static const List<BoxShadow> floating = [
    BoxShadow(color: Color(0x286D5DF6), blurRadius: 28, offset: Offset(0, 10)),
  ];
}
