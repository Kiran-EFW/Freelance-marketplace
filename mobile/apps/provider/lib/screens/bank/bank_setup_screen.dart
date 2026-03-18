import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:seva_ui_kit/ui_kit.dart';

/// Screen for setting up bank account details for payouts.
class BankSetupScreen extends ConsumerStatefulWidget {
  const BankSetupScreen({super.key});

  @override
  ConsumerState<BankSetupScreen> createState() => _BankSetupScreenState();
}

class _BankSetupScreenState extends ConsumerState<BankSetupScreen> {
  final _formKey = GlobalKey<FormState>();
  final _accountNumberController = TextEditingController();
  final _confirmAccountController = TextEditingController();
  final _ifscController = TextEditingController();
  final _accountHolderController = TextEditingController();
  final _upiIdController = TextEditingController();

  bool _isSubmitting = false;
  bool _isVerified = false;
  String? _bankName;
  String? _branchName;

  @override
  void dispose() {
    _accountNumberController.dispose();
    _confirmAccountController.dispose();
    _ifscController.dispose();
    _accountHolderController.dispose();
    _upiIdController.dispose();
    super.dispose();
  }

  Future<void> _lookupIfsc() async {
    final ifsc = _ifscController.text.trim();
    if (ifsc.length != 11) return;

    // TODO: Replace with actual IFSC lookup API call.
    // final result = await apiClient.lookupIfsc(ifsc);
    setState(() {
      _bankName = 'State Bank of India'; // Placeholder
      _branchName = 'Main Branch'; // Placeholder
    });
  }

