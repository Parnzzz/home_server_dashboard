import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:image_picker/image_picker.dart';

import '../../auth/controller/auth_controller.dart';
import '../../auth/model/auth_user.dart';
import '../controller/profile_controller.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final _formKey = GlobalKey<FormState>();
  final _displayNameController = TextEditingController();
  final _authController = Get.find<AuthController>();
  final _profileController = Get.find<ProfileController>();
  final _imagePicker = ImagePicker();

  Uint8List? _selectedAvatarBytes;
  bool _isEditing = false;

  @override
  void dispose() {
    _displayNameController.dispose();
    super.dispose();
  }

  void _beginEditing(UserProfile user) {
    _displayNameController.text = user.displayName;
    _selectedAvatarBytes = null;
    _profileController.errorMessage.value = null;
    setState(() => _isEditing = true);
  }

  void _cancelEditing() {
    _profileController.errorMessage.value = null;
    setState(() {
      _selectedAvatarBytes = null;
      _isEditing = false;
    });
  }

  Future<void> _pickAvatar() async {
    try {
      final image = await _imagePicker.pickImage(
        source: ImageSource.gallery,
        imageQuality: 80,
        maxWidth: 512,
        maxHeight: 512,
      );
      if (image == null) return;

      final bytes = await image.readAsBytes();
      if (!mounted || bytes.isEmpty) return;

      if (!_isEditing) {
        _displayNameController.text =
            _authController.user.value?.displayName ?? '';
      }

      _profileController.errorMessage.value = null;
      setState(() {
        _selectedAvatarBytes = bytes;
        _isEditing = true;
      });
    } catch (_) {
      if (!mounted) return;
      Get.snackbar(
        'Unable to choose photo',
        'Please try selecting the photo again.',
        snackPosition: SnackPosition.BOTTOM,
      );
    }
  }

  Future<void> _saveProfile() async {
    if (!_formKey.currentState!.validate()) return;

    final avatarUrl = _selectedAvatarBytes == null
        ? null
        : _avatarDataUri(_selectedAvatarBytes!);
    final updated = await _profileController.updateProfile(
      displayName: _displayNameController.text.trim(),
      avatarUrl: avatarUrl,
    );

    if (!mounted || !updated) return;

    setState(() {
      _selectedAvatarBytes = null;
      _isEditing = false;
    });
    Get.snackbar(
      'Profile updated',
      'Your changes have been saved.',
      snackPosition: SnackPosition.BOTTOM,
    );
  }

  String _avatarDataUri(Uint8List bytes) {
    var mimeType = 'image/jpeg';
    if (bytes.length >= 8 &&
        bytes[0] == 0x89 &&
        bytes[1] == 0x50 &&
        bytes[2] == 0x4e &&
        bytes[3] == 0x47) {
      mimeType = 'image/png';
    } else if (bytes.length >= 6 &&
        String.fromCharCodes(bytes.sublist(0, 3)) == 'GIF') {
      mimeType = 'image/gif';
    } else if (bytes.length >= 12 &&
        String.fromCharCodes(bytes.sublist(0, 4)) == 'RIFF' &&
        String.fromCharCodes(bytes.sublist(8, 12)) == 'WEBP') {
      mimeType = 'image/webp';
    }

    return 'data:$mimeType;base64,${base64Encode(bytes)}';
  }

  ImageProvider? _avatarImage(String? avatarUrl) {
    if (_selectedAvatarBytes != null) {
      return MemoryImage(_selectedAvatarBytes!);
    }
    if (avatarUrl == null || avatarUrl.isEmpty) return null;

    if (avatarUrl.startsWith('data:image/')) {
      final separator = avatarUrl.indexOf(',');
      if (separator < 0) return null;
      try {
        return MemoryImage(base64Decode(avatarUrl.substring(separator + 1)));
      } on FormatException {
        return null;
      }
    }

    return NetworkImage(avatarUrl);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: Obx(() {
        final user = _authController.user.value;

        if (user == null) {
          return const Center(child: Text('User profile not found'));
        }
        final avatarImage = _avatarImage(user.avatarUrl);

        return Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 500),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Center(
                    child: Semantics(
                      button: true,
                      label: 'Choose profile photo',
                      child: InkWell(
                        customBorder: const CircleBorder(),
                        onTap: _pickAvatar,
                        child: Stack(
                          clipBehavior: Clip.none,
                          children: [
                            CircleAvatar(
                              radius: 52,
                              backgroundColor:
                                  colorScheme.surfaceContainerHighest,
                              backgroundImage: avatarImage,
                              child: avatarImage == null
                                  ? Icon(
                                      Icons.person,
                                      size: 52,
                                      color: colorScheme.onSurfaceVariant,
                                    )
                                  : null,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),
                  if (_isEditing)
                    _buildEditForm(context)
                  else ...[
                    Card(
                      child: Column(
                        children: [
                          ListTile(
                            leading: const Icon(Icons.person_outline),
                            title: const Text('Display Name'),
                            subtitle: Text(user.displayName),
                          ),
                          const Divider(height: 1),
                          ListTile(
                            leading: const Icon(Icons.email_outlined),
                            title: const Text('Email'),
                            subtitle: Text(user.email),
                          ),
                          const Divider(height: 1),
                          ListTile(
                            leading: const Icon(
                              Icons.admin_panel_settings_outlined,
                            ),
                            title: const Text('Role'),
                            subtitle: Text(user.role.toUpperCase()),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 20),
                    FilledButton.icon(
                      onPressed: () => _beginEditing(user),
                      icon: const Icon(Icons.edit_outlined),
                      label: const Text('Edit profile'),
                    ),
                  ],
                ],
              ),
            ),
          ),
        );
      }),
    );
  }

  Widget _buildEditForm(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          TextFormField(
            controller: _displayNameController,
            textCapitalization: TextCapitalization.words,
            maxLength: 100,
            decoration: const InputDecoration(
              labelText: 'Display name',
              prefixIcon: Icon(Icons.person_outline),
              border: OutlineInputBorder(),
              counterText: '',
            ),
            validator: (value) {
              final name = value?.trim() ?? '';
              if (name.isEmpty) return 'Please enter a display name';
              if (name.length > 100) {
                return 'Name must be 100 characters or fewer';
              }
              return null;
            },
          ),
          const SizedBox(height: 16),
          Text(
            _selectedAvatarBytes == null
                ? 'Tap the photo above to choose a new avatar.'
                : 'New avatar selected. Save changes to apply it.',
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: colorScheme.onSurfaceVariant,
            ),
          ),
          Obx(() {
            final error = _profileController.errorMessage.value;
            if (error == null) return const SizedBox(height: 16);

            return Padding(
              padding: const EdgeInsets.only(top: 16),
              child: Text(
                error,
                textAlign: TextAlign.center,
                style: TextStyle(color: colorScheme.error),
              ),
            );
          }),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: Obx(() {
                  final isLoading = _profileController.isLoading.value;

                  return OutlinedButton(
                    onPressed: isLoading ? null : _cancelEditing,
                    child: const Text('Cancel'),
                  );
                }),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Obx(() {
                  final isLoading = _profileController.isLoading.value;

                  return FilledButton.icon(
                    onPressed: isLoading ? null : _saveProfile,
                    icon: isLoading
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.save_outlined),
                    label: Text(isLoading ? 'Saving...' : 'Save'),
                  );
                }),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
