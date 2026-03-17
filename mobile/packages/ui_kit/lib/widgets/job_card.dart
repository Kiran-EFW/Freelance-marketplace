import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../theme/colors.dart';
import 'status_badge.dart';

/// Displays a job summary in a card format.
///
/// Shows the job title, status badge, category, date, price,
/// and relevant contextual information.
class JobCard extends StatelessWidget {
  final String title;
  final String status;
  final String? categoryName;
  final DateTime? scheduledAt;
  final DateTime createdAt;
  final String? priceDisplay;
  final String? customerName;
  final String? providerName;
  final String? address;
  final String? urgency;
  final VoidCallback? onTap;

  const JobCard({
    super.key,
    required this.title,
    required this.status,
    this.categoryName,
    this.scheduledAt,
    required this.createdAt,
    this.priceDisplay,
    this.customerName,
    this.providerName,
    this.address,
    this.urgency,
    this.onTap,
  });

  String get _dateDisplay {
    final date = scheduledAt ?? createdAt;
    final now = DateTime.now();
    final diff = date.difference(now);

    if (diff.inDays == 0) return 'Today';
    if (diff.inDays == 1) return 'Tomorrow';
    if (diff.inDays == -1) return 'Yesterday';
    if (diff.inDays.abs() < 7) {
      return DateFormat('EEEE').format(date);
    }
    return DateFormat('d MMM yyyy').format(date);
  }

  String get _timeDisplay {
    final date = scheduledAt ?? createdAt;
    return DateFormat('h:mm a').format(date);
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Theme.of(context).cardColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: Theme.of(context).brightness == Brightness.dark
                ? SevaColors.neutral700
                : SevaColors.neutral200,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header row: title + status badge
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Text(
                    title,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(width: 8),
                StatusBadge(status: status),
              ],
            ),
            const SizedBox(height: 8),

            // Category
            if (categoryName != null) ...[
              Row(
                children: [
                  const Icon(
                    Icons.category_outlined,
                    size: 14,
                    color: SevaColors.textTertiary,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    categoryName!,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: SevaColors.textSecondary,
                        ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
            ],

            // Date and time
            Row(
              children: [
                const Icon(
                  Icons.calendar_today_outlined,
                  size: 14,
                  color: SevaColors.textTertiary,
                ),
                const SizedBox(width: 4),
                Text(
                  '$_dateDisplay at $_timeDisplay',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: SevaColors.textSecondary,
                      ),
                ),
                if (urgency != null && urgency != 'normal') ...[
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: urgency == 'emergency'
                          ? SevaColors.errorLight
                          : urgency == 'high'
                              ? SevaColors.warningLight
                              : Colors.transparent,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      urgency!.toUpperCase(),
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: urgency == 'emergency'
                            ? SevaColors.error
                            : urgency == 'high'
                                ? SevaColors.warning
                                : SevaColors.textTertiary,
                      ),
                    ),
                  ),
                ],
              ],
            ),
            const SizedBox(height: 4),

            // Location
            if (address != null) ...[
              Row(
                children: [
                  const Icon(
                    Icons.location_on_outlined,
                    size: 14,
                    color: SevaColors.textTertiary,
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      address!,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: SevaColors.textTertiary,
                          ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
            ],

            const Divider(height: 16),

            // Footer: price + person
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                if (priceDisplay != null)
                  Text(
                    priceDisplay!,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          color: SevaColors.primary,
                          fontWeight: FontWeight.w700,
                        ),
                  ),
                if (customerName != null)
                  Row(
                    children: [
                      const Icon(
                        Icons.person_outline,
                        size: 14,
                        color: SevaColors.textTertiary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        customerName!,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: SevaColors.textSecondary,
                            ),
                      ),
                    ],
                  ),
                if (providerName != null)
                  Row(
                    children: [
                      const Icon(
                        Icons.handyman_outlined,
                        size: 14,
                        color: SevaColors.textTertiary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        providerName!,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: SevaColors.textSecondary,
                            ),
                      ),
                    ],
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
