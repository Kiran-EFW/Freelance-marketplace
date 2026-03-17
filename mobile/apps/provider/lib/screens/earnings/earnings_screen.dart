import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class EarningsScreen extends ConsumerStatefulWidget {
  const EarningsScreen({super.key});

  @override
  ConsumerState<EarningsScreen> createState() => _EarningsScreenState();
}

class _EarningsScreenState extends ConsumerState<EarningsScreen> {
  EarningsSummary? _earnings;
  List<Payout> _payouts = [];
  bool _isLoading = true;
  String _selectedPeriod = 'month';

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    final providerRepo = ref.read(providerRepositoryProvider);
    final results = await Future.wait([
      providerRepo.getEarnings(period: _selectedPeriod),
      providerRepo.getPayoutHistory(),
    ]);

    if (mounted) {
      setState(() {
        _earnings = results[0] as EarningsSummary?;
        _payouts = (results[1] as PaginatedResult<Payout>).items;
        _isLoading = false;
      });
    }
  }

  Future<void> _requestPayout() async {
    if (_earnings == null || _earnings!.availableBalance <= 0) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Request Payout'),
        content: Text(
          'Request payout of INR ${_earnings!.availableBalance.toStringAsFixed(0)} to your bank account?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Request'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    final payout = await ref
        .read(providerRepositoryProvider)
        .requestPayout(_earnings!.availableBalance);

    if (mounted) {
      if (payout != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Payout requested!')),
        );
        _loadData();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to request payout')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Earnings'),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadData,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Summary cards
                    Row(
                      children: [
                        Expanded(
                          child: SevaStatCard(
                            label: 'Total Earnings',
                            value:
                                'INR ${_earnings?.totalEarnings.toStringAsFixed(0) ?? "0"}',
                            icon: Icons.account_balance_wallet,
                            iconColor: SevaColors.primary,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: SevaStatCard(
                            label: 'Available',
                            value:
                                'INR ${_earnings?.availableBalance.toStringAsFixed(0) ?? "0"}',
                            icon: Icons.payments,
                            iconColor: SevaColors.success,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: SevaStatCard(
                            label: 'Pending',
                            value:
                                'INR ${_earnings?.pendingPayout.toStringAsFixed(0) ?? "0"}',
                            icon: Icons.hourglass_empty,
                            iconColor: SevaColors.warning,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: SevaStatCard(
                            label: 'Avg. Job Value',
                            value:
                                'INR ${_earnings?.averageJobValue.toStringAsFixed(0) ?? "0"}',
                            icon: Icons.trending_up,
                            iconColor: SevaColors.secondary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 20),

                    // Period selector
                    Row(
                      children: [
                        Text('Period',
                            style: Theme.of(context).textTheme.titleMedium),
                        const Spacer(),
                        SegmentedButton<String>(
                          segments: const [
                            ButtonSegment(value: 'week', label: Text('Week')),
                            ButtonSegment(value: 'month', label: Text('Month')),
                            ButtonSegment(value: 'year', label: Text('Year')),
                          ],
                          selected: {_selectedPeriod},
                          onSelectionChanged: (value) {
                            setState(
                                () => _selectedPeriod = value.first);
                            _loadData();
                          },
                          style: const ButtonStyle(
                            visualDensity: VisualDensity.compact,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),

                    // Earnings chart placeholder
                    if (_earnings != null &&
                        _earnings!.chartData.isNotEmpty) ...[
                      SevaCard(
                        child: SizedBox(
                          height: 200,
                          child: _EarningsChart(data: _earnings!.chartData),
                        ),
                      ),
                      const SizedBox(height: 20),
                    ],

                    // Payout button
                    SevaButton(
                      label: 'Request Payout',
                      icon: Icons.account_balance,
                      onPressed: (_earnings?.availableBalance ?? 0) > 0
                          ? _requestPayout
                          : null,
                    ),
                    const SizedBox(height: 24),

                    // Payout history
                    Text('Payout History',
                        style: Theme.of(context).textTheme.headlineSmall),
                    const SizedBox(height: 12),
                    if (_payouts.isEmpty)
                      Padding(
                        padding: const EdgeInsets.all(20),
                        child: Center(
                          child: Text(
                            'No payouts yet',
                            style: Theme.of(context)
                                .textTheme
                                .bodyMedium
                                ?.copyWith(color: SevaColors.textTertiary),
                          ),
                        ),
                      )
                    else
                      ..._payouts.map((payout) => _PayoutTile(payout: payout)),
                  ],
                ),
              ),
            ),
    );
  }
}

class _EarningsChart extends StatelessWidget {
  final List<EarningsDataPoint> data;

  const _EarningsChart({required this.data});

  @override
  Widget build(BuildContext context) {
    if (data.isEmpty) return const SizedBox.shrink();

    final maxAmount =
        data.map((d) => d.amount).reduce((a, b) => a > b ? a : b);

    return Column(
      children: [
        Expanded(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: data.map((point) {
              final height = maxAmount > 0 ? (point.amount / maxAmount) : 0.0;
              return Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      Text(
                        point.amount > 999
                            ? '${(point.amount / 1000).toStringAsFixed(1)}K'
                            : point.amount.toStringAsFixed(0),
                        style: const TextStyle(
                          fontSize: 9,
                          color: SevaColors.textTertiary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Flexible(
                        child: FractionallySizedBox(
                          heightFactor: height.clamp(0.05, 1.0),
                          child: Container(
                            decoration: BoxDecoration(
                              color: SevaColors.primary,
                              borderRadius: BorderRadius.circular(4),
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
        ),
        const SizedBox(height: 8),
        Row(
          children: data.map((point) {
            return Expanded(
              child: Text(
                point.label,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 9,
                  color: SevaColors.textTertiary,
                ),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }
}

class _PayoutTile extends StatelessWidget {
  final Payout payout;

  const _PayoutTile({required this.payout});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: _statusColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(
          Icons.account_balance,
          color: _statusColor,
          size: 20,
        ),
      ),
      title: Text(
        '${payout.currency} ${payout.amount.toStringAsFixed(0)}',
        style: Theme.of(context).textTheme.titleSmall,
      ),
      subtitle: Text(
        DateFormat('d MMM yyyy').format(payout.createdAt),
        style: Theme.of(context)
            .textTheme
            .bodySmall
            ?.copyWith(color: SevaColors.textTertiary),
      ),
      trailing: StatusBadge(status: payout.status, isCompact: true),
    );
  }

  Color get _statusColor {
    switch (payout.status) {
      case 'paid':
        return SevaColors.success;
      case 'processing':
        return SevaColors.info;
      case 'pending':
        return SevaColors.warning;
      default:
        return SevaColors.neutral500;
    }
  }
}
