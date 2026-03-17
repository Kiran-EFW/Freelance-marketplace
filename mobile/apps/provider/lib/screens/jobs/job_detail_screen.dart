import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class ProviderJobDetailScreen extends ConsumerStatefulWidget {
  final String jobId;

  const ProviderJobDetailScreen({super.key, required this.jobId});

  @override
  ConsumerState<ProviderJobDetailScreen> createState() =>
      _ProviderJobDetailScreenState();
}

class _ProviderJobDetailScreenState
    extends ConsumerState<ProviderJobDetailScreen> {
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

  Future<void> _acceptJob() async {
    setState(() => _isActioning = true);
    final job = await ref.read(jobRepositoryProvider).acceptJob(widget.jobId);
    if (mounted) {
      setState(() {
        _isActioning = false;
        if (job != null) _job = job;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Job accepted!')),
      );
    }
  }

  Future<void> _declineJob() async {
    final reason = await showDialog<String>(
      context: context,
      builder: (context) {
        final controller = TextEditingController();
        return AlertDialog(
          title: const Text('Decline Job'),
          content: TextField(
            controller: controller,
            decoration:
                const InputDecoration(hintText: 'Reason (optional)'),
            maxLines: 2,
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context, controller.text),
              child: const Text('Decline',
                  style: TextStyle(color: SevaColors.error)),
            ),
          ],
        );
      },
    );

    if (reason == null || !mounted) return;

    setState(() => _isActioning = true);
    await ref.read(jobRepositoryProvider).declineJob(
          widget.jobId,
          reason: reason.isNotEmpty ? reason : null,
        );
    if (mounted) {
      setState(() => _isActioning = false);
      _loadJob();
    }
  }

  Future<void> _startJob() async {
    setState(() => _isActioning = true);
    final job = await ref
        .read(jobRepositoryProvider)
        .updateStatus(widget.jobId, 'in_progress');
    if (mounted) {
      setState(() {
        _isActioning = false;
        if (job != null) _job = job;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Job started!')),
      );
    }
  }

  Future<void> _completeJob() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Complete Job'),
        content: const Text(
          'Are you sure you want to mark this job as completed?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Complete'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    setState(() => _isActioning = true);
    final job =
        await ref.read(jobRepositoryProvider).completeJob(widget.jobId);
    if (mounted) {
      setState(() {
        _isActioning = false;
        if (job != null) _job = job;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Job completed!')),
      );
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
                      style: Theme.of(context)
                          .textTheme
                          .headlineMedium
                          ?.copyWith(fontWeight: FontWeight.w700),
                    ),
                  ),
                  StatusBadge(status: job.status.toJson()),
                ],
              ),
              if (job.categoryName != null) ...[
                const SizedBox(height: 4),
                Text(
                  job.categoryName!,
                  style: Theme.of(context)
                      .textTheme
                      .bodyMedium
                      ?.copyWith(color: SevaColors.textSecondary),
                ),
              ],
              const SizedBox(height: 16),

              // Price
              SevaCard(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Budget',
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(color: SevaColors.textTertiary)),
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
                    if (job.urgency != JobUrgency.normal)
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: job.urgency == JobUrgency.emergency
                              ? SevaColors.errorLight
                              : SevaColors.warningLight,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Text(
                          job.urgency.name.toUpperCase(),
                          style: TextStyle(
                            fontWeight: FontWeight.w700,
                            fontSize: 11,
                            color: job.urgency == JobUrgency.emergency
                                ? SevaColors.error
                                : SevaColors.warning,
                          ),
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
              const SizedBox(height: 16),

              // Customer info
              if (job.customerName != null) ...[
                Text('Customer',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                SevaCard(
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 20,
                        backgroundColor: SevaColors.secondaryFaded,
                        child: Text(
                          job.customerName![0].toUpperCase(),
                          style: const TextStyle(
                            fontWeight: FontWeight.w700,
                            color: SevaColors.secondary,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Text(
                        job.customerName!,
                        style: Theme.of(context).textTheme.titleSmall,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
              ],

              // Schedule and location
              if (job.scheduledAt != null) ...[
                SevaCard(
                  child: Row(
                    children: [
                      const Icon(Icons.calendar_today,
                          color: SevaColors.info, size: 20),
                      const SizedBox(width: 8),
                      Text(
                        DateFormat('EEEE, d MMMM yyyy, h:mm a')
                            .format(job.scheduledAt!),
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 8),
              ],
              if (job.address != null)
                SevaCard(
                  child: Row(
                    children: [
                      const Icon(Icons.location_on,
                          color: SevaColors.secondary, size: 20),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          job.address!,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
      ),
      bottomNavigationBar: _buildActions(context, job),
    );
  }

  Widget? _buildActions(BuildContext context, Job job) {
    switch (job.status) {
      case JobStatus.posted:
      case JobStatus.matched:
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: SevaButton(
                    label: 'Decline',
                    variant: SevaButtonVariant.outline,
                    isLoading: _isActioning,
                    onPressed: _declineJob,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  flex: 2,
                  child: SevaButton(
                    label: 'Accept Job',
                    icon: Icons.check,
                    isLoading: _isActioning,
                    onPressed: _acceptJob,
                  ),
                ),
              ],
            ),
          ),
        );

      case JobStatus.accepted:
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: SevaButton(
              label: 'Start Job',
              icon: Icons.play_arrow,
              isLoading: _isActioning,
              onPressed: _startJob,
            ),
          ),
        );

      case JobStatus.inProgress:
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: SevaButton(
              label: 'Complete Job',
              icon: Icons.check_circle_outline,
              isLoading: _isActioning,
              onPressed: _completeJob,
            ),
          ),
        );

      default:
        return null;
    }
  }
}
