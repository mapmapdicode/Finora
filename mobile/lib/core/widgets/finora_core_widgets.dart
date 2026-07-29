import 'package:flutter/material.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/core/theme/finora_tokens.dart';
import 'package:mobile/core/theme/finora_typography.dart';

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
      return '${(amount / 1000000).toStringAsFixed(amount % 1000000 == 0 ? 0 : 1)}M';
    }
    final digits = amount.abs().round().toString();
    final grouped = StringBuffer();
    for (var index = 0; index < digits.length; index++) {
      if (index > 0 && (digits.length - index) % 3 == 0) grouped.write(',');
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
