import 'package:flutter/material.dart';
import 'package:mobile/core/theme/finora_colors.dart';

/// Type ramp for calm, scan-friendly wealth management screens.
abstract final class FinoraTypography {
  static const display = TextStyle(
    fontFamily: 'Inter',
    fontSize: 32,
    height: 1.12,
    fontWeight: FontWeight.w800,
    letterSpacing: -1.0,
    color: FinoraColors.textPrimary,
  );
  static const h1 = TextStyle(
    fontFamily: 'Inter',
    fontSize: 24,
    height: 1.2,
    fontWeight: FontWeight.w800,
    letterSpacing: -0.6,
  );
  static const h2 = TextStyle(
    fontFamily: 'Inter',
    fontSize: 20,
    height: 1.25,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.35,
  );
  static const h3 = TextStyle(
    fontFamily: 'Inter',
    fontSize: 18,
    height: 1.3,
    fontWeight: FontWeight.w700,
  );
  static const title = TextStyle(
    fontFamily: 'Inter',
    fontSize: 16,
    height: 1.35,
    fontWeight: FontWeight.w700,
  );
  static const body = TextStyle(
    fontFamily: 'Inter',
    fontSize: 16,
    height: 1.45,
    fontWeight: FontWeight.w500,
  );
  static const bodySmall = TextStyle(
    fontFamily: 'Inter',
    fontSize: 14,
    height: 1.4,
    fontWeight: FontWeight.w500,
  );
  static const caption = TextStyle(
    fontFamily: 'Inter',
    fontSize: 12,
    height: 1.35,
    fontWeight: FontWeight.w500,
  );
  static const label = TextStyle(
    fontFamily: 'Inter',
    fontSize: 12,
    height: 1.2,
    fontWeight: FontWeight.w700,
    letterSpacing: 0.2,
  );

  static const money = TextStyle(
    fontFamily: 'Inter',
    fontSize: 28,
    height: 1.15,
    fontWeight: FontWeight.w800,
    letterSpacing: -0.5,
    fontFeatures: [FontFeature.tabularFigures()],
  );
}
