import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/providers.dart';
import '../../core/errors/api_exception.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});
  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _form = GlobalKey<FormState>();
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _otp = TextEditingController();
  bool _loading = false;
  bool _needsOtp = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _otp.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(authRepositoryProvider).login(
          email: _email.text, password: _password.text, otpCode: _otp.text);
      ref.invalidate(currentAccountProvider);
      ref.invalidate(companiesProvider);
      ref.invalidate(selectedCompanyProvider);
      if (mounted) {
        context.go('/companies');
      }
    } on ApiException catch (error) {
      if (mounted) {
        setState(() {
          _needsOtp = error.isTwoFactorRequired;
          _error = error.message;
        });
      }
    } finally {
      if (mounted) {
        setState(() => _loading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        body: SafeArea(
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 460),
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Form(
                  key: _form,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Image.asset('assets/branding/pcs_app_icon.png',
                          width: 76, height: 76),
                      const SizedBox(height: 18),
                      Text('Powerful Control System',
                          textAlign: TextAlign.center,
                          style: Theme.of(context).textTheme.headlineSmall),
                      const SizedBox(height: 24),
                      TextFormField(
                          controller: _email,
                          keyboardType: TextInputType.emailAddress,
                          autocorrect: false,
                          decoration:
                              const InputDecoration(labelText: 'Correo'),
                          validator: (value) => (value ?? '').contains('@')
                              ? null
                              : 'Ingresa un correo válido.'),
                      const SizedBox(height: 12),
                      TextFormField(
                          controller: _password,
                          obscureText: true,
                          decoration:
                              const InputDecoration(labelText: 'Contraseña'),
                          validator: (value) => (value ?? '').length >= 8
                              ? null
                              : 'Ingresa tu contraseña.'),
                      if (_needsOtp) ...[
                        const SizedBox(height: 12),
                        TextFormField(
                            controller: _otp,
                            keyboardType: TextInputType.number,
                            decoration: const InputDecoration(
                                labelText: 'Código de segundo factor'),
                            validator: (value) =>
                                _needsOtp && (value ?? '').isEmpty
                                    ? 'Ingresa el código.'
                                    : null)
                      ],
                      if (_error != null)
                        Padding(
                            padding: const EdgeInsets.only(top: 12),
                            child: Text(_error!,
                                style: TextStyle(
                                    color:
                                        Theme.of(context).colorScheme.error))),
                      const SizedBox(height: 18),
                      FilledButton(
                          onPressed: _loading ? null : _submit,
                          child: Text(
                              _loading ? 'Validando...' : 'Iniciar sesión')),
                      TextButton(
                          onPressed: () {},
                          child: const Text(
                              'Recuperar contraseña desde el portal seguro')),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );
}
