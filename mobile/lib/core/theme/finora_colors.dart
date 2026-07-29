import 'package:flutter/material.dart';

/// The Finora light design language: calm white surfaces, precise purple
/// actions, and a very small warm autumn accent. Legacy names are retained so
/// screens can share the new visual system without changing their behaviour.
abstract final class FinoraColors {
  static const midnight = Color(0xff232323);
  static const navy = Color(0xff3b326e);
  static const deepPurple = Color(0xff5142d7);
  static const violet = Color(0xff6d5df6);
  static const purple = Color(0xff8b7bff);
  static const aqua = Color(0xff5f95ea);
  static const goldLight = Color(0xfffff5ec);
  static const gold = Color(0xffb36a4c);
  static const goldDark = Color(0xff8d4c34);
  static const mist = Color(0xfffafafc);
  static const ink = Color(0xff232323);

  static const primaryDeep = deepPurple;
  static const primary = violet;
  static const primarySoft = Color(0xfff3f0ff);
  static const accentGold = gold;
  static const accentAmber = Color(0xfff5a623);

  static const success = Color(0xff36c275);
  static const warning = Color(0xfff5a623);
  static const danger = Color(0xfff04438);
  static const info = Color(0xff5f95ea);

  static const background = mist;
  static const surface = Colors.white;
  static const surfaceElevated = Color(0xffffffff);
  static const surfaceGlass = Color(0xfaffffff);
  static const textPrimary = ink;
  static const textSecondary = Color(0xff707070);
  static const textTertiary = Color(0xff9b9b9b);
  static const border = Color(0xffeceafb);
  static const borderStrong = Color(0xffdcd8fa);

  static const goldGradient = LinearGradient(
    colors: [
      Color(0xffa99eff),
      Color(0xff8b7bff),
      Color(0xff6d5df6),
      Color(0xff5142d7),
    ],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
  static const bgGradient = LinearGradient(
    colors: [Color(0x00ffffff), Color(0x22f3f0ff), Color(0x44eceafb)],
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
  );
}
