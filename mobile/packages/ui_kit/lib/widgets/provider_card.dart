import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../theme/colors.dart';
import 'star_rating.dart';
import 'status_badge.dart';

/// Displays a service provider summary in a card format.
///
/// Shows the provider's avatar, name, rating, distance, skills,
/// trust score, and verification status.
class ProviderCard extends StatelessWidget {
  final String name;
  final String? avatarUrl;
  final double rating;
  final int reviewCount;
  final double? distanceKm;
  final List<String> skills;
  final double trustScore;
  final bool isVerified;
  final String? hourlyRate;
  final String? currency;
  final int responseTimeMinutes;
  final VoidCallback? onTap;

  const ProviderCard({
    super.key,
    required this.name,
    this.avatarUrl,
    required this.rating,
    this.reviewCount = 0,
    this.distanceKm,
    this.skills = const [],
    this.trustScore = 0,
    this.isVerified = false,
    this.hourlyRate,
    this.currency,
    this.responseTimeMinutes = 0,
    this.onTap,
  });

  String get _distanceDisplay {
    if (distanceKm == null) return '';
    if (distanceKm! < 1) return '${(distanceKm! * 1000).round()} m';
    return '${distanceKm!.toStringAsFixed(1)} km';
  }

  String get _responseTimeDisplay {
    if (responseTimeMinutes <= 0) return '';
    if (responseTimeMinutes < 60) return '~${responseTimeMinutes}min';
    return '~${responseTimeMinutes ~/ 60}hr';
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
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Avatar
            Stack(
              children: [
                CircleAvatar(
                  radius: 28,
                  backgroundColor: SevaColors.primaryFaded,
                  backgroundImage: avatarUrl != null
                      ? CachedNetworkImageProvider(avatarUrl!)
                      : null,
                  child: avatarUrl == null
                      ? Text(
                          name.isNotEmpty ? name[0].toUpperCase() : '?',
                          style: const TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.w700,
                            color: SevaColors.primary,
                          ),
                        )
                      : null,
                ),
                if (isVerified)
                  Positioned(
                    bottom: 0,
                    right: 0,
                    child: Container(
                      padding: const EdgeInsets.all(2),
                      decoration: const BoxDecoration(
                        color: Colors.white,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.verified,
                        size: 16,
                        color: SevaColors.info,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(width: 12),

            // Details
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Name and trust score
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          name,
                          style:
                              Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.w600,
                                  ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (trustScore > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: SevaColors.trustColor(trustScore)
                                .withValues(alpha: 0.15),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            '${trustScore.round()}',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: SevaColors.trustColor(trustScore),
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 4),

                  // Rating
                  Row(
                    children: [
                      StarRating(rating: rating, size: 14),
                      const SizedBox(width: 4),
                      Text(
                        rating.toStringAsFixed(1),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                      Text(
                        ' ($reviewCount)',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: SevaColors.textTertiary,
                            ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),

                  // Skills chips
                  if (skills.isNotEmpty)
                    Wrap(
                      spacing: 4,
                      runSpacing: 4,
                      children: skills.take(3).map((skill) {
                        return Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 3,
                          ),
                          decoration: BoxDecoration(
                            color: SevaColors.primaryFaded,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            skill,
                            style: const TextStyle(
                              fontSize: 11,
                              color: SevaColors.primaryDark,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                  const SizedBox(height: 8),

                  // Meta row
                  Row(
                    children: [
                      if (distanceKm != null) ...[
                        Icon(
                          Icons.location_on_outlined,
                          size: 14,
                          color: SevaColors.textTertiary,
                        ),
                        const SizedBox(width: 2),
                        Text(
                          _distanceDisplay,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                        const SizedBox(width: 12),
                      ],
                      if (responseTimeMinutes > 0) ...[
                        Icon(
                          Icons.access_time,
                          size: 14,
                          color: SevaColors.textTertiary,
                        ),
                        const SizedBox(width: 2),
                        Text(
                          _responseTimeDisplay,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                        const SizedBox(width: 12),
                      ],
                      if (hourlyRate != null) ...[
                        Text(
                          '${currency ?? "INR"} $hourlyRate/hr',
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.primary,
                                    fontWeight: FontWeight.w600,
                                  ),
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
