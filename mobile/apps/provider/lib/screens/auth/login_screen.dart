import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import '../../main.dart';

class ProviderLoginScreen extends ConsumerStatefulWidget {
  const ProviderLoginScreen({super.key});

  @override
  ConsumerState<ProviderLoginScreen> createState() =>
      _ProviderLoginScreenState();
}

class _ProviderLoginScreenState extends ConsumerState<ProviderLoginScreen> {
  final _phoneController = TextEditingController();
  bool _isOtpSent = false;
  bool _isLoading = false;
  String? _errorText;
  String _phone = '';

  @override
  void dispose() {
    _phoneController.dispose();
    super.dispose();
  }

  Future<void> _requestOtp() async {
    final phone = _phoneController.text.trim();
    if (phone.length < 10) {
      setState(() => _errorText = 'Please enter a valid phone number');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorText = null;
    });

    final success = await ref.read(authServiceProvider).requestOtp(phone);

    if (mounted) {
      setState(() {
        _isLoading = false;
        if (success) {
          _isOtpSent = true;
          _phone = phone;
        } else {
          _errorText = 'Failed to send OTP. Please try again.';
        }
      });
    }
  }

  Future<void> _verifyOtp(String code) async {
    setState(() {
      _isLoading = true;
      _errorText = null;
    });

    final user = await ref.read(authServiceProvider).verifyOtp(_phone, code);

    if (mounted) {
      setState(() => _isLoading = false);

      if (user != null) {
        context.go('/');
      } else {
        setState(() => _errorText = 'Invalid OTP. Please try again.');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Spacer(flex: 1),

              // Branding
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: SevaColors.primaryFaded,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Icon(
                  Icons.handyman,
                  color: SevaColors.primary,
                  size: 36,
                ),
              ),
              const SizedBox(height: 24),
              Text(
                _isOtpSent ? 'Verify OTP' : 'Seva Provider',
                style: Theme.of(context).textTheme.displaySmall?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
              ),
              const SizedBox(height: 8),
              Text(
                _isOtpSent
                    ? 'Enter the 6-digit code sent to +91$_phone'
                    : 'Grow your business with trusted customers',
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      color: SevaColors.textSecondary,
                    ),
              ),
              const SizedBox(height: 40),

              if (!_isOtpSent) ...[
                SevaPhoneInput(
                  controller: _phoneController,
                  errorText: _errorText,
                  onChanged: (_) {
                    if (_errorText != null) {
                      setState(() => _errorText = null);
                    }
                  },
                ),
                const SizedBox(height: 24),
                SevaButton(
                  label: 'Send OTP',
                  isLoading: _isLoading,
                  onPressed: _requestOtp,
                ),
              ] else ...[
                if (_errorText != null) ...[
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: SevaColors.errorLight,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.error_outline,
                            color: SevaColors.error, size: 18),
                        const SizedBox(width: 8),
                        Text(
                          _errorText!,
                          style: const TextStyle(
                            color: SevaColors.error,
                            fontSize: 13,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                ],
                SevaOtpInput(
                  onCompleted: _verifyOtp,
                ),
                const SizedBox(height: 24),
                if (_isLoading)
                  const Center(child: CircularProgressIndicator())
                else ...[
                  Center(
                    child: TextButton(
                      onPressed: () => setState(() => _isOtpSent = false),
                      child: const Text('Change phone number'),
                    ),
                  ),
                  Center(
                    child: TextButton(
                      onPressed: _requestOtp,
                      child: const Text('Resend OTP'),
                    ),
                  ),
                ],
              ],

              const Spacer(flex: 2),
            ],
          ),
        ),
      ),
    );
  }
}
