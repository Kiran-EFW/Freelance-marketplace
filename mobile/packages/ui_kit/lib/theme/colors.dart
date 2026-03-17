import 'package:flutter/material.dart';

/// Seva brand color palette.
///
/// All colors are derived from the saffron/orange primary and designed
/// for WCAG AA contrast compliance on both light and dark surfaces.
class SevaColors {
  SevaColors._();

  // ---------------------------------------------------------------------------
  // Primary – Saffron / Orange
  // ---------------------------------------------------------------------------

  static const Color primary = Color(0xFFF97316);
  static const Color primaryLight = Color(0xFFFB923C);
  static const Color primaryDark = Color(0xFFEA580C);
  static const Color primaryFaded = Color(0xFFFFF7ED);

  static const MaterialColor primarySwatch = MaterialColor(
    0xFFF97316,
    <int, Color>{
      50: Color(0xFFFFF7ED),
      100: Color(0xFFFFEDD5),
      200: Color(0xFFFED7AA),
      300: Color(0xFFFDBA74),
      400: Color(0xFFFB923C),
      500: Color(0xFFF97316),
      600: Color(0xFFEA580C),
      700: Color(0xFFC2410C),
      800: Color(0xFF9A3412),
      900: Color(0xFF7C2D12),
    },
  );

  // ---------------------------------------------------------------------------
  // Secondary – Teal
  // ---------------------------------------------------------------------------

  static const Color secondary = Color(0xFF14B8A6);
  static const Color secondaryLight = Color(0xFF2DD4BF);
  static const Color secondaryDark = Color(0xFF0D9488);
  static const Color secondaryFaded = Color(0xFFF0FDFA);

  // ---------------------------------------------------------------------------
  // Semantic colors
  // ---------------------------------------------------------------------------

  static const Color success = Color(0xFF22C55E);
  static const Color successLight = Color(0xFFDCFCE7);
  static const Color warning = Color(0xFFEAB308);
  static const Color warningLight = Color(0xFFFEF9C3);
  static const Color error = Color(0xFFEF4444);
  static const Color errorLight = Color(0xFFFEE2E2);
  static const Color info = Color(0xFF3B82F6);
  static const Color infoLight = Color(0xFFDBEAFE);

  // ---------------------------------------------------------------------------
  // Neutrals
  // ---------------------------------------------------------------------------

  static const Color neutral50 = Color(0xFFFAFAFA);
  static const Color neutral100 = Color(0xFFF5F5F5);
  static const Color neutral200 = Color(0xFFE5E5E5);
  static const Color neutral300 = Color(0xFFD4D4D4);
  static const Color neutral400 = Color(0xFFA3A3A3);
  static const Color neutral500 = Color(0xFF737373);
  static const Color neutral600 = Color(0xFF525252);
  static const Color neutral700 = Color(0xFF404040);
  static const Color neutral800 = Color(0xFF262626);
  static const Color neutral900 = Color(0xFF171717);

  // ---------------------------------------------------------------------------
  // Surface colors
  // ---------------------------------------------------------------------------

  static const Color backgroundLight = Color(0xFFFAFAFA);
  static const Color backgroundDark = Color(0xFF121212);
  static const Color surfaceLight = Color(0xFFFFFFFF);
  static const Color surfaceDark = Color(0xFF1E1E1E);
  static const Color cardLight = Color(0xFFFFFFFF);
  static const Color cardDark = Color(0xFF2A2A2A);

  // ---------------------------------------------------------------------------
  // Text colors
  // ---------------------------------------------------------------------------

  static const Color textPrimary = Color(0xFF171717);
  static const Color textSecondary = Color(0xFF525252);
  static const Color textTertiary = Color(0xFFA3A3A3);
  static const Color textOnPrimary = Color(0xFFFFFFFF);
  static const Color textPrimaryDark = Color(0xFFFAFAFA);
  static const Color textSecondaryDark = Color(0xFFA3A3A3);

  // ---------------------------------------------------------------------------
  // Job status colors
  // ---------------------------------------------------------------------------

  static const Color statusDraft = Color(0xFF737373);
  static const Color statusPosted = Color(0xFF3B82F6);
  static const Color statusMatched = Color(0xFF8B5CF6);
  static const Color statusAccepted = Color(0xFF14B8A6);
  static const Color statusInProgress = Color(0xFFF97316);
  static const Color statusCompleted = Color(0xFF22C55E);
  static const Color statusCancelled = Color(0xFFEF4444);
  static const Color statusDisputed = Color(0xFFEAB308);

  // ---------------------------------------------------------------------------
  // Trust score colors
  // ---------------------------------------------------------------------------

  static const Color trustExcellent = Color(0xFF22C55E);
  static const Color trustVeryGood = Color(0xFF84CC16);
  static const Color trustGood = Color(0xFFEAB308);
  static const Color trustFair = Color(0xFFF97316);
  static const Color trustNew = Color(0xFFA3A3A3);

  /// Get the color for a given trust score (0-100).
  static Color trustColor(double score) {
    if (score >= 90) return trustExcellent;
    if (score >= 75) return trustVeryGood;
    if (score >= 60) return trustGood;
    if (score >= 40) return trustFair;
    return trustNew;
  }

  // ---------------------------------------------------------------------------
  // Star rating
  // ---------------------------------------------------------------------------

  static const Color starFilled = Color(0xFFFBBF24);
  static const Color starEmpty = Color(0xFFD4D4D4);
}
