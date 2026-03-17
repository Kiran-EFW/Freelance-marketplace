import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class JobListScreen extends ConsumerStatefulWidget {
  const JobListScreen({super.key});

  @override
  ConsumerState<JobListScreen> createState() => _JobListScreenState();
}

class _JobListScreenState extends ConsumerState<JobListScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final Map<String, List<Job>> _jobsByTab = {
    'available': [],
    'active': [],
    'completed': [],
  };
  final Map<String, bool> _loadingByTab = {
    'available': true,
    'active': true,
    'completed': true,
  };

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _loadAll();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadAll() async {
    await Future.wait([
      _loadJobs('available', 'posted'),
      _loadJobs('active', 'in_progress'),
      _loadJobs('completed', 'completed'),
    ]);
  }

  Future<void> _loadJobs(String tab, String status) async {
    setState(() => _loadingByTab[tab] = true);

    final jobRepo = ref.read(jobRepositoryProvider);
    final result = await jobRepo.getJobs(status: status, role: 'provider');

    if (mounted) {
      setState(() {
        _jobsByTab[tab] = result.items;
        _loadingByTab[tab] = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Jobs'),
        bottom: TabBar(
          controller: _tabController,
          labelColor: SevaColors.primary,
          unselectedLabelColor: SevaColors.textTertiary,
          indicatorColor: SevaColors.primary,
          tabs: [
            Tab(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('Available'),
                  if (_jobsByTab['available']!.isNotEmpty) ...[
                    const SizedBox(width: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: SevaColors.primary,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        '${_jobsByTab['available']!.length}',
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.white,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const Tab(text: 'Active'),
            const Tab(text: 'Completed'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _JobTab(
            jobs: _jobsByTab['available']!,
            isLoading: _loadingByTab['available']!,
            emptyMessage: 'No available jobs near you',
            emptySubtitle: 'Check back soon for new opportunities',
            onRefresh: () => _loadJobs('available', 'posted'),
            onTap: (job) => context.push('/job/${job.id}'),
          ),
          _JobTab(
            jobs: _jobsByTab['active']!,
            isLoading: _loadingByTab['active']!,
            emptyMessage: 'No active jobs',
            emptySubtitle: 'Accept a job to get started',
            onRefresh: () => _loadJobs('active', 'in_progress'),
            onTap: (job) => context.push('/job/${job.id}'),
          ),
          _JobTab(
            jobs: _jobsByTab['completed']!,
            isLoading: _loadingByTab['completed']!,
            emptyMessage: 'No completed jobs yet',
            emptySubtitle: 'Your completed jobs will appear here',
            onRefresh: () => _loadJobs('completed', 'completed'),
            onTap: (job) => context.push('/job/${job.id}'),
          ),
        ],
      ),
    );
  }
}

class _JobTab extends StatelessWidget {
  final List<Job> jobs;
  final bool isLoading;
  final String emptyMessage;
  final String emptySubtitle;
  final Future<void> Function() onRefresh;
  final void Function(Job) onTap;

  const _JobTab({
    required this.jobs,
    required this.isLoading,
    required this.emptyMessage,
    required this.emptySubtitle,
    required this.onRefresh,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (jobs.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.work_off_outlined,
                size: 64, color: SevaColors.neutral300),
            const SizedBox(height: 12),
            Text(emptyMessage,
                style: Theme.of(context)
                    .textTheme
                    .bodyLarge
                    ?.copyWith(color: SevaColors.textTertiary)),
            const SizedBox(height: 4),
            Text(emptySubtitle,
                style: Theme.of(context)
                    .textTheme
                    .bodySmall
                    ?.copyWith(color: SevaColors.textTertiary)),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: onRefresh,
      child: ListView.builder(
        padding: const EdgeInsets.all(20),
        itemCount: jobs.length,
        itemBuilder: (context, index) {
          final job = jobs[index];
          return Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: JobCard(
              title: job.title,
              status: job.status.toJson(),
              categoryName: job.categoryName,
              createdAt: job.createdAt,
              scheduledAt: job.scheduledAt,
              priceDisplay: job.budgetDisplay,
              customerName: job.customerName,
              address: job.address,
              urgency: job.urgency.name,
              onTap: () => onTap(job),
            ),
          );
        },
      ),
    );
  }
}
