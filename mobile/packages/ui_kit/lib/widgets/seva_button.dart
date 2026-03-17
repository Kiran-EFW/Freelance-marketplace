import 'package:flutter/material.dart';
import '../theme/colors.dart';

/// Visual variant for [SevaButton].
enum SevaButtonVariant { primary, secondary, outline, text, destructive }

/// Size preset for [SevaButton].
enum SevaButtonSize { small, medium, large }

/// A branded button with multiple variants, sizes, and loading state.
class SevaButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;
  final SevaButtonVariant variant;
  final SevaButtonSize size;
  final bool isLoading;
  final bool isFullWidth;
  final IconData? icon;
  final IconData? trailingIcon;

  const SevaButton({
    super.key,
    required this.label,
    this.onPressed,
    this.variant = SevaButtonVariant.primary,
    this.size = SevaButtonSize.medium,
    this.isLoading = false,
    this.isFullWidth = true,
    this.icon,
    this.trailingIcon,
  });

  double get _height {
    switch (size) {
      case SevaButtonSize.small:
        return 36;
      case SevaButtonSize.medium:
        return 48;
      case SevaButtonSize.large:
        return 56;
    }
  }

  double get _fontSize {
    switch (size) {
      case SevaButtonSize.small:
        return 12;
      case SevaButtonSize.medium:
        return 14;
      case SevaButtonSize.large:
        return 16;
    }
  }

  EdgeInsets get _padding {
    switch (size) {
      case SevaButtonSize.small:
        return const EdgeInsets.symmetric(horizontal: 12);
      case SevaButtonSize.medium:
        return const EdgeInsets.symmetric(horizontal: 20);
      case SevaButtonSize.large:
        return const EdgeInsets.symmetric(horizontal: 28);
    }
  }

  @override
  Widget build(BuildContext context) {
    final effectiveOnPressed = isLoading ? null : onPressed;

    Widget child;
    if (isLoading) {
      child = SizedBox(
        width: 20,
        height: 20,
        child: CircularProgressIndicator(
          strokeWidth: 2,
          valueColor: AlwaysStoppedAnimation<Color>(
            variant == SevaButtonVariant.primary ||
                    variant == SevaButtonVariant.destructive
                ? Colors.white
                : SevaColors.primary,
          ),
        ),
      );
    } else {
      final textWidget = Text(
        label,
        style: TextStyle(
          fontSize: _fontSize,
          fontWeight: FontWeight.w600,
        ),
      );

      if (icon != null || trailingIcon != null) {
        child = Row(
          mainAxisSize: isFullWidth ? MainAxisSize.max : MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (icon != null) ...[
              Icon(icon, size: _fontSize + 4),
              const SizedBox(width: 8),
            ],
            textWidget,
            if (trailingIcon != null) ...[
              const SizedBox(width: 8),
              Icon(trailingIcon, size: _fontSize + 4),
            ],
          ],
        );
      } else {
        child = textWidget;
      }
    }

    final buttonStyle = ButtonStyle(
      minimumSize: WidgetStatePropertyAll(
        Size(isFullWidth ? double.infinity : 0, _height),
      ),
      padding: WidgetStatePropertyAll(_padding),
      shape: WidgetStatePropertyAll(
        RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
      elevation: const WidgetStatePropertyAll(0),
    );

    switch (variant) {
      case SevaButtonVariant.primary:
        return ElevatedButton(
          onPressed: effectiveOnPressed,
          style: buttonStyle.copyWith(
            backgroundColor: const WidgetStatePropertyAll(SevaColors.primary),
            foregroundColor:
                const WidgetStatePropertyAll(SevaColors.textOnPrimary),
          ),
          child: child,
        );

      case SevaButtonVariant.secondary:
        return ElevatedButton(
          onPressed: effectiveOnPressed,
          style: buttonStyle.copyWith(
            backgroundColor:
                const WidgetStatePropertyAll(SevaColors.primaryFaded),
            foregroundColor:
                const WidgetStatePropertyAll(SevaColors.primaryDark),
          ),
          child: child,
        );

      case SevaButtonVariant.outline:
        return OutlinedButton(
          onPressed: effectiveOnPressed,
          style: buttonStyle.copyWith(
            foregroundColor: const WidgetStatePropertyAll(SevaColors.primary),
            side: const WidgetStatePropertyAll(
              BorderSide(color: SevaColors.primary),
            ),
          ),
          child: child,
        );

      case SevaButtonVariant.text:
        return TextButton(
          onPressed: effectiveOnPressed,
          style: buttonStyle.copyWith(
            foregroundColor: const WidgetStatePropertyAll(SevaColors.primary),
          ),
          child: child,
        );

      case SevaButtonVariant.destructive:
        return ElevatedButton(
          onPressed: effectiveOnPressed,
          style: buttonStyle.copyWith(
            backgroundColor: const WidgetStatePropertyAll(SevaColors.error),
            foregroundColor: const WidgetStatePropertyAll(Colors.white),
          ),
          child: child,
        );
    }
  }
}
