import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();
  final _postcodeController = TextEditingController();
  bool _isLoading = false;
  String? _errorText;

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    _postcodeController.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _errorText = null;
    });

    final user = await ref.read(authServiceProvider).register(
          name: _nameController.text.trim(),
          phone: _phoneController.text.trim(),
          role: 'customer',
          email: _emailController.text.trim().isNotEmpty
              ? _emailController.text.trim()
              : null,
          postcode: _postcodeController.text.trim().isNotEmpty
              ? _postcodeController.text.trim()
              : null,
        );

    if (mounted) {
      setState(() => _isLoading = false);

      if (user != null) {
        context.go('/');
      } else {
        setState(
            () => _errorText = 'Registration failed. Please try again.');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Create Account')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Join Seva',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Create your account to find trusted service providers',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: SevaColors.textSecondary,
                      ),
                ),
                const SizedBox(height: 32),

                if (_errorText != null) ...[
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: SevaColors.errorLight,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      _errorText!,
                      style: const TextStyle(color: SevaColors.error),
                    ),
                  ),
                  const SizedBox(height: 16),
                ],

                SevaInput(
                  label: 'Full Name',
                  hint: 'Enter your full name',
                  controller: _nameController,
                  prefixIcon: Icons.person_outline,
                  textInputAction: TextInputAction.next,
                  validator: (val) {
                    if (val == null || val.trim().isEmpty) {
                      return 'Please enter your name';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 16),

                SevaPhoneInput(
                  controller: _phoneController,
                ),
                const SizedBox(height: 16),

                SevaInput(
                  label: 'Email (optional)',
                  hint: 'your@email.com',
                  controller: _emailController,
                  prefixIcon: Icons.email_outlined,
                  keyboardType: TextInputType.emailAddress,
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: 16),

                SevaInput(
                  label: 'Postcode (optional)',
                  hint: 'e.g. 682001',
                  controller: _postcodeController,
                  prefixIcon: Icons.location_on_outlined,
                  keyboardType: TextInputType.number,
                  textInputAction: TextInputAction.done,
                ),
                const SizedBox(height: 32),

                SevaButton(
                  label: 'Create Account',
                  isLoading: _isLoading,
                  onPressed: _register,
                ),
                const SizedBox(height: 16),

                Center(
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        'Already have an account? ',
                        style:
                            Theme.of(context).textTheme.bodyMedium?.copyWith(
                                  color: SevaColors.textSecondary,
                                ),
                      ),
                      TextButton(
                        onPressed: () => context.pop(),
                        child: const Text('Login'),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
