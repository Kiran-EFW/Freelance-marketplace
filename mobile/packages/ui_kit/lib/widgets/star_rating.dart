import 'package:flutter/material.dart';
import '../theme/colors.dart';

/// Displays a star rating.
///
/// Supports two modes:
/// - **Display mode** (default): Shows a read-only star rating.
/// - **Input mode**: Allows tapping to select a rating via [onChanged].
class StarRating extends StatelessWidget {
  final double rating;
  final int maxStars;
  final double size;
  final ValueChanged<int>? onChanged;
  final Color? filledColor;
  final Color? emptyColor;

  const StarRating({
    super.key,
    required this.rating,
    this.maxStars = 5,
    this.size = 20,
    this.onChanged,
    this.filledColor,
    this.emptyColor,
  });

  bool get _isInteractive => onChanged != null;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: List.generate(maxStars, (index) {
        final starValue = index + 1;
        final fillAmount = (rating - index).clamp(0.0, 1.0);

        Widget star;

        if (fillAmount >= 1.0) {
          star = Icon(
            Icons.star_rounded,
            size: size,
            color: filledColor ?? SevaColors.starFilled,
          );
        } else if (fillAmount > 0) {
          star = Stack(
            children: [
              Icon(
                Icons.star_rounded,
                size: size,
                color: emptyColor ?? SevaColors.starEmpty,
              ),
              ClipRect(
                clipper: _StarClipper(fillAmount),
                child: Icon(
                  Icons.star_rounded,
                  size: size,
                  color: filledColor ?? SevaColors.starFilled,
                ),
              ),
            ],
          );
        } else {
          star = Icon(
            Icons.star_rounded,
            size: size,
            color: emptyColor ?? SevaColors.starEmpty,
          );
        }

        if (_isInteractive) {
          return GestureDetector(
            onTap: () => onChanged!(starValue),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 2),
              child: star,
            ),
          );
        }

        return star;
      }),
    );
  }
}

class _StarClipper extends CustomClipper<Rect> {
  final double fillPercentage;

  _StarClipper(this.fillPercentage);

  @override
  Rect getClip(Size size) {
    return Rect.fromLTWH(0, 0, size.width * fillPercentage, size.height);
  }

  @override
  bool shouldReclip(covariant _StarClipper oldClipper) {
    return oldClipper.fillPercentage != fillPercentage;
  }
}

/// A larger, interactive rating input with labels.
class RatingInput extends StatelessWidget {
  final int rating;
  final ValueChanged<int> onChanged;
  final double starSize;
  final bool showLabel;

  const RatingInput({
    super.key,
    required this.rating,
    required this.onChanged,
    this.starSize = 40,
    this.showLabel = true,
  });

  String get _label {
    switch (rating) {
      case 1:
        return 'Poor';
      case 2:
        return 'Below Average';
      case 3:
        return 'Average';
      case 4:
        return 'Good';
      case 5:
        return 'Excellent';
      default:
        return 'Tap to rate';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        StarRating(
          rating: rating.toDouble(),
          size: starSize,
          onChanged: onChanged,
        ),
        if (showLabel) ...[
          const SizedBox(height: 8),
          Text(
            _label,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: rating > 0
                      ? SevaColors.textPrimary
                      : SevaColors.textTertiary,
                  fontWeight: FontWeight.w500,
                ),
          ),
        ],
      ],
    );
  }
}
