import 'package:flutter/material.dart';

/// Finora's own digital-banking palette: midnight, violet, aqua, and gold.
abstract final class FinoraColors {
  static const midnight = Color(0xff140523);
  static const navy = Color(0xff270a3f);
  static const deepPurple = Color(0xff3d155f);
  static const violet = Color(0xff6d28d9);
  static const purple = Color(0xff8b5cf6);
  static const aqua = Color(0xff22d3ee);
  static const goldLight = Color(0xfff7d070);
  static const gold = Color(0xffdfac40);
  static const goldDark = Color(0xffb88220);
  static const mist = Color(0xfff4f6ff);
  static const ink = Color(0xff13203a);

  static const goldGradient = LinearGradient(
    colors: [
      Color(0xfffef08a),
      Color(0xfffbbf24),
      Color(0xffd97706),
      Color(0xffb45309),
    ],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
  static const bgGradient = LinearGradient(
    colors: [
      Color(0x220f172a),
      Color(0x331a052e),
      Color(0x55000000),
    ],
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
  );
}
