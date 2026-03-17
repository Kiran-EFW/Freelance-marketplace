import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class JobDetailScreen extends ConsumerStatefulWidget {
  final String jobId;

  const JobDetailScreen({super.key, required this.jobId});

  @override
  ConsumerState<JobDetailScreen> createState() => _JobDetailScreenState();
}

class _JobDetailScreenState extends ConsumerState<JobDetailScreen> {
  Job? _job;
  bool _isLoading = true;
  bool _isActioning = false;

  @override
  void initState() {
    super.initState();
    _loadJob();
  }

  Future<void> _loadJob() async {
    setState(() => _isLoading = true);
    final job = await ref.read(jobRepositoryProvider).getJob(widget.jobId);
    if (mounted) {
      setState(() {
        _job = job;
        _isLoading = false;
      });
    }
  }

  Future<void> _cancelJob() async {
    final reason = await showDialog<String>(
      context: context,
      builder: (context) {
        final controller = TextEditingController();
        return AlertDialog(
          title: const Text('Cancel Job'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('Are you sure you want to cancel this job?'),
              const SizedBox(height: 12),
              TextField(
                controller: controller,
                decoration: const InputDecoration(
                  hintText: 'Reason for cancellation',
                ),
                maxLines: 2,
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Keep Job'),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context, controller.text),
              child: const Text(
                'Cancel Job',
                style: TextStyle(color: SevaColors.error),
              ),
            ),
          ],
        );
      },
    );

    if (reason == null || !mounted) return;

    setState(() => _isActioning = true);
    final success = await ref
        .read(jobRepositoryProvider)
        .cancelJob(widget.jobId, reason: reason);

    if (mounted) {
      setState(() => _isActioning = false);
      if (success) {
        _loadJob();
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Job cancelled')),
        );
      }
    }
  }

  Future<void> _submitReview() async {
    int selectedRating = 0;
    final commentController = TextEditingController();

    final result = await showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Padding(
              padding: EdgeInsets.fromLTRB(
                20,
                20,
                20,
                MediaQuery.of(context).viewInsets.bottom + 20,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text('Rate this service',
                      style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 20),
                  RatingInput(
                    rating: selectedRating,
                    onChanged: (val) {
                      setModalState(() => selectedRating = val);
                    },
                  ),
                  const SizedBox(height: 16),
                  SevaInput(
                    label: 'Comment (optional)',
                    hint: 'How was the service?',
                    controller: commentController,
                    maxLines: 3,
                  ),
                  const SizedBox(height: 20),
                  SevaButton(
                    label: 'Submit Review',
                    onPressed: selectedRating > 0
                        ? () {
                            Navigator.pop(context, {
                              'rating': selectedRating,
                              'comment': commentController.text,
                            });
                          }
                        : null,
                  ),
                ],
              ),
            );
          },
        );
      },
    );

    if (result == null || !mounted) return;

    await ref.read(jobRepositoryProvider).submitReview(
          widget.jobId,
          rating: result['rating'] as int,
          comment: result['comment'] as String?,
        );

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Review submitted!')),
      );
      _loadJob();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    if (_job == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: Text('Job not found')),
      );
    }

    final job = _job!;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Job Details'),
        actions: [
          if (job.isCancellable)
            TextButton(
              onPressed: _isActioning ? null : _cancelJob,
              child: const Text(
                'Cancel',
                style: TextStyle(color: SevaColors.error),
              ),
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _loadJob,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Title and status
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Text(
                      job.title,
                      style:
                          Theme.of(context).textTheme.headlineMedium?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  StatusBadge(status: job.status.toJson()),
                ],
              ),
              const SizedBox(height: 4),
              if (job.categoryName != null)
                Text(
                  job.categoryName!,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: SevaColors.textSecondary,
                      ),
                ),
              const SizedBox(height: 20),

              // Price
              if (job.agreedPrice != null || job.budgetMin != null)
                SevaCard(
                  child: Row(
                    children: [
                      const Icon(Icons.currency_rupee,
                          color: SevaColors.primary),
                      const SizedBox(width: 8),
                      Text(
                        job.budgetDisplay,
                        style: Theme.of(context)
                            .textTheme
                            .headlineSmall
                            ?.copyWith(
                              color: SevaColors.primary,
                              fontWeight: FontWeight.w700,
                            ),
                      ),
                    ],
                  ),
                ),
              const SizedBox(height: 16),

              // Description
              Text('Description',
                  style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              Text(
                job.description,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: SevaColors.textSecondary,
                      height: 1.5,
                    ),
              ),
              const SizedBox(height: 20),

              // Timeline
              Text('Timeline',
                  style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              _TimelineItem(
                label: 'Created',
                value: DateFormat('d MMM yyyy, h:mm a').format(job.createdAt),
                isFirst: true,
              ),
              if (job.scheduledAt != null)
                _TimelineItem(
                  label: 'Scheduled',
                  value:
                      DateFormat('d MMM yyyy, h:mm a').format(job.scheduledAt!),
                ),
              if (job.startedAt != null)
                _TimelineItem(
                  label: 'Started',
                  value:
                      DateFormat('d MMM yyyy, h:mm a').format(job.startedAt!),
                ),
              if (job.completedAt != null)
                _TimelineItem(
                  label: 'Completed',
                  value:
                      DateFormat('d MMM yyyy, h:mm a').format(job.completedAt!),
                  isLast: true,
                ),
              const SizedBox(height: 20),

              // Provider info
              if (job.providerName != null) ...[
                Text('Provider',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                SevaCard(
                  onTap: job.providerId != null
                      ? () => context.push('/provider/${job.providerId}')
                      : null,
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 20,
                        backgroundColor: SevaColors.primaryFaded,
                        child: Text(
                          job.providerName![0].toUpperCase(),
                          style: const TextStyle(
                            fontWeight: FontWeight.w700,
                            color: SevaColors.primary,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              job.providerName!,
                              style: Theme.of(context).textTheme.titleSmall,
                            ),
                            if (job.providerRating != null)
                              Row(
                                children: [
                                  StarRating(
                                    rating: job.providerRating!,
                                    size: 12,
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    job.providerRating!.toStringAsFixed(1),
                                    style: Theme.of(context)
                                        .textTheme
                                        .bodySmall
                                        ?.copyWith(
                                          color: SevaColors.textTertiary,
                                        ),
                                  ),
                                ],
                              ),
                          ],
                        ),
                      ),
                      const Icon(Icons.chevron_right,
                          color: SevaColors.neutral400),
                    ],
                  ),
                ),
                const SizedBox(height: 20),
              ],

              // Location
              if (job.address != null) ...[
                Text('Location',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                SevaCard(
                  child: Row(
                    children: [
                      const Icon(Icons.location_on,
                          color: SevaColors.secondary),
                      const SizedBox(width: 8),
                      Expanded(child: Text(job.address!)),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
      bottomNavigationBar: job.status == JobStatus.completed
          ? SafeArea(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: SevaButton(
                  label: 'Leave a Review',
                  icon: Icons.star_outline,
                  onPressed: _submitReview,
                ),
              ),
            )
          : null,
    );
  }
}

class _TimelineItem extends StatelessWidget {
  final String label;
  final String value;
  final bool isFirst;
  final bool isLast;

  const _TimelineItem({
    required this.label,
    required this.value,
    this.isFirst = false,
    this.isLast = false,
  });

  @override
  Widget build(BuildContext context) {
    return IntrinsicHeight(
      child: Row(
        children: [
          SizedBox(
            width: 24,
            child: Column(
              children: [
                if (!isFirst)
                  Expanded(
                    child: Container(width: 2, color: SevaColors.neutral200),
                  ),
                Container(
                  width: 10,
                  height: 10,
                  decoration: const BoxDecoration(
                    color: SevaColors.primary,
                    shape: BoxShape.circle,
                  ),
                ),
                if (!isLast)
                  Expanded(
                    child: Container(width: 2, color: SevaColors.neutral200),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                          color: SevaColors.textTertiary,
                        ),
                  ),
                  Text(value, style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
