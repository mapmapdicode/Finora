import 'package:flutter/material.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/core/theme/finora_tokens.dart';
import 'package:mobile/core/theme/finora_typography.dart';

abstract final class FinoraTheme {
  static final light = ThemeData(
    colorScheme: ColorScheme.fromSeed(
      seedColor: FinoraColors.violet,
      brightness: Brightness.light,
      surface: FinoraColors.surface,
      error: FinoraColors.danger,
    ),
    useMaterial3: true,
    scaffoldBackgroundColor: FinoraColors.background,
    textTheme:
        const TextTheme(
          displayLarge: FinoraTypography.display,
          headlineLarge: FinoraTypography.h1,
          headlineMedium: FinoraTypography.h2,
          headlineSmall: FinoraTypography.h3,
          titleLarge: FinoraTypography.title,
          bodyLarge: FinoraTypography.body,
          bodyMedium: FinoraTypography.bodySmall,
          bodySmall: FinoraTypography.caption,
          labelLarge: FinoraTypography.label,
        ).apply(
          bodyColor: FinoraColors.textPrimary,
          displayColor: FinoraColors.textPrimary,
        ),
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.transparent,
      foregroundColor: FinoraColors.textPrimary,
      elevation: 0,
      centerTitle: false,
    ),
    cardTheme: CardThemeData(
      elevation: 0,
      color: FinoraColors.surface,
      shape: const RoundedRectangleBorder(borderRadius: FinoraRadius.lg),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: const Color(0xfffaf9ff),
      contentPadding: const EdgeInsets.symmetric(
        horizontal: FinoraSpace.md,
        vertical: FinoraSpace.md,
      ),
      border: OutlineInputBorder(
        borderRadius: FinoraRadius.md,
        borderSide: const BorderSide(color: FinoraColors.border),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: FinoraRadius.md,
        borderSide: const BorderSide(color: FinoraColors.violet, width: 1.5),
      ),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: FinoraColors.violet,
        foregroundColor: Colors.white,
        minimumSize: const Size(44, 48),
        shape: const RoundedRectangleBorder(borderRadius: FinoraRadius.md),
        padding: const EdgeInsets.symmetric(
          horizontal: FinoraSpace.lg,
          vertical: FinoraSpace.sm,
        ),
      ),
    ),
    bottomSheetTheme: const BottomSheetThemeData(
      backgroundColor: FinoraColors.surface,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
    ),
    snackBarTheme: SnackBarThemeData(
      behavior: SnackBarBehavior.floating,
      backgroundColor: FinoraColors.textPrimary,
      contentTextStyle: FinoraTypography.bodySmall.copyWith(
        color: Colors.white,
      ),
      shape: const RoundedRectangleBorder(borderRadius: FinoraRadius.md),
    ),
  );

  static final dark = light;
}