  Future<void> _verifyAccount() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isSubmitting = true);

    try {
      // TODO: Replace with actual API call.
      // await apiClient.setupBankAccount(
      //   accountNumber: _accountNumberController.text.trim(),
      //   ifscCode: _ifscController.text.trim().toUpperCase(),
      //   accountHolderName: _accountHolderController.text.trim(),
      //   upiId: _upiIdController.text.trim().isNotEmpty
      //       ? _upiIdController.text.trim()
      //       : null,
      // );
      await Future.delayed(const Duration(seconds: 2));

      if (mounted) {
        setState(() {
          _isSubmitting = false;
          _isVerified = true;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Bank account verified and saved successfully!'),
            backgroundColor: SevaColors.success,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isSubmitting = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Verification failed: $e'),
            backgroundColor: SevaColors.error,
          ),
        );
      }
    }
  }

  String? _validateAccountNumber(String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'Account number is required';
    }
    if (value.trim().length < 9 || value.trim().length > 18) {
      return 'Enter a valid account number (9-18 digits)';
    }
    return null;
  }

  String? _validateConfirmAccount(String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'Please confirm your account number';
    }
    if (value.trim() != _accountNumberController.text.trim()) {
      return 'Account numbers do not match';
    }
    return null;
  }

  String? _validateIfsc(String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'IFSC code is required';
    }
    final ifscRegex = RegExp(r'^[A-Z]{4}0[A-Z0-9]{6}$');
    if (!ifscRegex.hasMatch(value.trim().toUpperCase())) {
      return 'Enter a valid IFSC code (e.g. SBIN0001234)';
    }
    return null;
  }

  String? _validateAccountHolder(String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'Account holder name is required';
    }
    if (value.trim().length < 2) {
      return 'Enter a valid name';
    }
    return null;
  }

  String? _validateUpi(String? value) {
    if (value == null || value.trim().isEmpty) return null; // UPI is optional.
    final upiRegex = RegExp(r'^[\w.\-]+@[\w]+$');
    if (!upiRegex.hasMatch(value.trim())) {
      return 'Enter a valid UPI ID (e.g. name@upi)';
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Bank Account Setup'),
        centerTitle: false,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Success state
              if (_isVerified)
                SevaCard(
                  backgroundColor: SevaColors.successLight,
                  child: Row(
                    children: [
                      const Icon(Icons.check_circle,
                          color: SevaColors.success, size: 28),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Account Verified',
                              style: Theme.of(context)
                                  .textTheme
                                  .titleMedium
                                  ?.copyWith(
                                    color: SevaColors.success,
                                    fontWeight: FontWeight.w700,
                                  ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Your bank account has been verified and will be used for payouts.',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: SevaColors.success,
                                  ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),

              if (!_isVerified) ...[
                // Info card
                SevaCard(
                  backgroundColor: SevaColors.primaryFaded,
                  child: Row(
                    children: [
                      const Icon(Icons.account_balance,
                          color: SevaColors.primary, size: 28),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Payout Configuration',
                              style: Theme.of(context)
                                  .textTheme
                                  .titleSmall
                                  ?.copyWith(
                                    color: SevaColors.primaryDark,
                                    fontWeight: FontWeight.w600,
                                  ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Add your bank details to receive earnings from completed jobs.',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: SevaColors.textSecondary,
                                  ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 24),

                // Account holder name
                SevaInput(
                  label: 'Account Holder Name',
                  hint: 'Enter name as on bank account',
                  controller: _accountHolderController,
                  prefixIcon: Icons.person_outline,
                  textInputAction: TextInputAction.next,
                  keyboardType: TextInputType.name,
                  validator: _validateAccountHolder,
                ),

                const SizedBox(height: 20),

                // Account number
                SevaInput(
                  label: 'Bank Account Number',
                  hint: 'Enter account number',
                  controller: _accountNumberController,
                  prefixIcon: Icons.account_balance_outlined,
                  keyboardType: TextInputType.number,
                  textInputAction: TextInputAction.next,
                  inputFormatters: [
                    FilteringTextInputFormatter.digitsOnly,
                    LengthLimitingTextInputFormatter(18),
                  ],
                  validator: _validateAccountNumber,
                ),

                const SizedBox(height: 20),

                // Confirm account number
                SevaInput(
                  label: 'Confirm Account Number',
                  hint: 'Re-enter account number',
                  controller: _confirmAccountController,
                  prefixIcon: Icons.account_balance_outlined,
                  keyboardType: TextInputType.number,
                  textInputAction: TextInputAction.next,
                  inputFormatters: [
                    FilteringTextInputFormatter.digitsOnly,
                    LengthLimitingTextInputFormatter(18),
                  ],
                  validator: _validateConfirmAccount,
                ),

                const SizedBox(height: 20),

                // IFSC Code
                SevaInput(
                  label: 'IFSC Code',
                  hint: 'Enter IFSC code (e.g. SBIN0001234)',
                  controller: _ifscController,
                  prefixIcon: Icons.code,
                  textInputAction: TextInputAction.next,
                  inputFormatters: [
                    UpperCaseTextFormatter(),
                    LengthLimitingTextInputFormatter(11),
                    FilteringTextInputFormatter.allow(RegExp(r'[A-Z0-9]')),
                  ],
                  validator: _validateIfsc,
                  onChanged: (value) {
                    if (value.length == 11) _lookupIfsc();
                  },
                ),

                // Bank name display (auto-populated from IFSC)
                if (_bankName != null) ...[
                  const SizedBox(height: 8),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle,
                            size: 16, color: SevaColors.success),
                        const SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            '$_bankName${_branchName != null ? " - $_branchName" : ""}',
                            style:
                                Theme.of(context).textTheme.bodySmall?.copyWith(
                                      color: SevaColors.success,
                                      fontWeight: FontWeight.w500,
                                    ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 20),

                // UPI ID (optional)
                SevaInput(
                  label: 'UPI ID (Optional)',
                  hint: 'e.g. yourname@upi',
                  controller: _upiIdController,
                  prefixIcon: Icons.qr_code,
                  textInputAction: TextInputAction.done,
                  helperText: 'Optional. Can be used for instant payouts.',
                  validator: _validateUpi,
                ),

                const SizedBox(height: 12),

                // Security note
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Icon(Icons.lock_outline,
                          size: 16, color: SevaColors.textTertiary),
                      const SizedBox(width: 6),
                      Expanded(
                        child: Text(
                          'Your bank details are encrypted and stored securely. '
                          'They will only be used for processing your payouts.',
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: SevaColors.textTertiary,
                                  ),
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 28),

                // Verify button
                SevaButton(
                  label: 'Verify & Save',
                  icon: Icons.verified_outlined,
                  isLoading: _isSubmitting,
                  onPressed: !_isSubmitting ? _verifyAccount : null,
                ),
              ],

              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }
}

/// TextInputFormatter that converts input to uppercase.
class UpperCaseTextFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    return TextEditingValue(
      text: newValue.text.toUpperCase(),
      selection: newValue.selection,
    );
  }
}
