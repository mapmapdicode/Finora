import 'package:flutter/material.dart';
import 'package:mobile/app/app_dependencies.dart';
import 'package:mobile/core/theme/finora_theme.dart';
import 'package:mobile/core/theme/theme_controller.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';
import 'package:mobile/features/finora/presentation/finora_pages.dart';

/// Root composition widget. It owns dependencies shared for one app session.
class FinoraApp extends StatefulWidget {
  const FinoraApp({super.key, this.dependencies});

  final AppDependencies? dependencies;

  @override
  State<FinoraApp> createState() => _FinoraAppState();
}

class _FinoraAppState extends State<FinoraApp> {
  late final AppDependencies _dependencies =
      widget.dependencies ?? AppDependencies.production();
  late final LoginViewModel _loginViewModel = LoginViewModel(
    _dependencies.authRepository,
  );
  late final FinoraThemeController _themeController = FinoraThemeController();

  @override
  void initState() {
    super.initState();
    _themeController.restore();
  }

  @override
  void dispose() {
    _loginViewModel.dispose();
    _themeController.dispose();
    super.dispose();
  }

  Widget _buildLogin(BuildContext context) =>
      LoginPage(viewModel: _loginViewModel, homeBuilder: _buildHome);

  Widget _buildHome(BuildContext context) =>
      HomePage(api: _dependencies.apiClient, loginBuilder: _buildLogin);

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: _themeController,
      builder: (context, child) {
        return FinoraThemeScope(
          controller: _themeController,
          child: MaterialApp(
            title: 'Finora',
            debugShowCheckedModeBanner: false,
            theme: FinoraTheme.light,
            darkTheme: FinoraTheme.dark,
            themeMode: _themeController.mode,
            home: _buildLogin(context),
          ),
        );
      },
    );
  }
}
