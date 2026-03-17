import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';

import '../../main.dart';

/// Supported KYC document types.
enum KycDocumentType {
  aadhaar('Aadhaar Card', 'aadhaar'),
  pan('PAN Card', 'pan'),
  drivingLicense('Driving License', 'driving_license'),
  passport('Passport', 'passport'),
  voterIdCard('Voter ID Card', 'voter_id');

  const KycDocumentType(this.label, this.apiValue);
  final String label;
  final String apiValue;
}

/// Screen for uploading KYC verification documents.
class KycScreen extends ConsumerStatefulWidget {
  const KycScreen({super.key});

  @override
  ConsumerState<KycScreen> createState() => _KycScreenState();
}

class _KycScreenState extends ConsumerState<KycScreen> {
  KycDocumentType _selectedType = KycDocumentType.aadhaar;
  File? _frontImage;
  File? _backImage;
  bool _isUploading = false;
  double _uploadProgress = 0.0;
  KycStatus? _verificationStatus;
  String? _rejectionReason;

  final _imagePicker = ImagePicker();

  @override
  void initState() {
    super.initState();
    _loadVerificationStatus();
  }

  Future<void> _loadVerificationStatus() async {
    final authService = ref.read(authServiceProvider);
    final user = authService.currentUser;
    if (user != null) {
      setState(() {
        _verificationStatus = user.kycStatus;
      });
    }
  }

  Future<void> _pickImage({required bool isFront}) async {
    final source = await _showImageSourceDialog();
    if (source == null) return;

    final pickedFile = await _imagePicker.pickImage(
      source: source,
      maxWidth: 1920,
      maxHeight: 1080,
      imageQuality: 85,
    );

    if (pickedFile != null && mounted) {
      setState(() {
        if (isFront) {
          _frontImage = File(pickedFile.path);
        } else {
          _backImage = File(pickedFile.path);
        }
      });
    }
  }

  Future<ImageSource?> _showImageSourceDialog() async {
    return showModalBottomSheet<ImageSource>(
      context: context,
      builder: (context) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.camera_alt),
                title: const Text('Camera'),
                onTap: () => Navigator.pop(context, ImageSource.camera),
              ),
              ListTile(
                leading: const Icon(Icons.photo_library),
                title: const Text('Gallery'),
                onTap: () => Navigator.pop(context, ImageSource.gallery),
              ),
              const SizedBox(height: 8),
            ],
          ),
        );
      },
    );
  }

  Future<void> _uploadDocuments() async {
    if (_frontImage == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please add front side of the document'),
        ),
      );
      return;
    }

    setState(() {
      _isUploading = true;
      _uploadProgress = 0.0;
    });

    try {
      // Simulate upload progress.
      for (int i = 1; i <= 10; i++) {
        await Future.delayed(const Duration(milliseconds: 200));
        if (mounted) {
          setState(() => _uploadProgress = i / 10);
        }
      }

      // TODO: Replace with actual API call.
      // await apiClient.uploadKycDocument(
      //   type: _selectedType.apiValue,
      //   frontImage: _frontImage!.path,
      //   backImage: _backImage?.path,
      // );

      if (mounted) {
        setState(() {
          _verificationStatus = KycStatus.pending;
          _isUploading = false;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Documents uploaded successfully! Verification pending.'),
            backgroundColor: SevaColors.success,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isUploading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Upload failed: $e'),
            backgroundColor: SevaColors.error,
          ),
        );
      }
    }
  }

  void _resetForm() {
    setState(() {
      _frontImage = null;
      _backImage = null;
      _verificationStatus = null;
      _rejectionReason = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('KYC Verification'),
        centerTitle: false,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Verification status banner
            if (_verificationStatus != null)
              _StatusBanner(
                status: _verificationStatus!,
                rejectionReason: _rejectionReason,
                onReupload: _verificationStatus == KycStatus.rejected
                    ? _resetForm
                    : null,
              ),

            if (_verificationStatus == null ||
                _verificationStatus == KycStatus.rejected) ...[
              const SizedBox(height: 20),

              // Document type selector
              Text(
                'Document Type',
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      color: SevaColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(height: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  border: Border.all(color: SevaColors.neutral200),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: DropdownButtonHideUnderline(
                  child: DropdownButton<KycDocumentType>(
                    value: _selectedType,
                    isExpanded: true,
                    items: KycDocumentType.values.map((type) {
                      return DropdownMenuItem(
                        value: type,
                        child: Text(type.label),
                      );
                    }).toList(),
                    onChanged: (value) {
                      if (value != null) {
                        setState(() {
                          _selectedType = value;
                          _frontImage = null;
                          _backImage = null;
                        });
                      }
                    },
                  ),
                ),
              ),

              const SizedBox(height: 24),

              // Front side
              Text(
                'Front Side',
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      color: SevaColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(height: 8),
              _DocumentUploadArea(
                image: _frontImage,
                label: 'Tap to capture front side',
                onTap: () => _pickImage(isFront: true),
              ),

              const SizedBox(height: 20),

              // Back side
              Text(
                'Back Side (Optional)',
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      color: SevaColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
              ),
              const SizedBox(height: 8),
              _DocumentUploadArea(
                image: _backImage,
                label: 'Tap to capture back side',
                onTap: () => _pickImage(isFront: false),
              ),

              const SizedBox(height: 12),

              // Guidelines
              SevaCard(
                backgroundColor: SevaColors.infoLight,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.info_outline,
                            size: 18, color: SevaColors.info),
                        const SizedBox(width: 8),
                        Text(
                          'Upload Guidelines',
                          style:
                              Theme.of(context).textTheme.titleSmall?.copyWith(
                                    fontWeight: FontWeight.w600,
                                    color: SevaColors.info,
                                  ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    _guideline('Ensure the entire document is visible'),
                    _guideline('Photo should be clear and well-lit'),
                    _guideline('Avoid glare and reflections'),
                    _guideline('Name and ID number must be readable'),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Upload progress
              if (_isUploading)
                Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: Column(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(4),
                        child: LinearProgressIndicator(
                          value: _uploadProgress,
                          minHeight: 6,
                          backgroundColor: SevaColors.neutral200,
                          valueColor: const AlwaysStoppedAnimation<Color>(
                            SevaColors.primary,
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Uploading... ${(_uploadProgress * 100).toInt()}%',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: SevaColors.textTertiary,
                            ),
                      ),
                    ],
                  ),
                ),

              // Submit button
              SevaButton(
                label: 'Upload Documents',
                icon: Icons.cloud_upload,
                isLoading: _isUploading,
                onPressed: _frontImage != null && !_isUploading
                    ? _uploadDocuments
                    : null,
              ),
            ],

            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  Widget _guideline(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.only(top: 4),
            child: Icon(Icons.check_circle, size: 14, color: SevaColors.info),
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Text(
              text,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: SevaColors.textSecondary),
            ),
          ),
        ],
      ),
    );
  }
}

