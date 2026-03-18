import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class CreateJobScreen extends ConsumerStatefulWidget {
  final String? categoryId;
  final String? providerId;

  const CreateJobScreen({super.key, this.categoryId, this.providerId});

  @override
  ConsumerState<CreateJobScreen> createState() => _CreateJobScreenState();
}

class _CreateJobScreenState extends ConsumerState<CreateJobScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _budgetMinController = TextEditingController();
  final _budgetMaxController = TextEditingController();
  final _addressController = TextEditingController();

  int _currentStep = 0;
  String? _selectedCategoryId;
  String _urgency = 'normal';
  DateTime? _scheduledDate;
  TimeOfDay? _scheduledTime;
  bool _isSubmitting = false;
  List<Category> _categories = [];
  LocationResult? _location;

  @override
  void initState() {
    super.initState();
    _selectedCategoryId = widget.categoryId;
    _loadCategories();
    _detectLocation();
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _budgetMinController.dispose();
    _budgetMaxController.dispose();
    _addressController.dispose();
    super.dispose();
  }

  Future<void> _loadCategories() async {
    final result =
        await ref.read(providerRepositoryProvider).getCategories();
    if (mounted && result.isSuccess) {
      setState(() => _categories = result.dataOrNull ?? []);
    }
  }

  Future<void> _detectLocation() async {
    final location =
        await ref.read(locationServiceProvider).getCurrentLocation();
    if (mounted && location != null) {
      setState(() {
        _location = location;
        _addressController.text = location.formattedAddress ?? '';
      });
    }
  }

  Future<void> _selectDate() async {
    final date = await showDatePicker(
      context: context,
      initialDate: _scheduledDate ?? DateTime.now().add(const Duration(days: 1)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 90)),
    );
    if (date != null && mounted) {
      setState(() => _scheduledDate = date);
    }
  }

  Future<void> _selectTime() async {
    final time = await showTimePicker(
      context: context,
      initialTime: _scheduledTime ?? const TimeOfDay(hour: 9, minute: 0),
    );
    if (time != null && mounted) {
      setState(() => _scheduledTime = time);
    }
  }

  DateTime? get _scheduledAt {
    if (_scheduledDate == null) return null;
    final time = _scheduledTime ?? const TimeOfDay(hour: 9, minute: 0);
    return DateTime(
      _scheduledDate!.year,
      _scheduledDate!.month,
      _scheduledDate!.day,
      time.hour,
      time.minute,
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedCategoryId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select a category')),
      );
      return;
    }

    setState(() => _isSubmitting = true);

    final jobRepo = ref.read(jobRepositoryProvider);
    final result = await jobRepo.createJob(
      categoryId: _selectedCategoryId!,
      title: _titleController.text.trim(),
      description: _descriptionController.text.trim(),
      budgetMin: double.tryParse(_budgetMinController.text),
      budgetMax: double.tryParse(_budgetMaxController.text),
      postcode: _location?.postcode,
      latitude: _location?.latitude,
      longitude: _location?.longitude,
      address: _addressController.text.trim().isNotEmpty
          ? _addressController.text.trim()
          : null,
      scheduledAt: _scheduledAt,
      urgency: _urgency,
    );

    if (mounted) {
      setState(() => _isSubmitting = false);

      final job = result.dataOrNull;
      if (job != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Job posted successfully!')),
        );
        context.go('/job/${job.id}');
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Failed to create job. Please try again.'),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Create Job'),
      ),
      body: Form(
        key: _formKey,
        child: Stepper(
          currentStep: _currentStep,
          onStepContinue: () {
            if (_currentStep < 3) {
              setState(() => _currentStep++);
            } else {
              _submit();
            }
          },
          onStepCancel: () {
            if (_currentStep > 0) {
              setState(() => _currentStep--);
            } else {
              Navigator.pop(context);
            }
          },
          controlsBuilder: (context, details) {
            return Padding(
              padding: const EdgeInsets.only(top: 16),
              child: Row(
                children: [
                  Expanded(
                    child: SevaButton(
                      label: _currentStep == 3 ? 'Post Job' : 'Next',
                      isLoading: _isSubmitting,
                      onPressed: details.onStepContinue,
                    ),
                  ),
                  if (_currentStep > 0) ...[
                    const SizedBox(width: 12),
                    Expanded(
                      child: SevaButton(
                        label: 'Back',
                        variant: SevaButtonVariant.outline,
                        onPressed: details.onStepCancel,
                      ),
                    ),
                  ],
                ],
              ),
            );
          },
          steps: [
            // Step 1: Category
            Step(
              title: const Text('Category'),
              subtitle: _selectedCategoryId != null
                  ? Text(_categories
                      .where((c) => c.id == _selectedCategoryId)
                      .map((c) => c.name)
                      .firstOrNull ?? 'Selected')
                  : null,
              isActive: _currentStep >= 0,
              state: _currentStep > 0 ? StepState.complete : StepState.indexed,
              content: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('What type of service do you need?'),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _categories.map((cat) {
                      final isSelected = _selectedCategoryId == cat.id;
                      return FilterChip(
                        label: Text(cat.name),
                        selected: isSelected,
                        onSelected: (selected) {
                          setState(() {
                            _selectedCategoryId = selected ? cat.id : null;
                          });
                        },
                      );
                    }).toList(),
                  ),
                ],
              ),
            ),

            // Step 2: Details
            Step(
              title: const Text('Details'),
              isActive: _currentStep >= 1,
              state: _currentStep > 1 ? StepState.complete : StepState.indexed,
              content: Column(
                children: [
                  SevaInput(
                    label: 'Job Title',
                    hint: 'e.g. Fix leaking kitchen tap',
                    controller: _titleController,
                    validator: (val) {
                      if (val == null || val.trim().isEmpty) {
                        return 'Please enter a title';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  SevaInput(
                    label: 'Description',
                    hint: 'Describe the job in detail...',
                    controller: _descriptionController,
                    maxLines: 4,
                    validator: (val) {
                      if (val == null || val.trim().isEmpty) {
                        return 'Please enter a description';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  Text('Urgency',
                      style: Theme.of(context).textTheme.titleSmall),
                  const SizedBox(height: 8),
                  SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(
                        value: 'low',
                        label: Text('Low'),
                      ),
                      ButtonSegment(
                        value: 'normal',
                        label: Text('Normal'),
                      ),
                      ButtonSegment(
                        value: 'high',
                        label: Text('Urgent'),
                      ),
                      ButtonSegment(
                        value: 'emergency',
                        label: Text('SOS'),
                      ),
                    ],
                    selected: {_urgency},
                    onSelectionChanged: (value) {
                      setState(() => _urgency = value.first);
                    },
                  ),
                ],
              ),
            ),

            // Step 3: Schedule
            Step(
              title: const Text('Schedule'),
              isActive: _currentStep >= 2,
              state: _currentStep > 2 ? StepState.complete : StepState.indexed,
              content: Column(
                children: [
                  SevaInput(
                    label: 'Address',
                    hint: 'Service location',
                    controller: _addressController,
                    prefixIcon: Icons.location_on_outlined,
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: SevaButton(
                          label: _scheduledDate != null
                              ? DateFormat('d MMM yyyy')
                                  .format(_scheduledDate!)
                              : 'Select Date',
                          variant: SevaButtonVariant.outline,
                          icon: Icons.calendar_today_outlined,
                          onPressed: _selectDate,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: SevaButton(
                          label: _scheduledTime != null
                              ? _scheduledTime!.format(context)
                              : 'Select Time',
                          variant: SevaButtonVariant.outline,
                          icon: Icons.access_time,
                          onPressed: _selectTime,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // Step 4: Budget
            Step(
              title: const Text('Budget'),
              isActive: _currentStep >= 3,
              content: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Set your budget range (optional)'),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: SevaInput(
                          label: 'Minimum',
                          hint: '500',
                          controller: _budgetMinController,
                          keyboardType: TextInputType.number,
                          prefixIcon: Icons.currency_rupee,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: SevaInput(
                          label: 'Maximum',
                          hint: '2000',
                          controller: _budgetMaxController,
                          keyboardType: TextInputType.number,
                          prefixIcon: Icons.currency_rupee,
                        ),
                      ),
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
