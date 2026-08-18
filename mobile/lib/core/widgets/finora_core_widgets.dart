import 'package:flutter/material.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/core/theme/finora_tokens.dart';
import 'package:mobile/core/theme/finora_typography.dart';

/// Centers wide-screen content without changing the layout of phone screens.
/// Feature pages keep ownership of their own scrolling and padding; this only
/// prevents cards and forms from stretching across an entire tablet display.
class FinoraContentWidth extends StatelessWidget {
  const FinoraContentWidth({
    super.key,
    required this.child,
    this.maxWidth = 1040,
  });

  final Widget child;
  final double maxWidth;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, constraints) => Align(
      alignment: Alignment.topCenter,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: constraints.maxWidth >= 700 ? maxWidth : double.infinity,
        ),
        child: child,
      ),
    ),
  );
}

/// A quiet, shared backdrop for authenticated Finora screens. The artwork is
/// deliberately subdued so data cards remain the visual focus.
class FinoraAppBackground extends StatelessWidget {
  const FinoraAppBackground({super.key});

  @override
  Widget build(BuildContext context) => Stack(
    fit: StackFit.expand,
    children: [
      Opacity(
        opacity: 0.16,
        child: Image.asset(
          'assets/images/app_bg_maple_light.png',
          fit: BoxFit.cover,
          alignment: Alignment.topCenter,
        ),
      ),
      const DecoratedBox(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xA8FAFAFC), Color(0x66FAFAFC), Color(0xB3FAFAFC)],
            stops: [0, 0.42, 1],
          ),
        ),
      ),
    ],
  );
}

class FinoraCard extends StatelessWidget {
  const FinoraCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(FinoraSpace.md),
    this.onTap,
  });
  final Widget child;
  final EdgeInsetsGeometry padding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) => Material(
    color: FinoraColors.surface,
    borderRadius: FinoraRadius.lg,
    child: InkWell(
      onTap: onTap,
      borderRadius: FinoraRadius.lg,
      child: Container(
        padding: padding,
        decoration: BoxDecoration(
          borderRadius: FinoraRadius.lg,
          border: const Border.fromBorderSide(
            BorderSide(color: FinoraColors.border),
          ),
          boxShadow: FinoraElevation.card,
        ),
        child: child,
      ),
    ),
  );
}

class FinoraMoney extends StatelessWidget {
  const FinoraMoney(
    this.amount, {
    super.key,
    this.currency = 'VND',
    this.style,
    this.color,
    this.compact = false,
  });
  final num amount;
  final String currency;
  final TextStyle? style;
  final Color? color;
  final bool compact;

  String get _value {
    if (compact && amount.abs() >= 1000000) {
      final unit = currency.toUpperCase() == 'VND' ? 'tr' : 'M';
      return '${(amount / 1000000).toStringAsFixed(amount % 1000000 == 0 ? 0 : 1)}$unit';
    }
    final digits = amount.abs().round().toString();
    final grouped = StringBuffer();
    for (var index = 0; index < digits.length; index++) {
      if (index > 0 && (digits.length - index) % 3 == 0) grouped.write('.');
      grouped.write(digits[index]);
    }
    return '${amount < 0 ? '-' : ''}$grouped';
  }

  @override
  Widget build(BuildContext context) => Text(
    '$_value $currency',
    maxLines: 1,
    overflow: TextOverflow.ellipsis,
    style: (style ?? FinoraTypography.money).copyWith(
      color: color ?? Theme.of(context).colorScheme.onSurface,
    ),
  );
}

class FinoraEmptyState extends StatelessWidget {
  const FinoraEmptyState({
    super.key,
    required this.title,
    required this.icon,
    this.message,
    this.action,
  });
  final String title;
  final IconData icon;
  final String? message;
  final Widget? action;

  @override
  Widget build(BuildContext context) => Center(
    child: Padding(
      padding: const EdgeInsets.all(FinoraSpace.xl),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: const BoxDecoration(
              color: FinoraColors.primarySoft,
              shape: BoxShape.circle,
            ),
            child: Icon(icon, color: FinoraColors.primary),
          ),
          const SizedBox(height: FinoraSpace.md),
          Text(
            title,
            style: FinoraTypography.h3.copyWith(
              color: FinoraColors.textPrimary,
            ),
            textAlign: TextAlign.center,
          ),
          if (message != null) ...[
            const SizedBox(height: FinoraSpace.xs),
            Text(
              message!,
              style: FinoraTypography.bodySmall.copyWith(
                color: FinoraColors.textSecondary,
              ),
              textAlign: TextAlign.center,
            ),
          ],
          if (action != null) ...[
            const SizedBox(height: FinoraSpace.lg),
            action!,
          ],
        ],
      ),
    ),
  );
}
