import 'package:flutter/material.dart';
import '../theme/colors.dart';

/// A colored badge indicating the status of a job or other entity.
///
/// Maps status strings to brand-consistent colors and labels.
class StatusBadge extends StatelessWidget {
  final String status;
  final bool isCompact;

  const StatusBadge({
    super.key,
    required this.status,
    this.isCompact = false,
  });

  Color get _backgroundColor {
    switch (status.toLowerCase()) {
      case 'draft':
        return SevaColors.neutral200;
      case 'posted':
        return SevaColors.infoLight;
      case 'matched':
        return const Color(0xFFEDE9FE); // purple-100
      case 'accepted':
        return SevaColors.secondaryFaded;
      case 'in_progress':
        return SevaColors.primaryFaded;
      case 'completed':
        return SevaColors.successLight;
      case 'cancelled':
        return SevaColors.errorLight;
      case 'disputed':
        return SevaColors.warningLight;
      case 'verified':
        return SevaColors.successLight;
      case 'pending':
        return SevaColors.warningLight;
      case 'rejected':
        return SevaColors.errorLight;
      case 'processing':
        return SevaColors.infoLight;
      case 'paid':
        return SevaColors.successLight;
      default:
        return SevaColors.neutral200;
    }
  }

  Color get _textColor {
    switch (status.toLowerCase()) {
      case 'draft':
        return SevaColors.statusDraft;
      case 'posted':
        return SevaColors.statusPosted;
      case 'matched':
        return SevaColors.statusMatched;
      case 'accepted':
        return SevaColors.statusAccepted;
      case 'in_progress':
        return SevaColors.statusInProgress;
      case 'completed':
        return SevaColors.statusCompleted;
      case 'cancelled':
        return SevaColors.statusCancelled;
      case 'disputed':
        return SevaColors.statusDisputed;
      case 'verified':
        return SevaColors.success;
      case 'pending':
        return SevaColors.warning;
      case 'rejected':
        return SevaColors.error;
      case 'processing':
        return SevaColors.info;
      case 'paid':
        return SevaColors.success;
      default:
        return SevaColors.neutral500;
    }
  }

  String get _displayLabel {
    switch (status.toLowerCase()) {
      case 'in_progress':
        return 'In Progress';
      default:
        return status[0].toUpperCase() + status.substring(1).toLowerCase();
    }
  }

  IconData? get _icon {
    switch (status.toLowerCase()) {
      case 'completed':
        return Icons.check_circle_outline;
      case 'cancelled':
        return Icons.cancel_outlined;
      case 'in_progress':
        return Icons.autorenew;
      case 'verified':
        return Icons.verified_outlined;
      case 'disputed':
        return Icons.warning_amber_outlined;
      default:
        return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: isCompact ? 6 : 10,
        vertical: isCompact ? 2 : 4,
      ),
      decoration: BoxDecoration(
        color: _backgroundColor,
        borderRadius: BorderRadius.circular(isCompact ? 4 : 6),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_icon != null && !isCompact) ...[
            Icon(_icon, size: 12, color: _textColor),
            const SizedBox(width: 4),
          ],
          Text(
            _displayLabel,
            style: TextStyle(
              fontSize: isCompact ? 10 : 12,
              fontWeight: FontWeight.w600,
              color: _textColor,
            ),
          ),
        ],
      ),
    );
  }
}
