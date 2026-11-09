import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';

import '../controller/auth_controller.dart';

class ForgotPasswordScreen extends StatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  State<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends State<ForgotPasswordScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailFieldKey = GlobalKey<FormFieldState<String>>();
  final _emailController = TextEditingController();
  final _codeController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _authController = Get.find<AuthController>();

  Timer? _resendTimer;
  int _resendSeconds = 0;
  bool _codeRequested = false;

  @override
  void initState() {
    super.initState();
    _authController.clearError();
  }

  @override
  void dispose() {
    _resendTimer?.cancel();
    _emailController.dispose();
    _codeController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  Future<void> _requestCode() async {
    if (!(_emailFieldKey.currentState?.validate() ?? false)) return;

    final sent = await _authController.requestPasswordReset(
      email: _emailController.text,
    );
    if (!mounted || !sent) return;

    setState(() => _codeRequested = true);
    _startResendTimer();
  }

  void _startResendTimer() {
    _resendTimer?.cancel();
    setState(() => _resendSeconds = 60);
    _resendTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      if (_resendSeconds <= 1) {
        timer.cancel();
        setState(() => _resendSeconds = 0);
        return;
      }
      setState(() => _resendSeconds--);
    });
  }

  void _changeEmail() {
    _resendTimer?.cancel();
    _authController.clearError();
    _codeController.clear();
    _passwordController.clear();
    _confirmPasswordController.clear();
    setState(() {
      _codeRequested = false;
      _resendSeconds = 0;
    });
  }

  Future<void> _resetPassword() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;

    final reset = await _authController.resetPassword(
      email: _emailController.text,
      code: _codeController.text,
      newPassword: _passwordController.text,
    );
    if (!mounted || !reset) return;

    Get.offAllNamed('/login');
    Get.snackbar(
      'Password updated',
      'Sign in with your new password.',
      snackPosition: SnackPosition.BOTTOM,
    );
  }

  String? _validatePassword(String? value) {
    final password = value ?? '';
    final byteLength = utf8.encode(password).length;
    if (byteLength < 8) return 'Use at least 8 bytes';
    if (byteLength > 72) return 'Password must be 72 bytes or fewer';
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Forgot password')),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Icon(
                      _codeRequested
                          ? Icons.mark_email_read_outlined
                          : Icons.lock_reset_outlined,
                      size: 56,
                      color: colorScheme.primary,
                    ),
                    const SizedBox(height: 20),
                    Text(
                      _codeRequested ? 'Check your email' : 'Reset your password',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _codeRequested
                          ? 'If an account exists for this email, an 8-digit code was sent. The code expires in 15 minutes.'
                          : 'Enter the email address linked to your account. We will send a reset code if it exists.',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 28),
                    TextFormField(
                      key: _emailFieldKey,
                      controller: _emailController,
                      readOnly: _codeRequested,
                      keyboardType: TextInputType.emailAddress,
                      autofillHints: const [AutofillHints.email],
                      decoration: const InputDecoration(
                        labelText: 'Email',
                        prefixIcon: Icon(Icons.email_outlined),
                        border: OutlineInputBorder(),
                      ),
                      validator: (value) {
                        final email = value?.trim() ?? '';
                        if (email.isEmpty) return 'Please enter your email';
                        if (!GetUtils.isEmail(email)) {
                          return 'Please enter a valid email';
                        }
                        return null;
                      },
                    ),
                    if (_codeRequested) ...[
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _codeController,
                        keyboardType: TextInputType.number,
                        textInputAction: TextInputAction.next,
                        maxLength: 8,
                        inputFormatters: [
                          FilteringTextInputFormatter.digitsOnly,
                        ],
                        decoration: const InputDecoration(
                          labelText: '8-digit reset code',
                          prefixIcon: Icon(Icons.password_outlined),
                          border: OutlineInputBorder(),
                          counterText: '',
                        ),
                        validator: (value) {
                          if ((value ?? '').trim().length != 8) {
                            return 'Enter the 8-digit code from your email';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _passwordController,
                        obscureText: true,
                        textInputAction: TextInputAction.next,
                        decoration: const InputDecoration(
                          labelText: 'New password',
                          prefixIcon: Icon(Icons.lock_outline),
                          border: OutlineInputBorder(),
                        ),
                        validator: _validatePassword,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _confirmPasswordController,
                        obscureText: true,
                        textInputAction: TextInputAction.done,
                        onFieldSubmitted: (_) => _resetPassword(),
                        decoration: const InputDecoration(
                          labelText: 'Confirm new password',
                          prefixIcon: Icon(Icons.lock_outline),
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) {
                          if (value != _passwordController.text) {
                            return 'Passwords do not match';
                          }
                          return null;
                        },
                      ),
                    ],
                    Obx(() {
                      final error = _authController.errorMessage.value;
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
                    Obx(() {
                      final isLoading = _authController.isLoading.value;

                      return FilledButton(
                        onPressed: isLoading
                            ? null
                            : _codeRequested
                                ? _resetPassword
                                : _requestCode,
                        style: FilledButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                        ),
                        child: isLoading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : Text(
                                _codeRequested
                                    ? 'Reset password'
                                    : 'Send reset code',
                              ),
                      );
                    }),
                    if (_codeRequested) ...[
                      const SizedBox(height: 8),
                      TextButton(
                        onPressed: _resendSeconds > 0
                            ? null
                            : _requestCode,
                        child: Text(
                          _resendSeconds > 0
                              ? 'Resend code in ${_resendSeconds}s'
                              : 'Resend code',
                        ),
                      ),
                      TextButton(
                        onPressed: _changeEmail,
                        child: const Text('Use a different email'),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