class _DocumentUploadArea extends StatelessWidget {
  final File? image;
  final String label;
  final VoidCallback onTap;

  const _DocumentUploadArea({
    this.image,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 180,
        width: double.infinity,
        decoration: BoxDecoration(
          color: SevaColors.neutral100,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: SevaColors.neutral300,
            style: image == null ? BorderStyle.none : BorderStyle.solid,
          ),
        ),
        clipBehavior: Clip.antiAlias,
        child: image != null
            ? Stack(
                fit: StackFit.expand,
                children: [
                  Image.file(image!, fit: BoxFit.cover),
                  Positioned(
                    top: 8,
                    right: 8,
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.5),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.edit,
                        color: Colors.white,
                        size: 18,
                      ),
                    ),
                  ),
                ],
              )
            : Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.add_a_photo_outlined,
                    size: 40,
                    color: SevaColors.neutral400,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    label,
                    style: TextStyle(
                      color: SevaColors.neutral400,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
      ),
    );
  }
}

class _StatusBanner extends StatelessWidget {
  final KycStatus status;
  final String? rejectionReason;
  final VoidCallback? onReupload;

  const _StatusBanner({
    required this.status,
    this.rejectionReason,
    this.onReupload,
  });

  @override
  Widget build(BuildContext context) {
    Color bgColor;
    Color textColor;
    IconData icon;
    String title;
    String subtitle;

    switch (status) {
      case KycStatus.pending:
        bgColor = SevaColors.warningLight;
        textColor = SevaColors.warning;
        icon = Icons.hourglass_top;
        title = 'Verification Pending';
        subtitle =
            'Your documents have been submitted and are under review. This typically takes 1-2 business days.';
      case KycStatus.verified:
        bgColor = SevaColors.successLight;
        textColor = SevaColors.success;
        icon = Icons.verified;
        title = 'Verified';
        subtitle =
            'Your identity has been verified successfully. You can now receive jobs and payouts.';
      case KycStatus.rejected:
        bgColor = SevaColors.errorLight;
        textColor = SevaColors.error;
        icon = Icons.error_outline;
        title = 'Verification Rejected';
        subtitle =
            rejectionReason ?? 'Your documents were rejected. Please re-upload.';
      case KycStatus.notStarted:
        return const SizedBox.shrink();
    }

    return SevaCard(
      backgroundColor: bgColor,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: textColor, size: 24),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        color: textColor,
                        fontWeight: FontWeight.w700,
                      ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: textColor.withValues(alpha: 0.8),
                ),
          ),
          if (onReupload != null) ...[
            const SizedBox(height: 12),
            SevaButton(
              label: 'Re-upload Documents',
              variant: SevaButtonVariant.outline,
              size: SevaButtonSize.small,
              isFullWidth: false,
              onPressed: onReupload,
            ),
          ],
        ],
      ),
    );
  }
}
