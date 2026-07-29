import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/utils/vietnamese_money_input.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/core/theme/finora_tokens.dart';
import 'package:mobile/core/theme/finora_typography.dart';
import 'package:mobile/core/widgets/finora_core_widgets.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';
import 'package:mobile/features/loans/data/repositories/loan_repository_impl.dart';
import 'package:mobile/features/loans/data/services/loan_remote_service.dart';
import 'package:mobile/features/loans/presentation/screens/loan_page.dart';
import 'package:mobile/features/loans/presentation/view_models/loan_view_model.dart';
import 'package:webview_flutter/webview_flutter.dart';

part 'screens/scenario_pages.dart';
part 'screens/resource_support_pages.dart';
part 'models/presentation_models.dart';

abstract final class _I18n {
  static const Map<String, Map<String, String>> _strings = {
    'VN': {
      'subtitle': 'Quản lý Chi tiêu & Đầu tư Cá nhân',
      'bannerTitle': 'TÀI SẢN & DÒNG TIỀN',
      'bannerSub': 'Quản lý thu chi • Tăng trưởng danh mục đầu tư',
      'greeting': 'Chúc bạn một ngày tốt lành',
      'newAccount': 'TẠO TÀI KHOẢN MỚI',
      'loginBtn': 'Đăng nhập',
      'registerBtn': 'Tạo tài khoản',
      'emailLabel': 'Email đăng nhập',
      'passLabel': 'Mật khẩu',
      'confirmPassLabel': 'Nhập lại mật khẩu',
      'nameLabel': 'Họ và tên',
      'switchRegister': 'Chưa có tài khoản? Tạo tài khoản',
      'switchLogin': 'Đã có tài khoản? Đăng nhập',
      'netWorth': 'Tài sản ròng',
      'aiAssistant': 'Trợ lý AI',
      'security': 'Bảo mật 100%',
      'logExpense': 'Ghi thu chi',
      'portfolio': 'Danh mục',
      'budget': 'Ngân sách',
      'forecast': 'Dự báo',
      'notifications': 'Thông báo & Cảnh báo',
      'markAllRead': 'Đánh dấu đã đọc',
      'langTitle': 'Chọn Ngôn Ngữ',
    },
    'EN': {
      'subtitle': 'Personal Finance & Wealth Management',
      'bannerTitle': 'ASSETS & CASH FLOW',
      'bannerSub': 'Expense tracking • Investment portfolio growth',
      'greeting': 'Have a wonderful day',
      'newAccount': 'CREATE NEW ACCOUNT',
      'loginBtn': 'Sign In',
      'registerBtn': 'Create Account',
      'emailLabel': 'Email Address',
      'passLabel': 'Password',
      'confirmPassLabel': 'Confirm password',
      'nameLabel': 'Full Name',
      'switchRegister': "Don't have an account? Sign Up",
      'switchLogin': 'Already have an account? Sign In',
      'netWorth': 'Net Worth',
      'aiAssistant': 'AI Assistant',
      'security': '100% Secure',
      'logExpense': 'Log Expense',
      'portfolio': 'Portfolio',
      'budget': 'Budget',
      'forecast': 'Forecast',
      'notifications': 'Notifications & Alerts',
      'markAllRead': 'Mark all as read',
      'langTitle': 'Select Language',
    },
    'JP': {
      'subtitle': '個人財務および投資管理',
      'bannerTitle': '資産とキャッシュフロー',
      'bannerSub': '収支管理 • ポートフォリオ成長',
      'greeting': '良い一日をお過ごしください',
      'newAccount': '新規アカウント作成',
      'loginBtn': 'ログイン',
      'registerBtn': 'アカウント作成',
      'emailLabel': 'メールアドレス',
      'passLabel': 'パスワード',
      'confirmPassLabel': 'パスワードを再入力',
      'nameLabel': '氏名',
      'switchRegister': 'アカウントをお持ちでない方',
      'switchLogin': '既にアカウントをお持ちの方',
      'netWorth': '純資産',
      'aiAssistant': 'AIアシスタント',
      'security': '100% セキュリティ',
      'logExpense': '収支記録',
      'portfolio': 'ポートフォリオ',
      'budget': '予算',
      'forecast': '予測',
      'notifications': '通知とアラート',
      'markAllRead': 'すべて既読にする',
      'langTitle': '言語を選択',
    },
    'KR': {
      'subtitle': '개인 자산 및 투자 관리',
      'bannerTitle': '자산 및 현금 흐름',
      'bannerSub': '지출 추적 • 포트폴리오 성과',
      'greeting': '좋은 하루 보내세요',
      'newAccount': '새 계정 만들기',
      'loginBtn': '로그인',
      'registerBtn': '계정 만들기',
      'emailLabel': '이메일 주소',
      'passLabel': '비밀번호',
      'confirmPassLabel': '비밀번호 확인',
      'nameLabel': '이름',
      'switchRegister': '계정이 없으신가요? 가입하기',
      'switchLogin': '이미 계정이 있으신가요? 로그인',
      'netWorth': '순자산',
      'aiAssistant': 'AI 어시스턴트',
      'security': '100% 보안',
      'logExpense': '지출 기록',
      'portfolio': '포트폴리오',
      'budget': '예산',
      'forecast': '예측',
      'notifications': '알림 및 경고',
      'markAllRead': '모두 읽음으로 표시',
      'langTitle': '언어 선택',
    },
  };

  static String t(String lang, String key) {
    return _strings[lang]?[key] ?? _strings['VN']![key] ?? key;
  }

  static String getFlag(String code) {
    switch (code) {
      case 'EN':
        return '🇺🇸';
      case 'JP':
        return '🇯🇵';
      case 'KR':
        return '🇰🇷';
      default:
        return '🇻🇳';
    }
  }
}

class LoginPage extends StatefulWidget {
  const LoginPage({
    super.key,
    required this.viewModel,
    required this.homeBuilder,
  });
  final LoginViewModel viewModel;
  final WidgetBuilder homeBuilder;
  @override
  State<LoginPage> createState() => _LoginPageState();
}

String appAmountDisplayMode = 'full';

class _LoginPageState extends State<LoginPage> with TickerProviderStateMixin {
  // Local development account seeded by APP_SEED_DEMO for fast UI testing.
  final email = TextEditingController(text: 'thanhoangz');
  final password = TextEditingController(text: 'HoangThanZ6^');
  final confirmPassword = TextEditingController();
  final verificationCode = TextEditingController();
  final name = TextEditingController();
  bool registering = false;
  bool showingEmailVerification = false;
  bool obscurePassword = true;
  bool obscureConfirmPassword = true;
  String? registrationMessage;
  String currentLang = 'VN';
  // Notifications are not backed by a feed yet; never signal unread data
  // until the API supplies a real notification state.
  bool hasUnreadNotifications = false;

  late final AnimationController _entranceController;
  late final AnimationController _pulseController;

  late final Animation<double> _headerFade;
  late final Animation<Offset> _headerSlide;
  late final Animation<double> _formFade;
  late final Animation<Offset> _formSlide;
  late final Animation<double> _bottomNavFade;
  late final Animation<Offset> _bottomNavSlide;
  late final Animation<double> _bgScale;

  @override
  void initState() {
    super.initState();
    widget.viewModel.addListener(_onViewModelChanged);

    _entranceController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1100),
    );

    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 10),
    )..repeat(reverse: true);

    _bgScale = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );

    _headerFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.0, 0.5, curve: Curves.easeOut),
    );
    _headerSlide = Tween<Offset>(begin: const Offset(0, -0.2), end: Offset.zero)
        .animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.0, 0.6, curve: Curves.easeOutCubic),
          ),
        );

    _formFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.2, 0.7, curve: Curves.easeOut),
    );
    _formSlide = Tween<Offset>(begin: const Offset(0, 0.15), end: Offset.zero)
        .animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.2, 0.8, curve: Curves.easeOutCubic),
          ),
        );

    _bottomNavFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.45, 0.9, curve: Curves.easeOut),
    );
    _bottomNavSlide =
        Tween<Offset>(begin: const Offset(0, 0.25), end: Offset.zero).animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.45, 0.95, curve: Curves.easeOutCubic),
          ),
        );

    _entranceController.forward();
  }

  void _onViewModelChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> submit() async {
    final wasRegistering = registering;
    final authenticated = await widget.viewModel.authenticate(
      registering: registering,
      email: email.text,
      password: password.text,
      confirmPassword: confirmPassword.text,
      name: name.text,
    );
    if (!authenticated &&
        widget.viewModel.pendingVerificationEmail != null &&
        mounted) {
      setState(() {
        showingEmailVerification = true;
        if (wasRegistering) {
          registering = false;
          password.clear();
          confirmPassword.clear();
          name.clear();
          obscurePassword = true;
          obscureConfirmPassword = true;
        }
        registrationMessage = null;
      });
      return;
    }
    if (authenticated && mounted) {
      Navigator.of(
        context,
      ).pushReplacement(MaterialPageRoute(builder: widget.homeBuilder));
    }
  }

  Future<void> verifyEmail() async {
    final verified = await widget.viewModel.verifyEmail(verificationCode.text);
    if (verified && mounted) {
      Navigator.of(
        context,
      ).pushReplacement(MaterialPageRoute(builder: widget.homeBuilder));
    }
  }

  void _showAppSettingsSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: FinoraColors.surfaceElevated,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.settings_rounded,
                        color: FinoraColors.primary,
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      const Text(
                        'Cách hiển thị số tiền',
                        style: FinoraTypography.h3,
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(
                          Icons.close_rounded,
                          color: FinoraColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Chọn cách hiển thị số tiền trên toàn bộ ứng dụng.',
                    style: FinoraTypography.bodySmall.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildDisplayModeOption(
                    title: 'Viết tắt / Rút gọn (100)',
                    subtitle: 'Ví dụ: 100.000 VND hiển thị thành 100',
                    value: 'compact',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 10),
                  _buildDisplayModeOption(
                    title: 'Đầy đủ (100.000 VND)',
                    subtitle: 'Hiển thị nguyên văn giá trị gốc đầy đủ',
                    value: 'full',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 16),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildDisplayModeOption({
    required String title,
    required String subtitle,
    required String value,
    required StateSetter setModalState,
  }) {
    final isSelected = appAmountDisplayMode == value;
    return InkWell(
      onTap: () {
        setState(() {
          appAmountDisplayMode = value;
        });
        setModalState(() {});
      },
      borderRadius: BorderRadius.circular(16),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected ? FinoraColors.primarySoft : FinoraColors.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected ? FinoraColors.primary : FinoraColors.border,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: FinoraTypography.title.copyWith(
                      color: isSelected
                          ? FinoraColors.primaryDeep
                          : FinoraColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: FinoraTypography.caption.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              const Icon(
                Icons.check_circle_rounded,
                color: FinoraColors.primary,
                size: 22,
              ),
          ],
        ),
      ),
    );
  }

  void _showLanguageSelector() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (context) {
        final langs = [
          {'code': 'VN', 'flag': '🇻🇳', 'name': 'Tiếng Việt'},
          {'code': 'EN', 'flag': '🇺🇸', 'name': 'English'},
          {'code': 'JP', 'flag': '🇯🇵', 'name': '日本語'},
          {'code': 'KR', 'flag': '🇰🇷', 'name': '한국어'},
        ];
        return Container(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(
                    Icons.translate_rounded,
                    color: Color(0xfffbbf24),
                    size: 22,
                  ),
                  const SizedBox(width: 10),
                  Text(
                    _I18n.t(currentLang, 'langTitle'),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(
                      Icons.close_rounded,
                      color: Colors.white70,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              ...langs.map((l) {
                final isSelected = l['code'] == currentLang;
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: InkWell(
                    onTap: () {
                      setState(() => currentLang = l['code']!);
                      Navigator.pop(context);
                    },
                    borderRadius: BorderRadius.circular(16),
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 250),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0x33fbbf24)
                            : Colors.white.withValues(alpha: 0.08),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: isSelected
                              ? const Color(0xfffbbf24)
                              : Colors.white.withValues(alpha: 0.12),
                          width: isSelected ? 1.5 : 1,
                        ),
                      ),
                      child: Row(
                        children: [
                          Text(
                            l['flag']!,
                            style: const TextStyle(fontSize: 22),
                          ),
                          const SizedBox(width: 14),
                          Text(
                            l['name']!,
                            style: TextStyle(
                              color: isSelected
                                  ? const Color(0xfffbbf24)
                                  : Colors.white,
                              fontSize: 15,
                              fontWeight: isSelected
                                  ? FontWeight.bold
                                  : FontWeight.w600,
                            ),
                          ),
                          const Spacer(),
                          if (isSelected)
                            const Icon(
                              Icons.check_circle_rounded,
                              color: Color(0xfffbbf24),
                              size: 20,
                            ),
                        ],
                      ),
                    ),
                  ),
                );
              }),
              const SizedBox(height: 12),
            ],
          ),
        );
      },
    );
  }

  void _showNotificationSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: FinoraColors.surfaceElevated,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(24, 16, 24, 28),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(
                      Icons.notifications_none_rounded,
                      color: FinoraColors.primary,
                      size: 22,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        _I18n.t(currentLang, 'notifications'),
                        style: FinoraTypography.h3,
                      ),
                    ),
                    IconButton(
                      tooltip: 'Đóng thông báo',
                      onPressed: () => Navigator.pop(context),
                      icon: const Icon(
                        Icons.close_rounded,
                        color: FinoraColors.textSecondary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 18),
                const FinoraEmptyState(
                  icon: Icons.notifications_none_rounded,
                  title: 'Chưa có thông báo',
                  message:
                      'Thông báo bảo mật và hoạt động sẽ xuất hiện ở đây khi được đồng bộ.',
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  void dispose() {
    _entranceController.dispose();
    _pulseController.dispose();
    widget.viewModel.removeListener(_onViewModelChanged);
    email.dispose();
    password.dispose();
    confirmPassword.dispose();
    verificationCode.dispose();
    name.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: FinoraColors.background,
    body: Stack(
      children: [
        Positioned.fill(
          child: AnimatedBuilder(
            animation: _bgScale,
            builder: (context, child) {
              return Transform.scale(
                scale: _bgScale.value,
                child: Image.asset(
                  'assets/images/login_bg.png',
                  fit: BoxFit.cover,
                  alignment: Alignment.center,
                ),
              );
            },
          ),
        ),
        Positioned.fill(
          child: Container(
            decoration: const BoxDecoration(gradient: FinoraColors.bgGradient),
          ),
        ),
        SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final wide = constraints.maxWidth >= 760;
              final Widget form = showingEmailVerification
                  ? _EmailVerificationForm(
                      key: ValueKey(
                        widget.viewModel.pendingVerificationEmail ?? email.text,
                      ),
                      email:
                          widget.viewModel.pendingVerificationEmail ??
                          email.text,
                      code: verificationCode,
                      busy: widget.viewModel.isBusy,
                      error: widget.viewModel.error,
                      onVerify: verifyEmail,
                      onResend: widget.viewModel.resendVerificationEmail,
                      onBack: () {
                        setState(() {
                          showingEmailVerification = false;
                          verificationCode.clear();
                        });
                        widget.viewModel.cancelEmailVerification();
                      },
                    )
                  : _LoginForm(
                      registering: registering,
                      busy: widget.viewModel.isBusy,
                      error: widget.viewModel.error,
                      message: registrationMessage,
                      email: email,
                      password: password,
                      confirmPassword: confirmPassword,
                      name: name,
                      obscurePassword: obscurePassword,
                      obscureConfirmPassword: obscureConfirmPassword,
                      lang: currentLang,
                      onTogglePassword: () =>
                          setState(() => obscurePassword = !obscurePassword),
                      onToggleConfirmPassword: () => setState(
                        () => obscureConfirmPassword = !obscureConfirmPassword,
                      ),
                      onSubmit: submit,
                      onOpenEmailVerification:
                          widget.viewModel.pendingVerificationEmail == null
                          ? null
                          : () => setState(() {
                              showingEmailVerification = true;
                              registrationMessage = null;
                            }),
                      onSwitch: () {
                        setState(() {
                          registering = !registering;
                          showingEmailVerification = false;
                          confirmPassword.clear();
                          obscureConfirmPassword = true;
                          registrationMessage = null;
                        });
                        widget.viewModel.clearError();
                      },
                    );

              if (wide) {
                return Center(
                  child: SingleChildScrollView(
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 400),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        child: Column(
                          children: [
                            FadeTransition(
                              opacity: _headerFade,
                              child: SlideTransition(
                                position: _headerSlide,
                                child: _LoginHeader(
                                  lang: currentLang,
                                  hasUnread: hasUnreadNotifications,
                                  onSelectLang: _showLanguageSelector,
                                  onOpenNotifications: _showNotificationSheet,
                                  onOpenSettings: _showAppSettingsSheet,
                                ),
                              ),
                            ),
                            const SizedBox(height: 12),
                            FadeTransition(
                              opacity: _formFade,
                              child: SlideTransition(
                                position: _formSlide,
                                child: form,
                              ),
                            ),
                            const SizedBox(height: 12),
                            FadeTransition(
                              opacity: _bottomNavFade,
                              child: SlideTransition(
                                position: _bottomNavSlide,
                                child: _LoginBottomNav(lang: currentLang),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              }

              return Column(
                children: [
                  FadeTransition(
                    opacity: _headerFade,
                    child: SlideTransition(
                      position: _headerSlide,
                      child: _LoginHeader(
                        lang: currentLang,
                        hasUnread: hasUnreadNotifications,
                        onSelectLang: _showLanguageSelector,
                        onOpenNotifications: _showNotificationSheet,
                        onOpenSettings: _showAppSettingsSheet,
                      ),
                    ),
                  ),
                  Expanded(
                    child: LayoutBuilder(
                      builder: (context, constraints) {
                        return SingleChildScrollView(
                          physics: const BouncingScrollPhysics(),
                          padding: const EdgeInsets.symmetric(horizontal: 14),
                          child: ConstrainedBox(
                            constraints: BoxConstraints(
                              minHeight: constraints.maxHeight,
                            ),
                            child: Center(
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 8,
                                ),
                                child: FadeTransition(
                                  opacity: _formFade,
                                  child: SlideTransition(
                                    position: _formSlide,
                                    child: form,
                                  ),
                                ),
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  FadeTransition(
                    opacity: _bottomNavFade,
                    child: SlideTransition(
                      position: _bottomNavSlide,
                      child: _LoginBottomNav(lang: currentLang),
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ],
    ),
  );
}

class _LoginHeader extends StatelessWidget {
  const _LoginHeader({
    required this.lang,
    required this.hasUnread,
    required this.onSelectLang,
    required this.onOpenNotifications,
    this.onOpenSettings,
  });

  final String lang;
  final bool hasUnread;
  final VoidCallback onSelectLang;
  final VoidCallback onOpenNotifications;
  final VoidCallback? onOpenSettings;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 7),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.9),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: FinoraColors.border),
          boxShadow: [
            BoxShadow(
              color: FinoraColors.primary.withValues(alpha: 0.10),
              blurRadius: 14,
            ),
          ],
        ),
        child: Row(
          children: [
            const _BrandMark(size: 28),
            const SizedBox(width: 8),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    const Text(
                      'Finora',
                      style: TextStyle(
                        color: FinoraColors.textPrimary,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                        letterSpacing: -0.5,
                      ),
                    ),
                    const SizedBox(width: 5),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 5,
                        vertical: 1.5,
                      ),
                      decoration: BoxDecoration(
                        color: FinoraColors.primarySoft,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: const Text(
                        'WEALTH OS',
                        style: TextStyle(
                          color: FinoraColors.primary,
                          fontSize: 8,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.6,
                        ),
                      ),
                    ),
                  ],
                ),
                Text(
                  _I18n.t(lang, 'subtitle'),
                  style: TextStyle(
                    color: FinoraColors.textSecondary,
                    fontSize: 9,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
            const Spacer(),
            // Interactive Language Selector
            Semantics(
              button: true,
              label: 'Chọn ngôn ngữ, hiện tại $lang',
              child: InkWell(
                onTap: onSelectLang,
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: FinoraColors.primarySoft,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: FinoraColors.border),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        _I18n.getFlag(lang),
                        style: const TextStyle(fontSize: 11.5),
                      ),
                      const SizedBox(width: 3),
                      Text(
                        lang,
                        style: const TextStyle(
                          color: FinoraColors.textPrimary,
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(width: 2),
                      const Icon(
                        Icons.keyboard_arrow_down_rounded,
                        color: FinoraColors.textSecondary,
                        size: 14,
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(width: 6),
            // Interactive Notification Bell
            Semantics(
              button: true,
              label: hasUnread ? 'Thông báo, có thông báo mới' : 'Thông báo',
              child: InkWell(
                onTap: onOpenNotifications,
                borderRadius: BorderRadius.circular(99),
                child: Stack(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: FinoraColors.primarySoft,
                        shape: BoxShape.circle,
                        border: Border.all(color: FinoraColors.border),
                      ),
                      child: const Icon(
                        Icons.notifications_none_rounded,
                        color: FinoraColors.textPrimary,
                        size: 15,
                      ),
                    ),
                    if (hasUnread)
                      Positioned(
                        right: 2,
                        top: 2,
                        child: Container(
                          width: 7,
                          height: 7,
                          decoration: const BoxDecoration(
                            color: Color(0xffef4444),
                            shape: BoxShape.circle,
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ),
            if (onOpenSettings != null) ...[
              const SizedBox(width: 6),
              Semantics(
                button: true,
                label: 'Mở cài đặt',
                child: InkWell(
                  onTap: onOpenSettings,
                  borderRadius: BorderRadius.circular(99),
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: FinoraColors.primarySoft,
                      shape: BoxShape.circle,
                      border: Border.all(color: FinoraColors.border),
                    ),
                    child: const Icon(
                      Icons.settings_rounded,
                      color: FinoraColors.textPrimary,
                      size: 15,
                    ),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _LoginForm extends StatelessWidget {
  const _LoginForm({
    required this.registering,
    required this.busy,
    required this.error,
    required this.message,
    required this.email,
    required this.password,
    required this.confirmPassword,
    required this.name,
    required this.obscurePassword,
    required this.obscureConfirmPassword,
    required this.lang,
    required this.onTogglePassword,
    required this.onToggleConfirmPassword,
    required this.onSubmit,
    this.onOpenEmailVerification,
    required this.onSwitch,
  });

  final bool registering, busy, obscurePassword, obscureConfirmPassword;
  final String? error;
  final String? message;
  final TextEditingController email, password, confirmPassword, name;
  final String lang;
  final VoidCallback onTogglePassword,
      onToggleConfirmPassword,
      onSubmit,
      onSwitch;
  final VoidCallback? onOpenEmailVerification;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.94),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: FinoraColors.border),
        boxShadow: [
          BoxShadow(
            color: FinoraColors.primary.withValues(alpha: 0.12),
            blurRadius: 24,
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            _I18n.t(lang, 'greeting'),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              color: FinoraColors.textSecondary,
                              fontSize: 11.5,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        const SizedBox(width: 3),
                        const Text('👋', style: TextStyle(fontSize: 11.5)),
                      ],
                    ),
                    const SizedBox(height: 1),
                    ShaderMask(
                      shaderCallback: (bounds) => const LinearGradient(
                        colors: [FinoraColors.primary, FinoraColors.purple],
                      ).createShader(bounds),
                      child: Text(
                        registering
                            ? _I18n.t(lang, 'newAccount')
                            : (email.text.isNotEmpty
                                  ? email.text.split('@').first.toUpperCase()
                                  : 'FINORA'),
                        style: const TextStyle(
                          color: FinoraColors.primary,
                          fontSize: 17,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.4,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                decoration: BoxDecoration(
                  color: FinoraColors.primarySoft,
                  shape: BoxShape.circle,
                  border: Border.all(color: FinoraColors.border),
                ),
                child: IconButton(
                  constraints: const BoxConstraints(
                    minWidth: 34,
                    minHeight: 34,
                  ),
                  padding: EdgeInsets.zero,
                  onPressed: onSwitch,
                  tooltip: registering ? 'Quay lại đăng nhập' : 'Đổi tài khoản',
                  icon: Icon(
                    registering
                        ? Icons.login_rounded
                        : Icons.account_circle_outlined,
                    color: FinoraColors.primary,
                    size: 18,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),

          AnimatedSize(
            duration: const Duration(milliseconds: 350),
            curve: Curves.easeInOutCubic,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (registering) ...[
                  _CustomGlassTextField(
                    controller: name,
                    labelText: _I18n.t(lang, 'nameLabel'),
                    icon: Icons.person_outline_rounded,
                    textCapitalization: TextCapitalization.words,
                  ),
                  const SizedBox(height: 8),
                ],
                _CustomGlassTextField(
                  controller: email,
                  labelText: _I18n.t(lang, 'emailLabel'),
                  icon: Icons.alternate_email_rounded,
                  keyboardType: TextInputType.emailAddress,
                ),
                const SizedBox(height: 8),
                _CustomGlassTextField(
                  controller: password,
                  labelText: _I18n.t(lang, 'passLabel'),
                  icon: Icons.lock_outline_rounded,
                  obscureText: obscurePassword,
                  suffixIcon: IconButton(
                    onPressed: onTogglePassword,
                    tooltip: obscurePassword ? 'Hiện mật khẩu' : 'Ẩn mật khẩu',
                    icon: Icon(
                      obscurePassword
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                      color: FinoraColors.textSecondary,
                      size: 18,
                    ),
                  ),
                ),
                if (registering) ...[
                  const SizedBox(height: 8),
                  _CustomGlassTextField(
                    controller: confirmPassword,
                    labelText: _I18n.t(lang, 'confirmPassLabel'),
                    icon: Icons.lock_reset_outlined,
                    obscureText: obscureConfirmPassword,
                    suffixIcon: IconButton(
                      onPressed: onToggleConfirmPassword,
                      tooltip: obscureConfirmPassword
                          ? 'Hiện mật khẩu nhập lại'
                          : 'Ẩn mật khẩu nhập lại',
                      icon: Icon(
                        obscureConfirmPassword
                            ? Icons.visibility_outlined
                            : Icons.visibility_off_outlined,
                        color: FinoraColors.textSecondary,
                        size: 18,
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),

          if (message != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 11,
                  vertical: 7,
                ),
                decoration: BoxDecoration(
                  color: FinoraColors.primarySoft,
                  borderRadius: BorderRadius.circular(11),
                  border: Border.all(color: FinoraColors.primary),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(
                          Icons.mark_email_read_outlined,
                          color: FinoraColors.primary,
                          size: 16,
                        ),
                        const SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            message!,
                            style: const TextStyle(
                              color: FinoraColors.textPrimary,
                              fontSize: 11,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ),
                    if (onOpenEmailVerification != null)
                      TextButton.icon(
                        onPressed: onOpenEmailVerification,
                        icon: const Icon(Icons.verified_outlined, size: 16),
                        label: const Text('Xác thực email'),
                      ),
                  ],
                ),
              ),
            ),

          if (error != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 11,
                  vertical: 7,
                ),
                decoration: BoxDecoration(
                  color: const Color(0x1af04438),
                  borderRadius: BorderRadius.circular(11),
                  border: Border.all(color: FinoraColors.danger),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.error_outline_rounded,
                      color: FinoraColors.danger,
                      size: 16,
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(
                        error!,
                        style: const TextStyle(
                          color: FinoraColors.textPrimary,
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

          const SizedBox(height: 12),

          _AnimatedGoldButton(
            busy: busy,
            label: registering
                ? _I18n.t(lang, 'registerBtn')
                : _I18n.t(lang, 'loginBtn'),
            onTap: onSubmit,
          ),

          const SizedBox(height: 2),
          Align(
            alignment: Alignment.center,
            child: TextButton(
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 4,
                ),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
              onPressed: onSwitch,
              child: Text(
                registering
                    ? _I18n.t(lang, 'switchLogin')
                    : _I18n.t(lang, 'switchRegister'),
                style: const TextStyle(
                  color: Color(0xfffbbf24),
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),

          const Divider(color: FinoraColors.border, height: 20),

          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _CardQuickAction(
                icon: Icons.account_balance_wallet_outlined,
                label: _I18n.t(lang, 'netWorth'),
              ),
              _CardQuickAction(
                icon: Icons.smart_toy_outlined,
                label: _I18n.t(lang, 'aiAssistant'),
              ),
              _CardQuickAction(
                icon: Icons.lock_person_outlined,
                label: _I18n.t(lang, 'security'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _EmailVerificationForm extends StatefulWidget {
  const _EmailVerificationForm({
    super.key,
    required this.email,
    required this.code,
    required this.busy,
    required this.error,
    required this.onVerify,
    required this.onResend,
    required this.onBack,
  });

  final String email;
  final TextEditingController code;
  final bool busy;
  final String? error;
  final VoidCallback onVerify;
  final Future<bool> Function() onResend;
  final VoidCallback onBack;

  @override
  State<_EmailVerificationForm> createState() => _EmailVerificationFormState();
}

class _EmailVerificationFormState extends State<_EmailVerificationForm> {
  static const _resendCooldown = 60;
  late int _secondsRemaining;
  Timer? _cooldownTimer;
  bool _isResending = false;

  @override
  void initState() {
    super.initState();
    _secondsRemaining = _resendCooldown;
    _startCooldown();
  }

  void _startCooldown() {
    _cooldownTimer?.cancel();
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_secondsRemaining <= 1) {
        timer.cancel();
        if (mounted) setState(() => _secondsRemaining = 0);
        return;
      }
      if (mounted) setState(() => _secondsRemaining--);
    });
  }

  Future<void> _resendCode() async {
    if (_secondsRemaining > 0 || _isResending || widget.busy) return;
    setState(() => _isResending = true);
    final sent = await widget.onResend();
    if (!mounted) return;
    setState(() => _isResending = false);
    if (sent) {
      setState(() => _secondsRemaining = _resendCooldown);
      _startCooldown();
    }
  }

  @override
  void dispose() {
    _cooldownTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.94),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: FinoraColors.border),
        boxShadow: [
          BoxShadow(
            color: FinoraColors.primary.withValues(alpha: 0.12),
            blurRadius: 24,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 64,
            height: 64,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [FinoraColors.primary, FinoraColors.purple],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(20),
              boxShadow: [
                BoxShadow(
                  color: FinoraColors.primary.withValues(alpha: 0.28),
                  blurRadius: 18,
                  offset: const Offset(0, 8),
                ),
              ],
            ),
            child: const Icon(
              Icons.mark_email_read_rounded,
              color: Colors.white,
              size: 32,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'Kiểm tra hộp thư của bạn',
            textAlign: TextAlign.center,
            style: TextStyle(
              color: FinoraColors.textPrimary,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Chúng tôi đã gửi mã xác thực gồm 6 chữ số tới',
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: FinoraColors.textSecondary,
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            widget.email,
            textAlign: TextAlign.center,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: FinoraColors.primary,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 9),
            decoration: BoxDecoration(
              color: const Color(0xffecfdf5),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xffa7f3d0)),
            ),
            child: const Row(
              children: [
                Icon(
                  Icons.check_circle_rounded,
                  color: Color(0xff059669),
                  size: 17,
                ),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Mã xác thực đã được gửi thành công',
                    style: TextStyle(
                      color: Color(0xff047857),
                      fontSize: 11.5,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 18),
          const Align(
            alignment: Alignment.centerLeft,
            child: Text(
              'NHẬP MÃ XÁC THỰC',
              style: TextStyle(
                color: FinoraColors.textSecondary,
                fontSize: 10,
                letterSpacing: 0.8,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(height: 7),
          TextField(
            controller: widget.code,
            autofocus: true,
            textAlign: TextAlign.center,
            keyboardType: TextInputType.number,
            autofillHints: const [AutofillHints.oneTimeCode],
            maxLength: 6,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            style: const TextStyle(
              color: FinoraColors.textPrimary,
              fontSize: 24,
              letterSpacing: 10,
              fontWeight: FontWeight.w800,
            ),
            decoration: InputDecoration(
              counterText: '',
              hintText: '000000',
              hintStyle: TextStyle(
                color: FinoraColors.textTertiary.withValues(alpha: 0.55),
                fontSize: 23,
                letterSpacing: 8,
                fontWeight: FontWeight.w700,
              ),
              filled: true,
              fillColor: const Color(0xfffaf9ff),
              contentPadding: const EdgeInsets.symmetric(vertical: 13),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: const BorderSide(color: FinoraColors.border),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: const BorderSide(color: FinoraColors.border),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: const BorderSide(
                  color: FinoraColors.primary,
                  width: 1.8,
                ),
              ),
            ),
          ),
          if (widget.error != null) ...[
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0x1af04438),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                widget.error!,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: FinoraColors.danger,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
          const SizedBox(height: 16),
          _AnimatedGoldButton(
            busy: widget.busy,
            label: 'Xác thực và vào Finora',
            onTap: widget.onVerify,
          ),
          const SizedBox(height: 8),
          TextButton.icon(
            onPressed: widget.busy || _isResending || _secondsRemaining > 0
                ? null
                : _resendCode,
            icon: _isResending
                ? const SizedBox(
                    width: 15,
                    height: 15,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.refresh_rounded, size: 17),
            label: Text(
              _secondsRemaining > 0
                  ? 'Gửi lại mã sau ${_secondsRemaining}s'
                  : 'Gửi lại mã xác thực',
            ),
          ),
          TextButton(
            onPressed: widget.busy ? null : widget.onBack,
            child: const Text('Quay lại đăng nhập'),
          ),
        ],
      ),
    );
  }
}

class _CustomGlassTextField extends StatelessWidget {
  const _CustomGlassTextField({
    required this.controller,
    required this.labelText,
    required this.icon,
    this.obscureText = false,
    this.keyboardType,
    this.textCapitalization = TextCapitalization.none,
    this.suffixIcon,
    this.onChanged,
  });

  final TextEditingController controller;
  final String labelText;
  final IconData icon;
  final bool obscureText;
  final TextInputType? keyboardType;
  final TextCapitalization textCapitalization;
  final Widget? suffixIcon;
  final ValueChanged<String>? onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 2, bottom: 4),
          child: Text(
            labelText,
            style: TextStyle(
              color: FinoraColors.textSecondary,
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.2,
            ),
          ),
        ),
        TextField(
          controller: controller,
          obscureText: obscureText,
          keyboardType: keyboardType,
          textCapitalization: textCapitalization,
          onChanged: onChanged,
          inputFormatters: keyboardType == TextInputType.number
              ? [FilteringTextInputFormatter.digitsOnly]
              : null,
          style: const TextStyle(
            color: FinoraColors.textPrimary,
            fontSize: 13,
            fontWeight: FontWeight.w700,
          ),
          decoration: InputDecoration(
            hintText: labelText,
            hintStyle: const TextStyle(
              color: FinoraColors.textTertiary,
              fontSize: 12,
              fontWeight: FontWeight.w500,
            ),
            filled: true,
            fillColor: const Color(0xfffaf9ff),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 11,
            ),
            prefixIcon: Icon(icon, color: FinoraColors.primary, size: 18),
            suffixIcon: suffixIcon,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: FinoraColors.border,
                width: 1.2,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: FinoraColors.border,
                width: 1.2,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: FinoraColors.primary,
                width: 1.8,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _CardQuickAction extends StatelessWidget {
  const _CardQuickAction({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(7),
          decoration: BoxDecoration(
            color: FinoraColors.primarySoft,
            shape: BoxShape.circle,
            border: Border.all(color: FinoraColors.border),
          ),
          child: Icon(icon, color: FinoraColors.primary, size: 16),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(
            color: FinoraColors.textPrimary,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _AnimatedGoldButton extends StatefulWidget {
  const _AnimatedGoldButton({
    required this.busy,
    required this.label,
    required this.onTap,
  });

  final bool busy;
  final String label;
  final VoidCallback? onTap;

  @override
  State<_AnimatedGoldButton> createState() => _AnimatedGoldButtonState();
}

class _AnimatedGoldButtonState extends State<_AnimatedGoldButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      enabled: !widget.busy && widget.onTap != null,
      label: widget.label,
      child: GestureDetector(
        onTapDown: (_) => setState(() => _pressed = true),
        onTapUp: (_) => setState(() => _pressed = false),
        onTapCancel: () => setState(() => _pressed = false),
        child: AnimatedScale(
          scale: _pressed ? 0.96 : 1.0,
          duration: const Duration(milliseconds: 120),
          curve: Curves.easeOutCubic,
          child: Container(
            height: 48,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [FinoraColors.purple, FinoraColors.primary],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(22),
              boxShadow: [
                BoxShadow(
                  color: const Color(
                    0xff6d5df6,
                  ).withValues(alpha: _pressed ? 0.25 : 0.45),
                  blurRadius: _pressed ? 6 : 14,
                  offset: _pressed ? const Offset(0, 2) : const Offset(0, 4),
                ),
              ],
            ),
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                borderRadius: BorderRadius.circular(22),
                onTap: widget.busy ? null : widget.onTap,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    if (widget.busy)
                      const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2.2,
                        ),
                      )
                    else ...[
                      Container(
                        padding: const EdgeInsets.all(3),
                        decoration: BoxDecoration(
                          color: Colors.black.withValues(alpha: 0.15),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.arrow_forward_rounded,
                          color: Colors.white,
                          size: 16,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        widget.label,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.3,
                        ),
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

class _LoginBottomNav extends StatelessWidget {
  const _LoginBottomNav({required this.lang});
  final String lang;

  @override
  Widget build(BuildContext context) {
    final items = [
      _BottomNavItem(
        icon: Icons.receipt_long_outlined,
        label: _I18n.t(lang, 'logExpense'),
      ),
      _BottomNavItem(
        icon: Icons.show_chart_rounded,
        label: _I18n.t(lang, 'portfolio'),
      ),
      _BottomNavItem(
        icon: Icons.pie_chart_outline_rounded,
        label: _I18n.t(lang, 'budget'),
      ),
      _BottomNavItem(
        icon: Icons.auto_graph_rounded,
        label: _I18n.t(lang, 'forecast'),
      ),
      _BottomNavItem(
        icon: Icons.psychology_outlined,
        label: _I18n.t(lang, 'aiAssistant'),
      ),
    ];

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 8),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.88),
          borderRadius: BorderRadius.circular(28),
          border: Border.all(color: FinoraColors.border),
          boxShadow: [
            BoxShadow(
              color: FinoraColors.primary.withValues(alpha: 0.10),
              blurRadius: 16,
            ),
          ],
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: items,
        ),
      ),
    );
  }
}

class _BottomNavItem extends StatelessWidget {
  const _BottomNavItem({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: FinoraColors.primarySoft,
            shape: BoxShape.circle,
            border: Border.all(color: FinoraColors.border),
          ),
          child: Icon(icon, color: FinoraColors.primary, size: 20),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          style: const TextStyle(
            color: FinoraColors.textSecondary,
            fontSize: 11,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    );
  }
}

class _BrandMark extends StatelessWidget {
  const _BrandMark({this.size = 38});
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(size * 0.28),
        gradient: const LinearGradient(
          colors: [
            FinoraColors.purple,
            FinoraColors.primary,
            FinoraColors.deepPurple,
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        boxShadow: [
          BoxShadow(
            color: FinoraColors.primary.withValues(alpha: 0.35),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Center(
        child: Transform.rotate(
          angle: 0.785398,
          child: Container(
            width: size * 0.48,
            height: size * 0.48,
            decoration: BoxDecoration(
              border: Border.all(color: Colors.white, width: 2.2),
              borderRadius: BorderRadius.circular(size * 0.1),
            ),
            child: Center(
              child: Container(
                width: size * 0.22,
                height: size * 0.22,
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(size * 0.04),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _BankGridItem {
  const _BankGridItem({
    required this.title,
    required this.icon,
    required this.accent,
    required this.index,
    this.badge,
  });
  final String title;
  final IconData icon;
  final Color accent;
  final int index;
  final String? badge;
}

Widget _buildBankIconTile({
  required _BankGridItem item,
  required bool isSelected,
  required VoidCallback onTap,
}) {
  return InkWell(
    onTap: onTap,
    borderRadius: BorderRadius.circular(16),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Stack(
          clipBehavior: Clip.none,
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(16),
                color: isSelected
                    ? const Color(0xfffbbf24)
                    : item.accent.withValues(alpha: 0.12),
                border: Border.all(
                  color: isSelected
                      ? const Color(0xffd97706)
                      : item.accent.withValues(alpha: 0.25),
                  width: isSelected ? 2 : 1.2,
                ),
                boxShadow: isSelected
                    ? [
                        BoxShadow(
                          color: const Color(0xfffbbf24).withValues(alpha: 0.4),
                          blurRadius: 10,
                          spreadRadius: 1,
                        ),
                      ]
                    : [],
              ),
              child: Icon(
                item.icon,
                color: isSelected ? const Color(0xff1c1917) : item.accent,
                size: 24,
              ),
            ),
            if (item.badge != null)
              Positioned(
                top: -4,
                right: -6,
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 5,
                    vertical: 1.5,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xffef4444),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.white, width: 1.5),
                  ),
                  child: Text(
                    item.badge!,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 8.5,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 6),
        SizedBox(
          width: 68,
          child: Text(
            item.title,
            textAlign: TextAlign.center,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: isSelected
                  ? const Color(0xffd97706)
                  : const Color(0xff334155),
              fontSize: 11.5,
              fontWeight: isSelected ? FontWeight.w800 : FontWeight.w700,
              height: 1.15,
            ),
          ),
        ),
      ],
    ),
  );
}

Widget _buildBankSectionCard({
  required String sectionTitle,
  required List<_BankGridItem> items,
  required int selectedIndex,
  required ValueChanged<int> onSelect,
}) {
  return Container(
    margin: const EdgeInsets.only(bottom: 16),
    padding: const EdgeInsets.fromLTRB(10, 14, 10, 16),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: 0.94),
      borderRadius: BorderRadius.circular(26),
      border: Border.all(color: Colors.white, width: 1.5),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.05),
          blurRadius: 24,
          spreadRadius: 0,
        ),
      ],
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 8, bottom: 12),
          child: Text(
            sectionTitle,
            style: const TextStyle(
              color: Color(0xff1e293b),
              fontWeight: FontWeight.w900,
              fontSize: 14.5,
              letterSpacing: 0.2,
            ),
          ),
        ),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: items.map((item) {
            return Expanded(
              child: _buildBankIconTile(
                item: item,
                isSelected: selectedIndex == item.index,
                onTap: () => onSelect(item.index),
              ),
            );
          }).toList(),
        ),
      ],
    ),
  );
}

class HomePage extends StatefulWidget {
  const HomePage({super.key, required this.api, required this.loginBuilder});
  final ApiClient api;
  final WidgetBuilder loginBuilder;
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int index = 0;
  int refreshCounter = 0;
  final pages = const [
    NavItem('Tổng quan', Icons.dashboard_rounded),
    NavItem('Tài khoản', Icons.account_balance_rounded),
    NavItem('Giao dịch', Icons.receipt_long_rounded),
    NavItem('Khoản vay', Icons.request_quote_rounded),
    NavItem('Tài sản', Icons.inventory_2_rounded),
    NavItem('Bất động sản', Icons.home_work_rounded),
    NavItem('Ngân sách', Icons.pie_chart_rounded),
    NavItem('Dự báo', Icons.auto_graph_rounded),
    NavItem('Danh mục', Icons.category_rounded),
    NavItem('Ngân hàng', Icons.account_balance_wallet_rounded),
    NavItem('Tự động hóa', Icons.bolt_rounded),
    NavItem('Trợ lý AI', Icons.smart_toy_rounded),
    NavItem('Nhật ký hoạt động', Icons.history_rounded),
    NavItem('Cá nhân', Icons.person_rounded),
  ];
  @override
  Widget build(BuildContext context) {
    final compact = MediaQuery.of(context).size.width < 700;
    return Scaffold(
      extendBodyBehindAppBar: true,
      backgroundColor: FinoraColors.background,
      appBar: AppBar(
        backgroundColor: Colors.white.withValues(alpha: 0.92),
        elevation: 0,
        title: Row(
          children: [
            const _BrandMark(size: 32),
            const SizedBox(width: 8),
            Text(
              'finora',
              style: TextStyle(
                color: FinoraColors.ink,
                fontWeight: FontWeight.w900,
                fontSize: 22,
                letterSpacing: -1,
              ),
            ),
            if (!compact) ...[
              const SizedBox(width: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: FinoraColors.primarySoft,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: const Text(
                  'WEALTH OS',
                  style: TextStyle(
                    color: FinoraColors.primary,
                    fontSize: 9,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 0.8,
                  ),
                ),
              ),
            ],
          ],
        ),
        actions: [
          Semantics(
            button: true,
            label: 'Tìm kiếm',
            child: IconButton(
              tooltip: 'Tìm kiếm',
              onPressed: _showQuickActionSheet,
              icon: const Icon(Icons.search_rounded, color: FinoraColors.ink),
            ),
          ),
          const SizedBox(width: FinoraSpace.xs),
        ],
      ),
      drawer: null,
      body: Stack(
        children: [
          Positioned.fill(
            child: Opacity(
              opacity: 0.075,
              child: Image.asset(
                'assets/images/app_bg_maple_light.png',
                fit: BoxFit.cover,
                alignment: Alignment.center,
              ),
            ),
          ),
          Positioned.fill(
            child: Container(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: const [
                    Color(0x00fafafc),
                    Color(0x44f3f0ff),
                    Color(0x66fafafc),
                  ],
                ),
              ),
            ),
          ),
          SafeArea(
            child: Row(
              children: [
                if (!compact)
                  SizedBox(
                    width: 255,
                    child: Material(
                      color: Colors.white.withValues(alpha: 0.95),
                      child: _nav(),
                    ),
                  ),
                Expanded(child: _body()),
              ],
            ),
          ),
        ],
      ),
      bottomNavigationBar: compact
          ? SafeArea(
              top: false,
              child: Container(
                height: 74,
                decoration: BoxDecoration(
                  color: Colors.white,
                  border: Border(top: BorderSide(color: FinoraColors.border)),
                  boxShadow: const [
                    BoxShadow(
                      color: Color(0x0f000000),
                      blurRadius: 10,
                      offset: Offset(0, -2),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: _buildBottomNavItem(
                        icon: Icons.home_rounded,
                        label: 'Trang chủ',
                        isSelected: index == 0,
                        onTap: () => setState(() => index = 0),
                      ),
                    ),
                    Expanded(
                      child: _buildBottomNavItem(
                        icon: Icons.account_balance_wallet_rounded,
                        label: 'Tài khoản',
                        isSelected: index == 1,
                        onTap: () => setState(() => index = 1),
                      ),
                    ),
                    _buildQuickActionButton(),
                    Expanded(
                      child: _buildBottomNavItem(
                        icon: Icons.receipt_long_rounded,
                        label: 'Giao dịch',
                        isSelected: index == 2,
                        onTap: () => setState(() => index = 2),
                      ),
                    ),
                    Expanded(
                      child: _buildBottomNavItem(
                        icon: Icons.person_rounded,
                        label: 'Cá nhân',
                        isSelected: index == 13,
                        onTap: () => setState(() => index = 13),
                      ),
                    ),
                  ],
                ),
              ),
            )
          : null,
    );
  }

  Widget _buildQuickActionButton() => Semantics(
    button: true,
    label: 'Thao tác nhanh',
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: FinoraSpace.xs),
      child: SizedBox(
        width: 56,
        height: 58,
        child: Material(
          color: FinoraColors.primary,
          shape: const CircleBorder(),
          elevation: 5,
          child: InkWell(
            customBorder: const CircleBorder(),
            onTap: _showQuickActionSheet,
            child: const Icon(Icons.add_rounded, color: Colors.white, size: 28),
          ),
        ),
      ),
    ),
  );

  void _showQuickActionSheet() {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (sheetContext) => SafeArea(
        top: false,
        child: Container(
          padding: const EdgeInsets.fromLTRB(
            FinoraSpace.xl,
            FinoraSpace.sm,
            FinoraSpace.xl,
            FinoraSpace.xl,
          ),
          decoration: const BoxDecoration(
            color: FinoraColors.surfaceElevated,
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: const BoxDecoration(
                    color: FinoraColors.borderStrong,
                    borderRadius: FinoraRadius.full,
                  ),
                ),
              ),
              const SizedBox(height: FinoraSpace.lg),
              const Text('Thao tác nhanh', style: FinoraTypography.h3),
              const SizedBox(height: FinoraSpace.xs),
              const Text(
                'Ghi nhận thay đổi tài sản mà không rời khỏi luồng hiện tại.',
                style: FinoraTypography.bodySmall,
              ),
              const SizedBox(height: FinoraSpace.md),
              _quickActionTile(
                icon: Icons.add_card_rounded,
                title: 'Tạo giao dịch',
                subtitle: 'Ghi thu hoặc chi',
                onTap: () {
                  Navigator.pop(sheetContext);
                  showModalBottomSheet(
                    isScrollControlled: true,
                    backgroundColor: Colors.transparent,
                    context: context,
                    builder: (_) => _TransactionFormSheet(
                      api: widget.api,
                      onSuccess: () => setState(() => refreshCounter++),
                    ),
                  );
                },
              ),
              _quickActionTile(
                icon: Icons.request_quote_rounded,
                title: 'Cho vay mới',
                subtitle: 'Mở danh sách khoản vay để tạo khoản vay',
                onTap: () {
                  Navigator.pop(sheetContext);
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => LoanPage(
                        autoOpenCreate: true,
                        api: widget.api,
                        viewModel: LoanViewModel(
                          LoanRepositoryImpl(LoanRemoteService(widget.api)),
                        ),
                      ),
                    ),
                  );
                },
              ),
              _quickActionTile(
                icon: Icons.payments_rounded,
                title: 'Thu lãi / gốc',
                subtitle: 'Chọn khoản vay cần ghi nhận thu',
                onTap: () {
                  Navigator.pop(sheetContext);
                  setState(() => index = 3);
                },
              ),
              _quickActionTile(
                icon: Icons.swap_horiz_rounded,
                title: 'Chuyển tiền',
                subtitle: 'Chuyển tiền giữa các tài khoản của bạn',
                onTap: () {
                  Navigator.pop(sheetContext);
                  showModalBottomSheet(
                    isScrollControlled: true,
                    backgroundColor: Colors.transparent,
                    context: context,
                    builder: (_) => _TransferFormSheet(
                      api: widget.api,
                      onSuccess: () => setState(() => refreshCounter++),
                    ),
                  );
                },
              ),
              _quickActionTile(
                icon: Icons.add_business_rounded,
                title: 'Thêm tài sản',
                subtitle: 'Ghi nhận một tài sản mới',
                onTap: () {
                  Navigator.pop(sheetContext);
                  showModalBottomSheet(
                    isScrollControlled: true,
                    backgroundColor: Colors.transparent,
                    context: context,
                    builder: (_) => _AssetCreateSheet(
                      api: widget.api,
                      onSuccess: () => setState(() => refreshCounter++),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _quickActionTile({
    required IconData icon,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
  }) => Material(
    color: Colors.transparent,
    child: ListTile(
      contentPadding: const EdgeInsets.symmetric(vertical: FinoraSpace.xs),
      minVerticalPadding: FinoraSpace.xs,
      minLeadingWidth: 48,
      leading: Container(
        width: 44,
        height: 44,
        decoration: const BoxDecoration(
          color: FinoraColors.primarySoft,
          borderRadius: FinoraRadius.md,
        ),
        child: Icon(icon, color: FinoraColors.primary),
      ),
      title: Text(title, style: FinoraTypography.title),
      subtitle: Text(
        subtitle,
        style: FinoraTypography.caption.copyWith(
          color: FinoraColors.textSecondary,
        ),
      ),
      trailing: const Icon(
        Icons.chevron_right_rounded,
        color: FinoraColors.textSecondary,
      ),
      onTap: onTap,
    ),
  );

  void _showSettingsModal() {
    showModalBottomSheet(
      context: context,
      backgroundColor: FinoraColors.surfaceElevated,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.tune_rounded,
                        color: FinoraColors.primary,
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      const Text(
                        'Cách hiển thị số tiền',
                        style: FinoraTypography.h3,
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(
                          Icons.close_rounded,
                          color: FinoraColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Chọn cách hiển thị số tiền trên toàn bộ ứng dụng.',
                    style: FinoraTypography.bodySmall.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildDisplayModeTile(
                    title: 'Viết tắt / Rút gọn (100)',
                    subtitle: 'Ví dụ: 100.000 VND hiển thị thành 100',
                    value: 'compact',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 10),
                  _buildDisplayModeTile(
                    title: 'Đầy đủ (100.000 VND)',
                    subtitle: 'Hiển thị nguyên văn giá trị gốc đầy đủ',
                    value: 'full',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 16),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildDisplayModeTile({
    required String title,
    required String subtitle,
    required String value,
    required StateSetter setModalState,
  }) {
    final isSelected = appAmountDisplayMode == value;
    return InkWell(
      onTap: () {
        setState(() {
          appAmountDisplayMode = value;
        });
        setModalState(() {});
        try {
          widget.api.request('PUT', '/user/settings', {
            'amountDisplayMode': value,
          });
        } catch (_) {}
      },
      borderRadius: BorderRadius.circular(16),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected ? FinoraColors.primarySoft : FinoraColors.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected ? FinoraColors.primary : FinoraColors.border,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: FinoraTypography.title.copyWith(
                      color: isSelected
                          ? FinoraColors.primaryDeep
                          : FinoraColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: FinoraTypography.caption.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              const Icon(
                Icons.check_circle_rounded,
                color: FinoraColors.primary,
                size: 22,
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildBottomNavItem({
    required IconData icon,
    required String label,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final activeColor = isDark ? FinoraColors.accentGold : FinoraColors.violet;
    final inactiveColor = isDark ? Colors.white60 : FinoraColors.textSecondary;
    return Semantics(
      button: true,
      selected: isSelected,
      label: label,
      child: InkWell(
        onTap: onTap,
        borderRadius: FinoraRadius.sm,
        child: SizedBox(
          height: 54,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                color: isSelected ? activeColor : inactiveColor,
                size: 20,
              ),
              const SizedBox(height: 2),
              Text(
                label,
                style: TextStyle(
                  color: isSelected ? activeColor : inactiveColor,
                  fontSize: 10,
                  fontWeight: isSelected ? FontWeight.w800 : FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _nav() => ListView(
    padding: const EdgeInsets.fromLTRB(14, 80, 14, 20),
    children: [
      const Padding(
        padding: EdgeInsets.fromLTRB(6, 4, 6, 16),
        child: Row(
          children: [
            Icon(
              Icons.account_circle_rounded,
              color: Color(0xfffbbf24),
              size: 18,
            ),
            SizedBox(width: 8),
            Text(
              'DANH MỤC TIỆN ÍCH',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w900,
                letterSpacing: 1.2,
                color: Color(0xfffbbf24),
              ),
            ),
          ],
        ),
      ),
      _buildBankSectionCard(
        sectionTitle: '👑 Quản lý Tài sản & Dòng tiền',
        items: const [
          _BankGridItem(
            title: 'Tài khoản',
            icon: Icons.account_balance_wallet_rounded,
            accent: Color(0xff38bdf8),
            index: 1,
          ),
          _BankGridItem(
            title: 'Giao dịch',
            icon: Icons.receipt_long_rounded,
            accent: Color(0xff4ade80),
            index: 2,
          ),
          _BankGridItem(
            title: 'Hạn mức',
            icon: Icons.pie_chart_outline_rounded,
            accent: Color(0xfff43f5e),
            index: 6,
          ),
          _BankGridItem(
            title: 'Danh mục',
            icon: Icons.cases_rounded,
            accent: Color(0xffa855f7),
            index: 8,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
      _buildBankSectionCard(
        sectionTitle: '🔥 Đầu tư & Bất động sản',
        items: const [
          _BankGridItem(
            title: 'Bất động sản',
            icon: Icons.home_work_rounded,
            accent: Color(0xfff43f5e),
            index: 5,
          ),
          _BankGridItem(
            title: 'Tài sản quý',
            icon: Icons.diamond_rounded,
            accent: Color(0xfffbbf24),
            index: 4,
          ),
          _BankGridItem(
            title: 'Khoản vay',
            icon: Icons.request_quote_rounded,
            accent: Color(0xffef4444),
            index: 3,
          ),
          _BankGridItem(
            title: 'Ngân hàng',
            icon: Icons.account_balance_rounded,
            accent: Color(0xfffbbf24),
            badge: 'Mới',
            index: 9,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
      _buildBankSectionCard(
        sectionTitle: '⚡ Tiện ích & Automation',
        items: const [
          _BankGridItem(
            title: 'Tự động hóa',
            icon: Icons.auto_fix_high_rounded,
            accent: Color(0xff38bdf8),
            index: 10,
          ),
          _BankGridItem(
            title: 'Dự báo',
            icon: Icons.trending_up_rounded,
            accent: Color(0xff4ade80),
            index: 7,
          ),
          _BankGridItem(
            title: 'Trợ lý AI',
            icon: Icons.smart_toy_rounded,
            accent: Color(0xffa855f7),
            index: 11,
          ),
          _BankGridItem(
            title: 'Nhật ký',
            icon: Icons.history_toggle_off_rounded,
            accent: Color(0xffe2e8f0),
            index: 12,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
    ],
  );

  Widget _body() {
    switch (index) {
      case 0:
        return DashboardPage(
          key: ValueKey('dash_$refreshCounter'),
          api: widget.api,
          onNavigate: (i) => setState(() => index = i),
        );
      case 1:
        return ResourcePage(
          key: ValueKey('acc_$refreshCounter'),
          api: widget.api,
          title: 'Tài khoản',
          path: '/accounts',
          fields: const [
            FieldSpec('name', 'Tên tài khoản'),
            FieldSpec('type', 'Loại', initial: 'cash'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
            FieldSpec('portfolioId', 'Mã danh mục'),
          ],
        );
      case 2:
        return TransactionsPage(
          key: ValueKey('tx_$refreshCounter'),
          api: widget.api,
        );
      case 3:
        return LoanPage(
          key: ValueKey('loan_$refreshCounter'),
          api: widget.api,
          viewModel: LoanViewModel(
            LoanRepositoryImpl(LoanRemoteService(widget.api)),
          ),
        );
      case 4:
        return ValuedResourcePage(
          api: widget.api,
          title: 'Tài sản',
          path: '/assets',
          fields: const [
            FieldSpec('name', 'Tên tài sản'),
            FieldSpec('type', 'Loại', initial: 'other'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
          ],
        );
      case 5:
        return ValuedResourcePage(
          api: widget.api,
          title: 'Bất động sản',
          path: '/properties',
          fields: const [
            FieldSpec('name', 'Tên tài sản'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
            FieldSpec('address', 'Địa chỉ'),
          ],
        );
      case 6:
        return BudgetPage(api: widget.api);
      case 7:
        return ForecastPage(api: widget.api);
      case 8:
        return ResourcePage(
          api: widget.api,
          title: 'Danh mục',
          path: '/portfolios',
          fields: const [
            FieldSpec('name', 'Tên danh mục'),
            FieldSpec('baseCurrency', 'Tiền tệ cơ sở', initial: 'VND'),
          ],
        );
      case 9:
        return BankPage(api: widget.api);
      case 10:
        return AutomationPage(api: widget.api);
      case 11:
        return AssistantPage(api: widget.api);
      case 13:
        return ProfilePage(
          api: widget.api,
          onOpenSettings: _showSettingsModal,
          onLogout: _logout,
        );
      default:
        return ReadonlyPage(
          api: widget.api,
          title: 'Nhật ký hoạt động',
          path: '/audit-logs',
        );
    }
  }

  void _logout() {
    widget.api.token = null;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: widget.loginBuilder),
      (_) => false,
    );
  }
}

class PageFrame extends StatelessWidget {
  const PageFrame({
    super.key,
    required this.title,
    required this.child,
    this.action,
  });
  final String title;
  final Widget child;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 76, 24, 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'FINORA / QUẢN LÝ TÀI SẢN',
                        style: TextStyle(
                          color: FinoraColors.violet,
                          fontSize: 10,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 1.2,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        title,
                        style: TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.w800,
                          color: FinoraColors.textPrimary,
                          letterSpacing: -0.4,
                        ),
                      ),
                    ],
                  ),
                ),
                if (action case final Widget action) action,
              ],
            ),
            const SizedBox(height: 16),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key, required this.api, this.onNavigate});
  final ApiClient api;
  final ValueChanged<int>? onNavigate;

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  Map? netWorth;
  List accounts = [], transactions = [], portfolios = [];
  String? error;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    setState(() {
      loading = true;
      error = null;
    });
    try {
      accounts = await widget.api.request('GET', '/accounts') as List;
      final tx = await widget.api.request('GET', '/transactions?limit=5');
      transactions = (tx as Map)['items'] as List? ?? [];
      netWorth = await widget.api.request('GET', '/net-worth') as Map;
    } catch (e) {
      error = presentableError(e);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Widget _buildQuickActionsRow() {
    return Padding(
      padding: const EdgeInsets.only(top: 14, bottom: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: [
          Expanded(
            child: _buildCircularQuickAction(
              icon: Icons.add_circle_outline_rounded,
              label: 'Tạo giao dịch',
              color: const Color(0xfffbbf24),
              onTap: () {
                showModalBottomSheet(
                  isScrollControlled: true,
                  backgroundColor: Colors.transparent,
                  context: context,
                  builder: (_) => _TransactionFormSheet(
                    api: widget.api,
                    onSuccess: () => load(),
                  ),
                );
              },
            ),
          ),
          Expanded(
            child: _buildCircularQuickAction(
              icon: Icons.account_balance_wallet_rounded,
              label: 'Tài khoản',
              color: const Color(0xff38bdf8),
              onTap: () => widget.onNavigate?.call(1),
            ),
          ),
          Expanded(
            child: _buildCircularQuickAction(
              icon: Icons.request_quote_rounded,
              label: 'Khoản vay',
              color: const Color(0xfff97316),
              onTap: () => widget.onNavigate?.call(3),
            ),
          ),
          Expanded(
            child: _buildCircularQuickAction(
              icon: Icons.auto_fix_high_rounded,
              label: 'Tự động',
              color: const Color(0xff4ade80),
              onTap: () => widget.onNavigate?.call(10),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCircularQuickAction({
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback onTap,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 52,
            height: 52,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: isDark
                  ? Colors.white.withValues(alpha: 0.94)
                  : Theme.of(context).colorScheme.surface,
              border: Border.all(
                color: isDark ? Colors.white : FinoraColors.border,
                width: 1.5,
              ),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.06),
                  blurRadius: 14,
                ),
              ],
            ),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(height: 6),
          FittedBox(
            fit: BoxFit.scaleDown,
            child: Text(
              label,
              style: TextStyle(
                color: isDark ? Colors.white : FinoraColors.textPrimary,
                fontSize: 11.5,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: 'Tổng quan',
    action: IconButton(
      onPressed: load,
      tooltip: 'Tải lại dữ liệu',
      icon: const Icon(Icons.refresh_rounded, color: FinoraColors.accentGold),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: FinoraColors.primary),
          )
        : (error != null && accounts.isEmpty && transactions.isEmpty)
        ? FinoraEmptyState(
            title: 'Chưa thể tải tổng quan',
            message: 'Kiểm tra kết nối rồi thử lại.',
            icon: Icons.cloud_off_rounded,
            action: FilledButton.icon(
              onPressed: load,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Thử lại'),
            ),
          )
        : RefreshIndicator(
            color: FinoraColors.primary,
            onRefresh: load,
            child: ListView(
              padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
              children: [
                if (error != null) ErrorBox(error!),
                _BalanceHero(
                  value: _formatMoney(netWorth?['netWorth']),
                  currency: netWorth?['baseCurrency']?.toString() ?? 'VND',
                  accountCount: accounts.length,
                ),
                _buildQuickActionsRow(),
                const SizedBox(height: FinoraSpace.lg),
                const Text('Tài sản và nghĩa vụ', style: FinoraTypography.h3),
                const SizedBox(height: FinoraSpace.sm),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      Metric(
                        'Tiền mặt',
                        _formatMoney(netWorth?['cash']),
                        Icons.payments_rounded,
                        accent: FinoraColors.info,
                      ),
                      const SizedBox(width: FinoraSpace.sm),
                      Metric(
                        'Nợ phải trả',
                        _formatMoney(netWorth?['liabilities']),
                        Icons.credit_card_off_rounded,
                        accent: FinoraColors.danger,
                      ),
                      const SizedBox(width: FinoraSpace.sm),
                      Metric(
                        'Tài khoản',
                        accounts.length.toString(),
                        Icons.account_balance_rounded,
                        accent: FinoraColors.primary,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: FinoraSpace.xl),
                Row(
                  children: [
                    const Expanded(
                      child: Text(
                        'Giao dịch gần đây',
                        style: FinoraTypography.h3,
                      ),
                    ),
                    TextButton(
                      onPressed: () => widget.onNavigate?.call(2),
                      child: const Text('Xem tất cả'),
                    ),
                  ],
                ),
                const SizedBox(height: FinoraSpace.xs),
                if (transactions.isEmpty)
                  FinoraEmptyState(
                    title: 'Chưa có giao dịch',
                    message: 'Tạo giao dịch đầu tiên để theo dõi dòng tiền.',
                    icon: Icons.receipt_long_outlined,
                  )
                else
                  ...transactions.map(
                    (x) => FinoraListTile(
                      icon: x['type'] == 'income'
                          ? Icons.south_west_rounded
                          : Icons.north_east_rounded,
                      iconColor: x['type'] == 'income'
                          ? FinoraColors.success
                          : FinoraColors.danger,
                      title: (x['name']?.toString().trim().isNotEmpty == true)
                          ? x['name'].toString().trim()
                          : (x['note']?.toString().isNotEmpty == true
                                ? x['note'].toString()
                                : (x['type'] == 'income'
                                      ? 'Thu nhập'
                                      : 'Chi tiêu')),
                      subtitle: x['occurredAt']?.toString() ?? '',
                      amount:
                          "${x['type'] == 'income' ? '+' : '-'}${_formatMoney(x['amount'])} ${x['currency'] ?? 'VND'}",
                    ),
                  ),
              ],
            ),
          ),
  );

  String _formatMoney(dynamic raw) {
    if (raw == null) return '0';
    final d = double.tryParse(raw.toString());
    if (d == null) return raw.toString();
    final value = appAmountDisplayMode == 'compact' ? (d / 1000.0) : d;
    final parts = value.toStringAsFixed(2).split('.');
    final intPart = parts[0].replaceAllRegExp(
      RegExp(r'\B(?=(\d{3})+(?!\d))'),
      ',',
    );
    return parts[1] == '00' ? intPart : '$intPart.${parts[1]}';
  }
}

extension on String {
  String replaceAllRegExp(RegExp regex, String replacement) {
    return replaceAllMapped(regex, (match) => replacement);
  }
}

class _BalanceHero extends StatefulWidget {
  const _BalanceHero({
    required this.value,
    required this.currency,
    required this.accountCount,
  });
  final String value, currency;
  final int accountCount;

  @override
  State<_BalanceHero> createState() => _BalanceHeroState();
}

class _BalanceHeroState extends State<_BalanceHero> {
  bool hideBalance = false;

  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    padding: const EdgeInsets.all(22),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      borderRadius: BorderRadius.circular(28),
      color: FinoraColors.primaryDeep,
      border: Border.all(
        color: Colors.white.withValues(alpha: 0.35),
        width: 1.5,
      ),
      boxShadow: [
        BoxShadow(
          color: const Color(0x335b2382),
          blurRadius: 32,
          spreadRadius: 0,
        ),
      ],
    ),
    child: Stack(
      children: [
        Positioned(
          right: -25,
          bottom: -25,
          child: Icon(
            Icons.account_circle_rounded,
            size: 150,
            color: Colors.white.withValues(alpha: 0.12),
          ),
        ),
        Positioned(
          right: -10,
          top: -20,
          child: Container(
            width: 130,
            height: 130,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xfffbbf24).withValues(alpha: 0.15),
            ),
          ),
        ),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: const Color(0xfffbbf24).withValues(alpha: 0.22),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.account_balance_wallet_rounded,
                    color: Color(0xfffbbf24),
                    size: 18,
                  ),
                ),
                const SizedBox(width: 8),
                const Expanded(
                  child: Text(
                    'TỔNG TÀI SẢN RÒNG',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: Color(0xfffbbf24),
                      fontWeight: FontWeight.w900,
                      fontSize: 11.5,
                      letterSpacing: 1.2,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xfffbbf24),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Text(
                    'WEALTH OS VIP',
                    style: TextStyle(
                      color: Color(0xff1c1917),
                      fontSize: 9,
                      fontWeight: FontWeight.w900,
                      letterSpacing: 0.8,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              hideBalance ? '••••••••••••' : widget.value,
              style: const TextStyle(
                fontSize: 30,
                fontWeight: FontWeight.w900,
                color: Colors.white,
                letterSpacing: -0.5,
                shadows: [
                  Shadow(
                    color: Color(0x66000000),
                    blurRadius: 10,
                    offset: Offset(0, 4),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 2),
            Text(
              widget.currency,
              style: const TextStyle(
                color: Color(0xfffef08a),
                fontWeight: FontWeight.w900,
                fontSize: 13.5,
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 18),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 210),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 7,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.16),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: Colors.white.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.check_circle_rounded,
                          color: Color(0xff4ade80),
                          size: 14,
                        ),
                        const SizedBox(width: 6),
                        Flexible(
                          child: Text(
                            '${widget.accountCount} tài khoản đang theo dõi',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 11.5,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                Container(
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.16),
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.3),
                    ),
                  ),
                  child: IconButton(
                    onPressed: () => setState(() => hideBalance = !hideBalance),
                    icon: Icon(
                      hideBalance
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      color: const Color(0xfffbbf24),
                      size: 20,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ],
    ),
  );
}

class Metric extends StatelessWidget {
  const Metric(
    this.label,
    this.value,
    this.icon, {
    super.key,
    required this.accent,
  });
  final String label, value;
  final IconData icon;
  final Color accent;

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: 190,
      padding: const EdgeInsets.all(FinoraSpace.md),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: isDark ? const Color(0xff20152c) : FinoraColors.surfaceElevated,
        borderRadius: FinoraRadius.xl,
        border: Border.all(
          color: isDark ? const Color(0xff463652) : FinoraColors.border,
        ),
        boxShadow: isDark ? const [] : FinoraElevation.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(FinoraSpace.xs),
            decoration: BoxDecoration(
              color: accent.withValues(alpha: 0.14),
              borderRadius: FinoraRadius.sm,
            ),
            child: Icon(icon, color: accent, size: 20),
          ),
          const SizedBox(height: FinoraSpace.sm),
          Text(
            value,
            style: TextStyle(
              fontWeight: FontWeight.w900,
              color: isDark
                  ? const Color(0xfff4effa)
                  : FinoraColors.textPrimary,
              fontSize: 17,
            ),
          ),
          const SizedBox(height: FinoraSpace.xxs),
          Text(
            label,
            style: TextStyle(
              color: isDark
                  ? const Color(0xffa792be)
                  : FinoraColors.textSecondary,
              fontWeight: FontWeight.w600,
              fontSize: 11.5,
            ),
          ),
        ],
      ),
    );
  }
}

class ResourcePage extends StatefulWidget {
  const ResourcePage({
    super.key,
    required this.api,
    required this.title,
    required this.path,
    required this.fields,
  });
  final ApiClient api;
  final String title, path;
  final List<FieldSpec> fields;
  @override
  State<ResourcePage> createState() => _ResourcePageState();
}

class _ResourcePageState extends State<ResourcePage> {
  late Map<String, TextEditingController> ctl;
  List items = [];
  bool loading = true;
  String? error;
  @override
  void initState() {
    super.initState();
    ctl = {
      for (final f in widget.fields)
        f.key: TextEditingController(text: f.initial),
    };
    load();
  }

  @override
  void dispose() {
    for (final x in ctl.values) {
      x.dispose();
    }
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final data = await widget.api.request('GET', widget.path);
      items = data is List ? data : ((data as Map)['items'] as List? ?? []);
      error = null;
    } catch (e) {
      error = presentableError(e);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> create() async {
    try {
      await widget.api.request('POST', widget.path, {
        for (final x in ctl.entries)
          if (x.value.text.trim().isNotEmpty) x.key: x.value.text.trim(),
      });
      for (final f in widget.fields) {
        ctl[f.key]!.text = f.initial;
      }
      if (mounted) Navigator.pop(context);
      await load();
    } catch (e) {
      if (mounted) showError(context, presentableError(e));
    }
  }

  void openForm() {
    if (widget.path == '/accounts' || widget.title == 'Tài khoản') {
      showModalBottomSheet(
        isScrollControlled: true,
        backgroundColor: Colors.transparent,
        context: context,
        builder: (_) => _AccountFormSheet(
          api: widget.api,
          onSuccess: () {
            Navigator.pop(context);
            load();
          },
        ),
      );
      return;
    }

    showModalBottomSheet(
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      context: context,
      builder: (_) => FinoraSheet(
        title: 'Thêm ${widget.title}',
        subtitle: 'Thông tin sẽ được lưu vào tài khoản hiện tại.',
        child: Column(
          children: [
            ...widget.fields.map(
              (f) => Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: _CustomGlassTextField(
                  controller: ctl[f.key]!,
                  labelText: f.label,
                  icon: Icons.edit_note_rounded,
                ),
              ),
            ),
            const SizedBox(height: 8),
            _AnimatedGoldButton(
              busy: false,
              label: 'Lưu thông tin',
              onTap: create,
            ),
          ],
        ),
      ),
    );
  }

  Future<void> deleteItem(String id, String itemName) async {
    final deletionNote = widget.path == '/accounts'
        ? 'Chỉ những tài khoản không có giao dịch hoặc dòng tiền liên kết mới có thể xóa.'
        : 'Thao tác này có thể không hoàn tác được nếu bản ghi đã có dữ liệu liên kết.';
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xee120320),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
          side: BorderSide(color: Colors.white.withValues(alpha: 0.3)),
        ),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xffef4444).withValues(alpha: 0.15),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.delete_outline_rounded,
                color: Color(0xffef4444),
                size: 22,
              ),
            ),
            const SizedBox(width: 10),
            Text(
              'Xóa ${widget.title}',
              style: const TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
        content: Text(
          'Bạn có chắc chắn muốn xóa "$itemName"?\n$deletionNote',
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.85),
            fontSize: 13,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Hủy', style: TextStyle(color: Colors.white70)),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: const Color(0xffef4444),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
              ),
            ),
            onPressed: () => Navigator.pop(context, true),
            child: const Text(
              'Xóa ngay',
              style: TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      await widget.api.request('DELETE', '${widget.path}/$id');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xff15803d),
            content: Text('Đã xóa "$itemName" thành công!'),
          ),
        );
      }
      await load();
    } catch (e) {
      if (mounted) {
        final rawErr = presentableError(e);
        final msg = rawErr.contains('ACCOUNT_HAS_TRANSACTIONS')
            ? 'Không thể xóa "$itemName" vì đang có giao dịch hoặc dòng tiền liên kết.'
            : 'Không thể xóa "$itemName": $rawErr';
        showError(context, msg);
      }
    }
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: widget.title,
    action: Row(
      children: [
        IconButton(
          onPressed: load,
          icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
        ),
        const SizedBox(width: 4),
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xfffbbf24),
            foregroundColor: const Color(0xff1c1917),
          ),
          onPressed: openForm,
          icon: const Icon(Icons.add_rounded),
          label: const Text(
            'Thêm',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
        ),
      ],
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: FinoraColors.primary),
          )
        : (error != null && items.isEmpty)
        ? FinoraEmptyState(
            title: 'Chưa thể tải ${widget.title.toLowerCase()}',
            message: 'Kiểm tra kết nối rồi thử lại.',
            icon: Icons.cloud_off_rounded,
            action: FilledButton.icon(
              onPressed: load,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Thử lại'),
            ),
          )
        : ListView(
            padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
            children: [
              const _ScreenIntro(
                'Quản lý thông tin tập trung, rõ ràng và an toàn.',
              ),
              if (error != null) ErrorBox(error!),
              if (items.isEmpty)
                FinoraEmptyState(
                  title: 'Chưa có ${widget.title.toLowerCase()}',
                  message: 'Thêm bản ghi đầu tiên để bắt đầu quản lý.',
                  icon: _iconForTitle(widget.title),
                  action: FilledButton.icon(
                    onPressed: openForm,
                    icon: const Icon(Icons.add_rounded),
                    label: Text('Thêm ${widget.title.toLowerCase()}'),
                  ),
                ),
              ...items.map((x) {
                final id = x['id']?.toString() ?? '';
                final name =
                    x['name']?.toString() ??
                    x['counterparty']?.toString() ??
                    x['id']?.toString() ??
                    '-';
                final tile = FinoraListTile(
                  icon: _iconForTitle(widget.title),
                  title: name,
                  subtitle: _details(x),
                  badge:
                      x['status']?.toString() ??
                      x['currency']?.toString() ??
                      '',
                );

                if (id.isEmpty) return tile;

                return Dismissible(
                  key: ValueKey('res_${widget.title}_$id'),
                  direction: DismissDirection.endToStart,
                  confirmDismiss: (_) async {
                    await deleteItem(id, name);
                    return false;
                  },
                  background: Container(
                    margin: const EdgeInsets.only(bottom: 10),
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    alignment: Alignment.centerRight,
                    decoration: BoxDecoration(
                      color: const Color(0xddef4444),
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0x66ef4444),
                          blurRadius: 16,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        Icon(
                          Icons.delete_forever_rounded,
                          color: Colors.white,
                          size: 24,
                        ),
                        SizedBox(width: 6),
                        Text(
                          'Xóa',
                          style: TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                  child: tile,
                );
              }),
            ],
          ),
  );
  String _details(dynamic x) {
    if (x is! Map) return '';

    final parts = <String>[];

    if (x.containsKey('type') && x['type'] != null) {
      final t = x['type'].toString();
      final typeLabel = switch (t) {
        'cash' => 'Tiền mặt',
        'bank' => 'Ngân hàng',
        'gold' => 'Vàng kim',
        'real_estate' => 'Bất động sản',
        'income' => 'Thu nhập',
        'expense' => 'Chi tiêu',
        _ => t,
      };
      parts.add('Loại: $typeLabel');
    }

    if (x.containsKey('balance') &&
        x['balance'] != null &&
        x['balance'].toString().isNotEmpty) {
      final b = double.tryParse(x['balance'].toString());
      final curr = x['currency']?.toString() ?? 'VND';
      parts.add(
        'Số dư: ${b == null ? x['balance'] : formatCurrency(b, currency: curr)}',
      );
    }

    if (x.containsKey('principal') && x['principal'] != null) {
      final principal = double.tryParse(x['principal'].toString());
      final curr = x['currency']?.toString() ?? 'VND';
      parts.add(
        'Gốc: ${principal == null ? x['principal'] : formatCurrency(principal, currency: curr)}',
      );
    }
    if (x.containsKey('rate') && x['rate'] != null) {
      parts.add('Lãi suất: ${x['rate']}%');
    }

    final rawDate =
        x['createdAt']?.toString() ??
        x['updatedAt']?.toString() ??
        x['occurredAt']?.toString();
    final formattedDate = _formatDate(rawDate);
    if (formattedDate.isNotEmpty) {
      parts.add('Ngày: $formattedDate');
    }

    if (parts.isNotEmpty) {
      return parts.join(' • ');
    }

    const noisyKeys = {
      'id',
      'userId',
      'portfolioId',
      'createdAt',
      'updatedAt',
      'occurredAt',
      'name',
      'type',
      'currency',
      'balance',
    };

    final fallbackEntries = x.entries
        .where(
          (e) =>
              !noisyKeys.contains(e.key) &&
              e.value != null &&
              e.value.toString().isNotEmpty,
        )
        .take(2)
        .map((e) => '${e.key}: ${e.value}')
        .join(' • ');

    return fallbackEntries;
  }
}

String _formatDate(String? raw) {
  if (raw == null || raw.trim().isEmpty) return '';
  try {
    final dt = DateTime.parse(raw).toLocal();
    final day = dt.day.toString().padLeft(2, '0');
    final month = dt.month.toString().padLeft(2, '0');
    final year = dt.year;
    final hour = dt.hour.toString().padLeft(2, '0');
    final minute = dt.minute.toString().padLeft(2, '0');
    return '$day/$month/$year $hour:$minute';
  } catch (_) {
    return raw;
  }
}

class PersonalInfoPage extends StatefulWidget {
  const PersonalInfoPage({super.key, required this.api});
  final ApiClient api;

  @override
  State<PersonalInfoPage> createState() => _PersonalInfoPageState();
}

class _PersonalInfoPageState extends State<PersonalInfoPage> {
  String _email = '';
  String _phone = '';
  String _address = '';

  void _showEditBasicInfoModal() {
    final phoneCtrl = TextEditingController(text: _phone);
    final emailCtrl = TextEditingController(text: _email);
    final addressCtrl = TextEditingController(text: _address);
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(
          left: 20,
          right: 20,
          top: 20,
          bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
        ),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 36,
                  height: 4,
                  decoration: BoxDecoration(
                    color: const Color(0xffcbd5e1),
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Chỉnh sửa thông tin cơ bản',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Color(0xff1e293b),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: phoneCtrl,
                decoration: const InputDecoration(
                  labelText: 'Số điện thoại',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: emailCtrl,
                decoration: const InputDecoration(
                  labelText: 'Email liên hệ',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: addressCtrl,
                decoration: const InputDecoration(
                  labelText: 'Địa chỉ liên hệ',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xff6b21a8),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  onPressed: () {
                    setState(() {
                      _phone = phoneCtrl.text.trim();
                      _email = emailCtrl.text.trim();
                      _address = addressCtrl.text.trim();
                    });
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text('Đã cập nhật thông tin trong phiên này.'),
                      ),
                    );
                  },
                  child: const Text(
                    'Lưu thay đổi',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showLimitChangeModal() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: const Color(0xffcbd5e1),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Thay đổi hạn mức giao dịch',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Color(0xff1e293b),
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              'Tính năng điều chỉnh hạn mức chưa được kết nối với dữ liệu Finora.',
              style: TextStyle(fontSize: 13, color: Color(0xff64748b)),
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xffcbd5e1),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                onPressed: null,
                child: const Text(
                  'Chưa khả dụng',
                  style: TextStyle(
                    color: Color(0xff64748b),
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              child: TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: const Text('Đóng'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    const double progress = 0;

    return Scaffold(
      backgroundColor: const Color(0xfff8fafc),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Color(0xff1e293b)),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'Thông tin cá nhân',
          style: TextStyle(
            color: Color(0xff1e293b),
            fontSize: 17,
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        physics: const BouncingScrollPhysics(),
        children: [
          // Section 1: Dữ liệu sinh trắc học
          _buildSectionHeader('Dữ liệu sinh trắc học'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 10,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Khuôn mặt',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Color(0xff1e293b),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xffdcfce7),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'Đã cập nhật',
                    style: TextStyle(
                      color: Color(0xff15803d),
                      fontSize: 11.5,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 2: Hạn mức tài khoản
          _buildSectionHeader('Hạn mức tài khoản'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Hạn mức hiện tại của Bạn là:',
                      style: TextStyle(
                        fontSize: 13.5,
                        fontWeight: FontWeight.w600,
                        color: Color(0xff334155),
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 3.5,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xffdcfce7),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Text(
                        'Tối đa',
                        style: TextStyle(
                          color: Color(0xff15803d),
                          fontSize: 11.5,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                const Text(
                  'Chưa đồng bộ hạn mức',
                  style: TextStyle(
                    fontSize: 14.5,
                    fontWeight: FontWeight.w900,
                    color: Color(0xff6b21a8),
                    height: 1.3,
                  ),
                ),
                const SizedBox(height: 14),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: progress,
                    minHeight: 5,
                    backgroundColor: const Color(0xfff1f5f9),
                    valueColor: const AlwaysStoppedAnimation<Color>(
                      Color(0xff8b5cf6),
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                const Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Đã giao dịch trong ngày',
                          style: TextStyle(
                            fontSize: 11,
                            color: Color(0xff64748b),
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          '—',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xff0f172a),
                          ),
                        ),
                      ],
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          'Hạn mức còn lại',
                          style: TextStyle(
                            fontSize: 11,
                            color: Color(0xff64748b),
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          '—',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xff0f172a),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: 14),
                Semantics(
                  button: true,
                  label: 'Thay đổi hạn mức giao dịch',
                  child: GestureDetector(
                    onTap: _showLimitChangeModal,
                    child: const Text(
                      'Thay đổi hạn mức giao dịch',
                      style: TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.bold,
                        color: Color(0xffd97706),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 3: Thông tin cơ bản
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildSectionHeader('Thông tin cơ bản'),
              Semantics(
                button: true,
                label: 'Chỉnh sửa thông tin cơ bản',
                child: GestureDetector(
                  onTap: _showEditBasicInfoModal,
                  child: const Row(
                    children: [
                      Icon(
                        Icons.edit_outlined,
                        size: 14,
                        color: Color(0xffd97706),
                      ),
                      SizedBox(width: 4),
                      Text(
                        'Chỉnh sửa',
                        style: TextStyle(
                          fontSize: 12.5,
                          fontWeight: FontWeight.bold,
                          color: Color(0xffd97706),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildInfoRow('Tên đăng nhập', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow(
                  'Mã khách hàng',
                  'Chưa đồng bộ',
                  showInfoIcon: true,
                ),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Số điện thoại', _phone),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Email', _email),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Địa chỉ liên hệ', _address, isMultiLine: true),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 4: Thông tin định danh
          _buildSectionHeader('Thông tin định danh'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildInfoRow('Số CMND/CCCD/Hộ chiếu', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Họ và tên', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Ngày sinh', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Giới tính', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Quốc tịch', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Ngày cấp', 'Chưa đồng bộ'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Nơi cấp', 'Chưa đồng bộ', isMultiLine: true),
              ],
            ),
          ),
          const SizedBox(height: 30),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.bold,
        color: Color(0xff1e293b),
      ),
    );
  }

  Widget _buildInfoRow(
    String label,
    String value, {
    bool showInfoIcon = false,
    bool isMultiLine = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 11),
      child: Row(
        crossAxisAlignment: isMultiLine
            ? CrossAxisAlignment.start
            : CrossAxisAlignment.center,
        children: [
          Expanded(
            flex: 4,
            child: Row(
              children: [
                Flexible(
                  child: Text(
                    label,
                    style: const TextStyle(
                      fontSize: 12.5,
                      color: Color(0xff64748b),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                if (showInfoIcon) ...[
                  const SizedBox(width: 4),
                  const Icon(
                    Icons.info_outline_rounded,
                    size: 14,
                    color: Color(0xffd97706),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            flex: 6,
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xff0f172a),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.api,
    required this.onOpenSettings,
    required this.onLogout,
  });

  final ApiClient api;
  final VoidCallback onOpenSettings;
  final VoidCallback onLogout;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  // Retained for the legacy profile fallback below.
  int _satisfactionRating = 0;
  List<Map<String, dynamic>> _linkedAccounts = const [];
  String _sepaySubtitle = 'Kết nối ngân hàng qua SePay';
  String _syncLabel = 'Chưa liên kết';

  @override
  void initState() {
    super.initState();
    _loadSePaySummary();
  }

  Future<void> _loadSePaySummary() async {
    try {
      final data = await widget.api.request('GET', '/me/sepay') as Map;
      final accounts = (data['bankAccounts'] as List?) ?? const [];
      if (!mounted) return;
      if (accounts.isEmpty) {
        setState(() {
          _linkedAccounts = const [];
          _sepaySubtitle = 'Kết nối ngân hàng qua SePay';
          _syncLabel = 'Chưa liên kết';
        });
        return;
      }
      final profile = data['profile'] as Map?;
      final synced = profile?['lastSyncedAt']?.toString();
      setState(() {
        _linkedAccounts = accounts
            .whereType<Map>()
            .map((account) => Map<String, dynamic>.from(account))
            .toList(growable: false);
        _syncLabel = _relativeSync(synced);
        _sepaySubtitle =
            '${accounts.length} tài khoản đã liên kết · $_syncLabel';
      });
    } catch (_) {
      // This menu remains usable while the profile endpoint is temporarily unavailable.
    }
  }

  String _relativeSync(String? raw) {
    final value = raw == null ? null : DateTime.tryParse(raw);
    if (value == null || value.year == 1970) return 'Chưa đồng bộ';
    final minutes = DateTime.now().difference(value.toLocal()).inMinutes;
    if (minutes < 1) return 'Đồng bộ vừa xong';
    if (minutes < 60) return 'Đồng bộ $minutes phút trước';
    return 'Đồng bộ ${minutes ~/ 60} giờ trước';
  }

  Future<void> _openSePay() async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => SePayConnectionPage(api: widget.api)),
    );
    await _loadSePaySummary();
  }

  void _openPersonalInfo() {
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => SafeArea(
        top: false,
        child: Container(
          padding: const EdgeInsets.all(FinoraSpace.xl),
          decoration: const BoxDecoration(
            color: FinoraColors.surfaceElevated,
            borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
          ),
          child: FinoraEmptyState(
            title: 'Hồ sơ chưa đồng bộ',
            message:
                'Finora chưa nhận được dữ liệu hồ sơ từ nguồn tài khoản. Thông tin sẽ xuất hiện khi nguồn dữ liệu được kết nối.',
            icon: Icons.person_outline_rounded,
            action: FilledButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Đã hiểu'),
            ),
          ),
        ),
      ),
    );
  }

  void _showSupportModal(String title, String detail) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text(
          title,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
        ),
        content: Text(
          detail,
          style: const TextStyle(fontSize: 13.5, color: Color(0xff334155)),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text(
              'Đóng',
              style: TextStyle(
                color: Color(0xff6b21a8),
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _showAppearancePicker() async {
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (sheetContext) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(
            FinoraSpace.lg,
            FinoraSpace.sm,
            FinoraSpace.lg,
            FinoraSpace.xl,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Giao diện',
                style: Theme.of(sheetContext).textTheme.titleLarge,
              ),
              const SizedBox(height: FinoraSpace.xs),
              Text(
                'Chọn chế độ hiển thị phù hợp với bạn.',
                style: Theme.of(sheetContext).textTheme.bodyMedium,
              ),
              const SizedBox(height: FinoraSpace.md),
              const ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Icon(Icons.light_mode_outlined),
                title: Text('Đang dùng giao diện sáng'),
                subtitle: Text(
                  'Giao diện tối sẽ được bổ sung khi hệ thống hỗ trợ đầy đủ.',
                ),
              ),
              const SizedBox(height: FinoraSpace.sm),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: () => Navigator.pop(sheetContext),
                  child: const Text('Đóng'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) => _buildFinoraProfile(context);

  Widget _buildFinoraProfile(BuildContext context) => PageFrame(
    title: 'Cá nhân',
    child: SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildProfileOverview(),
          const SizedBox(height: FinoraSpace.xl),
          const Text('Tổng quan kết nối', style: FinoraTypography.h3),
          const SizedBox(height: FinoraSpace.sm),
          FinoraCard(
            onTap: _openSePay,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildConnectionSummary(),
                const SizedBox(height: FinoraSpace.md),
                _buildConnectionNotice(),
                if (_linkedAccounts.isNotEmpty) ...[
                  const SizedBox(height: FinoraSpace.sm),
                  Wrap(
                    spacing: FinoraSpace.xs,
                    runSpacing: FinoraSpace.xs,
                    children: _linkedAccounts
                        .take(3)
                        .map(_buildBankChip)
                        .toList(growable: false),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: FinoraSpace.xl),
          const Text('Thiết lập cá nhân', style: FinoraTypography.h3),
          const SizedBox(height: FinoraSpace.sm),
          FinoraCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                _buildMenuItemTile(
                  icon: Theme.of(context).brightness == Brightness.dark
                      ? Icons.dark_mode_rounded
                      : Icons.light_mode_rounded,
                  title: 'Giao diện',
                  subtitle: 'Giao diện sáng',
                  iconColor: FinoraColors.primary,
                  onTap: _showAppearancePicker,
                ),
                const Divider(height: 1, indent: 56),
                _buildMenuItemTile(
                  icon: Icons.tune_rounded,
                  title: 'Hiển thị số tiền',
                  subtitle: 'Chọn cách hiển thị số liệu tài chính',
                  iconColor: FinoraColors.primary,
                  onTap: widget.onOpenSettings,
                ),
                const Divider(height: 1, indent: 56),
                _buildMenuItemTile(
                  icon: Icons.account_balance_rounded,
                  title: 'Liên kết ngân hàng',
                  subtitle: _sepaySubtitle,
                  iconColor: FinoraColors.info,
                  onTap: _openSePay,
                ),
              ],
            ),
          ),
          const SizedBox(height: FinoraSpace.xl),
          const Text('Hỗ trợ & bảo mật', style: FinoraTypography.h3),
          const SizedBox(height: FinoraSpace.sm),
          FinoraCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                _buildMenuItemTile(
                  icon: Icons.help_outline_rounded,
                  title: 'Trung tâm hỗ trợ',
                  subtitle: 'Hướng dẫn và hỗ trợ sử dụng Finora',
                  iconColor: FinoraColors.accentAmber,
                  onTap: () => _showSupportModal(
                    'Trung tâm hỗ trợ',
                    'Kênh hỗ trợ chưa được cấu hình cho không gian làm việc này.',
                  ),
                ),
                const Divider(height: 1, indent: 56),
                _buildMenuItemTile(
                  icon: Icons.verified_user_outlined,
                  title: 'Dữ liệu của bạn được bảo vệ',
                  subtitle:
                      'Finora chỉ dùng dữ liệu cần thiết để cập nhật tài chính',
                  iconColor: FinoraColors.success,
                  onTap: () => _showSupportModal(
                    'Bảo vệ dữ liệu',
                    'Finora bảo vệ dữ liệu kết nối của bạn và chỉ dùng chúng để cung cấp các tính năng quản lý tài chính.',
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: FinoraSpace.xl),
          OutlinedButton.icon(
            onPressed: widget.onLogout,
            icon: const Icon(Icons.logout_rounded),
            label: const Text('Đăng xuất'),
            style: OutlinedButton.styleFrom(
              foregroundColor: FinoraColors.danger,
              side: const BorderSide(color: FinoraColors.danger),
              minimumSize: const Size.fromHeight(48),
            ),
          ),
        ],
      ),
    ),
  );

  Widget _buildProfileOverview() {
    final connected = _linkedAccounts.isNotEmpty;
    final mappedCount = _linkedAccounts
        .where((account) => account['mapping'] != null)
        .length;

    return Container(
      decoration: BoxDecoration(
        gradient: FinoraColors.goldGradient,
        borderRadius: FinoraRadius.xl,
        boxShadow: FinoraElevation.floating,
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: _openPersonalInfo,
          borderRadius: FinoraRadius.xl,
          child: Padding(
            padding: const EdgeInsets.all(FinoraSpace.lg),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const CircleAvatar(
                      radius: 25,
                      backgroundColor: Color(0x24ffffff),
                      foregroundColor: Colors.white,
                      child: Icon(Icons.person_outline_rounded, size: 28),
                    ),
                    const SizedBox(width: FinoraSpace.sm),
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Hồ sơ Finora',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 17,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          SizedBox(height: 2),
                          Text(
                            'Không gian tài chính cá nhân',
                            style: TextStyle(
                              color: Color(0xdfffffff),
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Icon(
                      Icons.chevron_right_rounded,
                      color: Colors.white,
                    ),
                  ],
                ),
                const SizedBox(height: FinoraSpace.lg),
                Text(
                  connected
                      ? 'Dữ liệu tài chính đang được cập nhật'
                      : 'Sẵn sàng thiết lập không gian tài chính',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 15,
                    height: 1.35,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: FinoraSpace.md),
                Row(
                  children: [
                    Expanded(
                      child: _buildOverviewMetric(
                        '${_linkedAccounts.length}',
                        'ngân hàng liên kết',
                      ),
                    ),
                    const SizedBox(width: FinoraSpace.sm),
                    Expanded(
                      child: _buildOverviewMetric(
                        '$mappedCount',
                        'nguồn dữ liệu đã gán',
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildOverviewMetric(String value, String label) => Container(
    padding: const EdgeInsets.symmetric(
      horizontal: FinoraSpace.sm,
      vertical: FinoraSpace.xs,
    ),
    decoration: const BoxDecoration(
      color: Color(0x1fffffff),
      borderRadius: FinoraRadius.sm,
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: Color(0xdfffffff),
            fontSize: 10,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    ),
  );

  Widget _buildConnectionSummary() => Row(
    children: [
      Container(
        width: 42,
        height: 42,
        decoration: const BoxDecoration(
          color: FinoraColors.primarySoft,
          borderRadius: FinoraRadius.sm,
        ),
        child: const Icon(
          Icons.account_balance_rounded,
          color: FinoraColors.primary,
        ),
      ),
      const SizedBox(width: FinoraSpace.sm),
      Expanded(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Tài khoản ngân hàng', style: FinoraTypography.title),
            const SizedBox(height: FinoraSpace.xxs),
            Text(_sepaySubtitle, style: FinoraTypography.bodySmall),
          ],
        ),
      ),
      const Icon(Icons.chevron_right_rounded),
    ],
  );

  Widget _buildConnectionNotice() => Container(
    width: double.infinity,
    padding: const EdgeInsets.symmetric(
      horizontal: FinoraSpace.sm,
      vertical: FinoraSpace.xs,
    ),
    decoration: const BoxDecoration(
      color: Color(0xfff8f7ff),
      borderRadius: FinoraRadius.sm,
    ),
    child: Row(
      children: [
        Icon(
          _linkedAccounts.isEmpty
              ? Icons.info_outline_rounded
              : Icons.sync_rounded,
          size: 16,
          color: _linkedAccounts.isEmpty
              ? FinoraColors.textSecondary
              : FinoraColors.success,
        ),
        const SizedBox(width: FinoraSpace.xs),
        Expanded(
          child: Text(
            _linkedAccounts.isEmpty
                ? 'Liên kết ngân hàng để Finora cập nhật dòng tiền tự động.'
                : _syncLabel,
            style: FinoraTypography.caption.copyWith(
              color: FinoraColors.textSecondary,
            ),
          ),
        ),
      ],
    ),
  );

  Widget _buildBankChip(Map<String, dynamic> account) {
    final bankName = account['bankName']?.toString().trim();
    final maskedNumber = account['accountNumberMasked']?.toString().trim();
    final label = (bankName?.isNotEmpty ?? false)
        ? bankName!
        : 'Tài khoản đã liên kết';
    return Container(
      constraints: const BoxConstraints(maxWidth: 190),
      padding: const EdgeInsets.symmetric(
        horizontal: FinoraSpace.sm,
        vertical: FinoraSpace.xs,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.primarySoft,
        borderRadius: FinoraRadius.full,
      ),
      child: Text(
        (maskedNumber?.isNotEmpty ?? false) ? '$label · $maskedNumber' : label,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: FinoraTypography.caption.copyWith(
          color: FinoraColors.primaryDeep,
        ),
      ),
    );
  }

  // ignore: unused_element
  Widget _buildLegacyProfile(BuildContext context) {
    return PageFrame(
      title: 'Cá nhân',
      child: ListView(
        physics: const BouncingScrollPhysics(),
        children: [
          // Header Card (Avatar + Name + CIF + Badge) -> Click opens Personal Info Screen
          InkWell(
            onTap: _openPersonalInfo,
            borderRadius: BorderRadius.circular(20),
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.06),
                    blurRadius: 16,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Column(
                children: [
                  Stack(
                    alignment: Alignment.center,
                    children: [
                      Container(
                        width: 76,
                        height: 76,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          gradient: LinearGradient(
                            colors: [
                              Color(0xffd97706),
                              Color(0xfffbbf24),
                              Color(0xfff59e0b),
                            ],
                          ),
                        ),
                      ),
                      Container(
                        width: 70,
                        height: 70,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          color: Colors.white,
                        ),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(35),
                          child: const Icon(
                            Icons.person_rounded,
                            size: 40,
                            color: Color(0xff6b21a8),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'Hồ sơ Finora',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                          color: Color(0xff2e1065),
                          letterSpacing: 0.3,
                        ),
                      ),
                      SizedBox(width: 4),
                      Icon(
                        Icons.chevron_right_rounded,
                        color: Color(0xffd97706),
                        size: 22,
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Thông tin tài khoản',
                    style: TextStyle(
                      fontSize: 11.5,
                      color: Color(0xff64748b),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 3.5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xffdcfce7),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xff86efac)),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.check_circle_rounded,
                          color: Color(0xff16a34a),
                          size: 13,
                        ),
                        SizedBox(width: 4),
                        Text(
                          'Đã xác thực',
                          style: TextStyle(
                            color: Color(0xff15803d),
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 14),

          // Main Support & Contact Services Menu (matching Image 1)
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 14,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildMenuItemTile(
                  icon: Icons.chat_rounded,
                  title: 'Trung tâm hỗ trợ Finora',
                  iconColor: const Color(0xffd97706),
                  onTap: () => _showSupportModal(
                    'Trung tâm hỗ trợ',
                    'Kênh hỗ trợ sẽ được cấu hình theo không gian làm việc của bạn.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.forum_rounded,
                  title: 'Gửi phản hồi',
                  iconColor: const Color(0xff2563eb),
                  onTap: () => _showSupportModal(
                    'Gửi phản hồi',
                    'Tính năng gửi phản hồi sẽ được kết nối trong phiên bản sau.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.mark_chat_read_rounded,
                  title: 'Trợ giúp sử dụng',
                  iconColor: const Color(0xff0284c7),
                  onTap: () => _showSupportModal(
                    'Trợ giúp sử dụng',
                    'Xem hướng dẫn sử dụng Finora.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.email_rounded,
                  title: 'Gửi Email',
                  iconColor: const Color(0xff06b6d4),
                  onTap: () => _showSupportModal(
                    'Gửi Email',
                    'Kênh email hỗ trợ chưa được cấu hình.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.phone_rounded,
                  title: 'Gọi đường dây hỗ trợ',
                  iconColor: const Color(0xff16a34a),
                  onTap: () => _showSupportModal(
                    'Gọi đường dây hỗ trợ',
                    'Đường dây hỗ trợ chưa được cấu hình.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.warning_amber_rounded,
                  title: 'Yêu cầu trợ giúp/Báo lỗi',
                  iconColor: const Color(0xffef4444),
                  onTap: () => _showSupportModal(
                    'Báo lỗi & Trợ giúp',
                    'Gửi yêu cầu hỗ trợ sự cố giao dịch.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.tune_rounded,
                  title: 'Cấu hình hiển thị số tiền',
                  subtitle: 'Rút gọn (100) so với đầy đủ (100.000 VND)',
                  iconColor: const Color(0xff7c3aed),
                  onTap: widget.onOpenSettings,
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.account_balance,
                  title: 'Liên kết ngân hàng qua SePay',
                  subtitle: _sepaySubtitle,
                  iconColor: const Color(0xff2563eb),
                  onTap: _openSePay,
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Satisfaction Rating Card (matching Image 1)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 14,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                const Text(
                  'Mức độ hài lòng với Finora?',
                  style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w700,
                    color: Color(0xff334155),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: List.generate(5, (index) {
                    final starIndex = index + 1;
                    final isFilled = starIndex <= _satisfactionRating;
                    return Semantics(
                      button: true,
                      selected: _satisfactionRating == starIndex,
                      label:
                          '$starIndex sao${_satisfactionRating == starIndex ? ', đang chọn' : ''}',
                      child: GestureDetector(
                        onTap: () {
                          setState(() {
                            _satisfactionRating = starIndex;
                          });
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(
                                'Cảm ơn bạn đã đánh giá $_satisfactionRating sao!',
                              ),
                              duration: const Duration(seconds: 2),
                            ),
                          );
                        },
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 6),
                          child: Icon(
                            isFilled
                                ? Icons.star_rounded
                                : Icons.star_outline_rounded,
                            size: 34,
                            color: isFilled
                                ? const Color(0xfff59e0b)
                                : const Color(0xffcbd5e1),
                          ),
                        ),
                      ),
                    );
                  }),
                ),
                const SizedBox(height: 12),
                const Divider(height: 1),
                const SizedBox(height: 10),
                Semantics(
                  button: true,
                  label: 'Mở lịch sử đánh giá',
                  child: GestureDetector(
                    onTap: () => _showSupportModal(
                      'Lịch sử đánh giá',
                      'Bạn chưa có lịch sử đánh giá nào trước đó.',
                    ),
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'Lịch sử đánh giá',
                          style: TextStyle(
                            fontSize: 12.5,
                            fontWeight: FontWeight.bold,
                            color: Color(0xffd97706),
                          ),
                        ),
                        SizedBox(width: 4),
                        Icon(
                          Icons.chevron_right_rounded,
                          size: 18,
                          color: Color(0xffd97706),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          // App Version (matching Image 1)
          const Center(
            child: Text(
              'Phiên bản: 10.12.52',
              style: TextStyle(
                fontSize: 12,
                color: Color(0xff94a3b8),
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          const SizedBox(height: 14),

          // Logout Button Card (matching Image 1)
          InkWell(
            onTap: widget.onLogout,
            borderRadius: BorderRadius.circular(18),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 14,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: const Row(
                children: [
                  Icon(
                    Icons.output_rounded,
                    color: Color(0xffef4444),
                    size: 20,
                  ),
                  SizedBox(width: 12),
                  Text(
                    'Đăng xuất',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: Color(0xffef4444),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildMenuItemTile({
    required IconData icon,
    required String title,
    String? subtitle,
    required Color iconColor,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 13),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(7),
              decoration: BoxDecoration(
                color: iconColor.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(icon, color: iconColor, size: 18),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xff1e293b),
                    ),
                  ),
                  if (subtitle != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      style: const TextStyle(
                        fontSize: 10.5,
                        color: Color(0xff64748b),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const Icon(
              Icons.chevron_right_rounded,
              color: Color(0xffd97706),
              size: 20,
            ),
          ],
        ),
      ),
    );
  }
}

class TransactionsPage extends StatefulWidget {
  const TransactionsPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<TransactionsPage> createState() => _TransactionsPageState();
}

class _TransactionsPageState extends State<TransactionsPage> {
  List items = [];
  bool loading = true;
  String? error;
  String selectedFilter = 'all'; // 'all', 'income', 'expense', 'transfer'
  final searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    load();
  }

  @override
  void dispose() {
    searchController.dispose();
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final x =
          await widget.api.request('GET', '/transactions?limit=100') as Map;
      items = x['items'] as List? ?? [];
      error = null;
    } catch (e) {
      error = presentableError(e);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  void form() => showModalBottomSheet(
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    context: context,
    builder: (_) =>
        _TransactionFormSheet(api: widget.api, onSuccess: () => load()),
  );

  double get _totalIncome {
    double sum = 0;
    for (final item in items) {
      if (item['type'] == 'income') {
        sum += double.tryParse(item['amount']?.toString() ?? '0') ?? 0;
      }
    }
    return sum;
  }

  double get _totalExpense {
    double sum = 0;
    for (final item in items) {
      if (item['type'] == 'expense') {
        sum += double.tryParse(item['amount']?.toString() ?? '0') ?? 0;
      }
    }
    return sum;
  }

  double get _netCashFlow => _totalIncome - _totalExpense;

  List get _filteredItems {
    final query = searchController.text.trim().toLowerCase();
    return items.where((x) {
      final type = x['type']?.toString() ?? '';
      if (selectedFilter == 'income' && type != 'income') return false;
      if (selectedFilter == 'expense' && type != 'expense') return false;
      if (selectedFilter == 'transfer' && type != 'transfer') return false;

      if (query.isNotEmpty) {
        final note = x['note']?.toString().toLowerCase() ?? '';
        final amountStr = x['amount']?.toString() ?? '';
        return note.contains(query) || amountStr.contains(query);
      }
      return true;
    }).toList();
  }

  Map<String, List> get _groupedItems {
    final Map<String, List> map = {};
    for (final item in _filteredItems) {
      final rawDate =
          item['occurredAt']?.toString() ?? item['createdAt']?.toString();
      final dateKey = _formatGroupDate(rawDate);
      map.putIfAbsent(dateKey, () => []).add(item);
    }
    return map;
  }

  String _formatGroupDate(String? raw) {
    if (raw == null || raw.isEmpty) return 'Khác';
    try {
      final dt = DateTime.parse(raw).toLocal();
      final now = DateTime.now();
      if (dt.year == now.year && dt.month == now.month && dt.day == now.day) {
        return 'Hôm nay - ${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}';
      }
      final yesterday = now.subtract(const Duration(days: 1));
      if (dt.year == yesterday.year &&
          dt.month == yesterday.month &&
          dt.day == yesterday.day) {
        return 'Hôm qua - ${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}';
      }
      return '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year}';
    } catch (_) {
      return 'Khác';
    }
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Lịch sử giao dịch',
    action: Row(
      children: [
        IconButton(
          onPressed: load,
          icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
        ),
        const SizedBox(width: 4),
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xfffbbf24),
            foregroundColor: const Color(0xff1c1917),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
            ),
          ),
          onPressed: form,
          icon: const Icon(Icons.add_rounded, size: 18),
          label: const Text(
            'Thêm mới',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
        ),
      ],
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: FinoraColors.primary),
          )
        : (error != null && items.isEmpty)
        ? FinoraEmptyState(
            title: 'Chưa thể tải giao dịch',
            message: 'Kiểm tra kết nối rồi thử lại.',
            icon: Icons.cloud_off_rounded,
            action: FilledButton.icon(
              onPressed: load,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Thử lại'),
            ),
          )
        : ListView(
            physics: const BouncingScrollPhysics(),
            padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
            children: [
              // KPI Analytics Cards Header
              Row(
                children: [
                  Expanded(
                    child: _buildMetricCard(
                      title: 'Tổng thu',
                      amount: _totalIncome,
                      isIncome: true,
                      icon: Icons.south_west_rounded,
                      accentColor: const Color(0xff4ade80),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _buildMetricCard(
                      title: 'Tổng chi',
                      amount: _totalExpense,
                      isIncome: false,
                      icon: Icons.north_east_rounded,
                      accentColor: const Color(0xfffb7185),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              _buildNetFlowCard(),
              const SizedBox(height: 16),

              // Search Bar & Filter Pills
              _buildSearchAndFilterSection(),
              const SizedBox(height: 16),

              if (error != null) ErrorBox(error!),

              // Grouped List Items
              if (_groupedItems.isEmpty)
                FinoraEmptyState(
                  title: searchController.text.isEmpty
                      ? 'Chưa có giao dịch'
                      : 'Không tìm thấy giao dịch',
                  message: searchController.text.isEmpty
                      ? 'Tạo giao dịch đầu tiên để theo dõi dòng tiền.'
                      : 'Thử đổi từ khoá hoặc bộ lọc.',
                  icon: Icons.receipt_long_outlined,
                  action: searchController.text.isEmpty
                      ? FilledButton.icon(
                          onPressed: form,
                          icon: const Icon(Icons.add_rounded),
                          label: const Text('Tạo giao dịch'),
                        )
                      : null,
                )
              else
                ..._groupedItems.entries.map((entry) {
                  final dateLabel = entry.key;
                  final list = entry.value;
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 8,
                        ),
                        child: Row(
                          children: [
                            Icon(
                              Icons.calendar_today_rounded,
                              size: 14,
                              color: Colors.white.withValues(alpha: 0.6),
                            ),
                            const SizedBox(width: 6),
                            Text(
                              dateLabel,
                              style: TextStyle(
                                color: Colors.white.withValues(alpha: 0.9),
                                fontSize: 13,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(10),
                              ),
                              child: Text(
                                '${list.length} giao dịch',
                                style: const TextStyle(
                                  color: Colors.white70,
                                  fontSize: 10.5,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      ...list.map((x) => _buildTransactionCard(x)),
                      const SizedBox(height: 8),
                    ],
                  );
                }),
            ],
          ),
  );

  Widget _buildMetricCard({
    required String title,
    required double amount,
    required bool isIncome,
    required IconData icon,
    required Color accentColor,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: accentColor.withValues(alpha: 0.25),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 14,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: accentColor.withValues(alpha: 0.12),
                  shape: BoxShape.circle,
                ),
                child: Icon(icon, color: accentColor, size: 12),
              ),
              const SizedBox(width: 6),
              Text(
                title,
                style: const TextStyle(
                  color: Color(0xff64748b),
                  fontSize: 10.5,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          FittedBox(
            fit: BoxFit.scaleDown,
            child: Text(
              '${isIncome ? '+' : '-'}${formatCurrency(amount)}',
              style: TextStyle(
                color: accentColor,
                fontSize: 14.5,
                fontWeight: FontWeight.w900,
                letterSpacing: -0.2,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNetFlowCard() {
    final net = _netCashFlow;
    final isPositive = net >= 0;
    final accentColor = isPositive
        ? const Color(0xff10b981)
        : const Color(0xfff43f5e);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: accentColor.withValues(alpha: 0.25),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 14,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: accentColor.withValues(alpha: 0.12),
              shape: BoxShape.circle,
            ),
            child: Icon(
              isPositive
                  ? Icons.trending_up_rounded
                  : Icons.trending_down_rounded,
              color: accentColor,
              size: 16,
            ),
          ),
          const SizedBox(width: 10),
          const Text(
            'Dòng tiền Ròng:',
            style: TextStyle(
              color: Color(0xff1e293b),
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const Spacer(),
          Text(
            '${isPositive ? '+' : ''}${formatCurrency(net)}',
            style: TextStyle(
              color: accentColor,
              fontSize: 14,
              fontWeight: FontWeight.w900,
              letterSpacing: -0.2,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSearchAndFilterSection() {
    return Column(
      children: [
        // Search Input Bar
        Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.05),
                blurRadius: 14,
                offset: const Offset(0, 3),
              ),
            ],
          ),
          child: TextField(
            controller: searchController,
            onChanged: (_) => setState(() {}),
            style: const TextStyle(
              color: Color(0xff0f172a),
              fontSize: 12.5,
              fontWeight: FontWeight.w600,
            ),
            decoration: InputDecoration(
              hintText: 'Tìm kiếm theo ghi chú, số tiền...',
              hintStyle: const TextStyle(
                color: Color(0xff94a3b8),
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              prefixIcon: const Icon(
                Icons.search_rounded,
                color: Color(0xffd97706),
                size: 18,
              ),
              suffixIcon: searchController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(
                        Icons.clear_rounded,
                        color: Color(0xff64748b),
                        size: 16,
                      ),
                      onPressed: () {
                        searchController.clear();
                        setState(() {});
                      },
                    )
                  : null,
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(vertical: 10),
            ),
          ),
        ),
        const SizedBox(height: 8),

        // Filter Pills
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: [
              _buildFilterPill('all', 'Tất cả', Icons.apps_rounded),
              const SizedBox(width: 6),
              _buildFilterPill('income', 'Thu nhập', Icons.south_west_rounded),
              const SizedBox(width: 6),
              _buildFilterPill('expense', 'Chi tiêu', Icons.north_east_rounded),
              const SizedBox(width: 6),
              _buildFilterPill(
                'transfer',
                'Chuyển tiền',
                Icons.swap_horiz_rounded,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildFilterPill(String filterKey, String label, IconData icon) {
    final isSelected = selectedFilter == filterKey;
    return InkWell(
      onTap: () => setState(() => selectedFilter = filterKey),
      borderRadius: BorderRadius.circular(14),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6.5),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xfffbbf24) : Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: isSelected
                ? const Color(0xffd97706)
                : Colors.white.withValues(alpha: 0.8),
            width: 1,
          ),
          boxShadow: [
            BoxShadow(
              color: isSelected
                  ? const Color(0x22d97706)
                  : Colors.black.withValues(alpha: 0.04),
              blurRadius: isSelected ? 10 : 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 13,
              color: isSelected
                  ? const Color(0xff1c1917)
                  : const Color(0xff64748b),
            ),
            const SizedBox(width: 5),
            Text(
              label,
              style: TextStyle(
                color: isSelected
                    ? const Color(0xff1c1917)
                    : const Color(0xff334155),
                fontSize: 11.5,
                fontWeight: isSelected ? FontWeight.w900 : FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTransactionCard(Map item) {
    final isIncome = item['type'] == 'income';
    final isTransfer = item['type'] == 'transfer';
    final name = item['name']?.toString().trim();
    final note = item['note']?.toString().trim();
    final title = (name != null && name.isNotEmpty)
        ? name
        : ((note != null && note.isNotEmpty)
              ? note
              : (isIncome
                    ? 'Thu nhập'
                    : (isTransfer ? 'Chuyển tiền' : 'Chi tiêu')));
    final amountVal = double.tryParse(item['amount']?.toString() ?? '0') ?? 0.0;
    final icon = isIncome
        ? Icons.south_west_rounded
        : (isTransfer ? Icons.swap_horiz_rounded : Icons.north_east_rounded);
    final color = isIncome
        ? const Color(0xff4ade80)
        : (isTransfer ? const Color(0xff38bdf8) : const Color(0xfffb7185));

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: FinoraListTile(
        icon: icon,
        iconColor: color,
        title: title,
        subtitle: _formatDate(
          item['occurredAt']?.toString() ?? item['createdAt']?.toString(),
        ),
        amount:
            "${isIncome ? '+' : (isTransfer ? '' : '-')}${formatCurrency(amountVal)}",
      ),
    );
  }
}

class _TransactionFormSheet extends StatefulWidget {
  const _TransactionFormSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_TransactionFormSheet> createState() => _TransactionFormSheetState();
}

class _TransactionFormSheetState extends State<_TransactionFormSheet> {
  String type = 'expense';
  String? selectedAccountId;
  List<Map<String, dynamic>> accounts = [];
  bool loadingAccounts = true;

  final nameController = TextEditingController();
  final amountController = TextEditingController();
  final noteController = TextEditingController();
  DateTime selectedDate = DateTime.now();
  bool submitting = false;

  @override
  void initState() {
    super.initState();
    _loadAccounts();
  }

  Future<void> _loadAccounts() async {
    try {
      final res = await widget.api.request('GET', '/accounts');
      final list = res is List
          ? res
          : (res is Map ? res['items'] as List? ?? [] : []);
      setState(() {
        accounts = List<Map<String, dynamic>>.from(
          list.map((e) => Map<String, dynamic>.from(e as Map)),
        );
        if (accounts.isNotEmpty) {
          selectedAccountId = accounts.first['id']?.toString();
        }
        loadingAccounts = false;
      });
    } catch (_) {
      if (mounted) setState(() => loadingAccounts = false);
    }
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: selectedDate,
      firstDate: DateTime(2020),
      lastDate: DateTime(2030),
      builder: (context, child) {
        return Theme(
          data: ThemeData.light().copyWith(
            colorScheme: const ColorScheme.light(
              primary: Color(0xffd97706),
              onPrimary: Colors.white,
              surface: Colors.white,
              onSurface: Color(0xff0f172a),
            ),
          ),
          child: child!,
        );
      },
    );
    if (picked != null) {
      setState(() => selectedDate = picked);
    }
  }

  Future<void> _submit() async {
    final rawName = nameController.text.trim();
    final rawAmount = amountController.text.trim();
    if (rawAmount.isEmpty) {
      showError(context, 'Vui lòng nhập số tiền.');
      return;
    }

    final parsedVal = parseSmartAmount(rawAmount);
    if (parsedVal <= 0) {
      showError(context, 'Số tiền không hợp lệ.');
      return;
    }

    if (selectedAccountId == null || selectedAccountId!.isEmpty) {
      showError(context, 'Vui lòng chọn hoặc tạo tài khoản trước.');
      return;
    }

    final txnName = rawName.isNotEmpty
        ? rawName
        : (type == 'income' ? 'Thu nhập' : 'Chi tiêu');

    setState(() => submitting = true);
    try {
      await widget.api.request('POST', '/transactions', {
        'accountId': selectedAccountId,
        'name': txnName,
        'type': type,
        'amount': parsedVal.toStringAsFixed(0),
        'currency': 'VND',
        'note': noteController.text.trim(),
        'occurredAt': selectedDate.toUtc().toIso8601String(),
        'status': 'posted',
      });
      if (mounted) {
        if (Navigator.of(context).canPop()) {
          Navigator.of(context).pop();
        }
        widget.onSuccess();
      }
    } catch (e) {
      if (mounted) showError(context, presentableError(e));
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    nameController.dispose();
    amountController.dispose();
    noteController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final parsedAmount = parseSmartAmount(amountController.text);
    final isToday =
        DateTime.now().year == selectedDate.year &&
        DateTime.now().month == selectedDate.month &&
        DateTime.now().day == selectedDate.day;
    final dateLabel = isToday
        ? 'Hôm nay (${selectedDate.day.toString().padLeft(2, '0')}/${selectedDate.month.toString().padLeft(2, '0')}/${selectedDate.year})'
        : '${selectedDate.day.toString().padLeft(2, '0')}/${selectedDate.month.toString().padLeft(2, '0')}/${selectedDate.year}';

    return FinoraSheet(
      title: 'Tạo giao dịch',
      subtitle:
          'Ghi nhận thu chi nhanh. Số tiền được tự động nội suy thông minh.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => type = 'expense'),
                  borderRadius: BorderRadius.circular(16),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: type == 'expense'
                          ? const Color(0xffffe4e6)
                          : const Color(0xfff1f5f9),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: type == 'expense'
                            ? const Color(0xfff43f5e)
                            : const Color(0xffcbd5e1),
                        width: type == 'expense' ? 2 : 1,
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.north_east_rounded,
                          color: type == 'expense'
                              ? const Color(0xffbe123c)
                              : const Color(0xff64748b),
                          size: 18,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Chi tiêu',
                          style: TextStyle(
                            color: type == 'expense'
                                ? const Color(0xffbe123c)
                                : const Color(0xff475569),
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => type = 'income'),
                  borderRadius: BorderRadius.circular(16),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: type == 'income'
                          ? const Color(0xffdcfce7)
                          : const Color(0xfff1f5f9),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: type == 'income'
                            ? const Color(0xff22c55e)
                            : const Color(0xffcbd5e1),
                        width: type == 'income' ? 2 : 1,
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.south_west_rounded,
                          color: type == 'income'
                              ? const Color(0xff15803d)
                              : const Color(0xff64748b),
                          size: 18,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Thu nhập',
                          style: TextStyle(
                            color: type == 'income'
                                ? const Color(0xff15803d)
                                : const Color(0xff475569),
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          if (loadingAccounts)
            const Center(
              child: CircularProgressIndicator(color: Color(0xffd97706)),
            )
          else if (accounts.isEmpty)
            const Text(
              'Chưa có tài khoản nào. Vui lòng tạo tài khoản trước.',
              style: TextStyle(
                color: Color(0xffef4444),
                fontSize: 13,
                fontWeight: FontWeight.bold,
              ),
            )
          else
            DropdownButtonFormField<String>(
              initialValue: selectedAccountId,
              dropdownColor: Colors.white,
              style: const TextStyle(
                color: Color(0xff0f172a),
                fontWeight: FontWeight.w800,
                fontSize: 14,
              ),
              items: accounts.map((acc) {
                return DropdownMenuItem<String>(
                  value: acc['id']?.toString(),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.account_balance_wallet_rounded,
                        color: Color(0xffd97706),
                        size: 18,
                      ),
                      const SizedBox(width: 8),
                      Text(acc['name']?.toString() ?? 'Tài khoản'),
                    ],
                  ),
                );
              }).toList(),
              onChanged: (v) => setState(() => selectedAccountId = v),
              decoration: InputDecoration(
                labelText: 'Chọn tài khoản',
                labelStyle: const TextStyle(
                  color: Color(0xff64748b),
                  fontWeight: FontWeight.w600,
                ),
                filled: true,
                fillColor: const Color(0xfff8fafc),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffcbd5e1),
                    width: 1.2,
                  ),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffcbd5e1),
                    width: 1.2,
                  ),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffd97706),
                    width: 1.8,
                  ),
                ),
              ),
            ),
          const SizedBox(height: 14),
          _CustomGlassTextField(
            controller: nameController,
            labelText: 'Tên giao dịch (VD: Lương tháng, Cà phê...)',
            icon: Icons.title_rounded,
          ),
          // Quick Amount Suggestion Chips
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                _buildQuickAmountChip('100k', '100k'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('200k', '200k'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('500k', '500k'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('1M', '1tr'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('2M', '2tr'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('5M', '5tr'),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _CustomGlassTextField(
            controller: amountController,
            keyboardType: TextInputType.text,
            labelText: 'Số tiền (vd: 100k, 1tr, 30t, 30m)',
            icon: Icons.attach_money_rounded,
            onChanged: (_) => setState(() {}),
          ),
          if (amountController.text.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 6, left: 12),
              child: Row(
                children: [
                  const Icon(
                    Icons.auto_awesome_rounded,
                    color: Color(0xffd97706),
                    size: 14,
                  ),
                  const SizedBox(width: 4),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xfffef3c7),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      '= ${formatCurrency(parsedAmount)}',
                      style: const TextStyle(
                        color: Color(0xffb45309),
                        fontWeight: FontWeight.w900,
                        fontSize: 13,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 14),
          InkWell(
            onTap: _pickDate,
            borderRadius: BorderRadius.circular(16),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                color: const Color(0xfff8fafc),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xffcbd5e1), width: 1.2),
              ),
              child: Row(
                children: [
                  const Icon(
                    Icons.calendar_today_rounded,
                    color: Color(0xffd97706),
                    size: 20,
                  ),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Ngày thực hiện',
                        style: TextStyle(
                          color: Color(0xff64748b),
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        dateLabel,
                        style: const TextStyle(
                          color: Color(0xff0f172a),
                          fontWeight: FontWeight.w800,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  const Spacer(),
                  const Icon(
                    Icons.edit_calendar_rounded,
                    color: Color(0xff94a3b8),
                    size: 18,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 14),
          _CustomGlassTextField(
            controller: noteController,
            labelText: 'Ghi chú (Tùy chọn)',
            icon: Icons.note_alt_rounded,
          ),
          const SizedBox(height: 18),
          _AnimatedGoldButton(
            busy: submitting,
            label: 'Lưu giao dịch',
            onTap: _submit,
          ),
        ],
      ),
    );
  }

  Widget _buildQuickAmountChip(String label, String valueToSet) {
    return InkWell(
      onTap: () {
        setState(() {
          amountController.text = valueToSet;
        });
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        decoration: BoxDecoration(
          color: const Color(0xfffef3c7),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xfffde68a)),
        ),
        child: Text(
          '+$label',
          style: const TextStyle(
            color: Color(0xffb45309),
            fontSize: 11.5,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}

class _AssetCreateSheet extends StatefulWidget {
  const _AssetCreateSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_AssetCreateSheet> createState() => _AssetCreateSheetState();
}

class _AssetCreateSheetState extends State<_AssetCreateSheet> {
  final name = TextEditingController();
  String assetType = 'other';
  bool submitting = false;

  Future<void> _submit() async {
    if (name.text.trim().isEmpty) {
      showError(context, 'Nhập tên tài sản.');
      return;
    }
    setState(() => submitting = true);
    try {
      await widget.api.request('POST', '/assets', {
        'name': name.text.trim(),
        'assetType': assetType,
      });
      if (!mounted) return;
      Navigator.pop(context);
      widget.onSuccess();
    } catch (error) {
      if (mounted) showError(context, presentableError(error));
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    name.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.lg,
        FinoraSpace.sm,
        FinoraSpace.lg,
        FinoraSpace.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surfaceElevated,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: const BoxDecoration(
                  color: FinoraColors.borderStrong,
                  borderRadius: FinoraRadius.full,
                ),
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            const Text('Thêm tài sản', style: FinoraTypography.h3),
            const SizedBox(height: FinoraSpace.xs),
            Text(
              'Bạn có thể bổ sung định giá sau khi tạo.',
              style: FinoraTypography.bodySmall.copyWith(
                color: FinoraColors.textSecondary,
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            TextField(
              controller: name,
              textCapitalization: TextCapitalization.words,
              decoration: const InputDecoration(labelText: 'Tên tài sản'),
            ),
            const SizedBox(height: FinoraSpace.sm),
            DropdownButtonFormField<String>(
              initialValue: assetType,
              decoration: const InputDecoration(labelText: 'Loại tài sản'),
              items: const [
                DropdownMenuItem(value: 'other', child: Text('Khác')),
                DropdownMenuItem(
                  value: 'gold',
                  child: Text('Vàng / kim loại quý'),
                ),
                DropdownMenuItem(value: 'vehicle', child: Text('Phương tiện')),
                DropdownMenuItem(
                  value: 'collectible',
                  child: Text('Tài sản sưu tầm'),
                ),
              ],
              onChanged: submitting
                  ? null
                  : (value) => setState(() => assetType = value ?? 'other'),
            ),
            const SizedBox(height: FinoraSpace.lg),
            FilledButton.icon(
              onPressed: submitting ? null : _submit,
              icon: submitting
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Icon(Icons.add_business_rounded),
              label: const Text('Tạo tài sản'),
            ),
          ],
        ),
      ),
    ),
  );
}

class _TransferFormSheet extends StatefulWidget {
  const _TransferFormSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_TransferFormSheet> createState() => _TransferFormSheetState();
}

class _TransferFormSheetState extends State<_TransferFormSheet> {
  final amountController = TextEditingController();
  final noteController = TextEditingController();
  List<Map<String, dynamic>> accounts = [];
  String? fromAccountId;
  String? toAccountId;
  bool loading = true;
  bool submitting = false;

  @override
  void initState() {
    super.initState();
    _loadAccounts();
  }

  Future<void> _loadAccounts() async {
    try {
      final result = await widget.api.request('GET', '/accounts');
      final rows = result is List
          ? result
          : (result is Map ? result['items'] as List? ?? [] : []);
      if (!mounted) return;
      setState(() {
        accounts = rows
            .map((item) => Map<String, dynamic>.from(item as Map))
            .toList();
        if (accounts.length >= 2) {
          fromAccountId = accounts.first['id']?.toString();
          toAccountId = accounts[1]['id']?.toString();
        }
        loading = false;
      });
    } catch (error) {
      if (mounted) {
        setState(() => loading = false);
        showError(context, presentableError(error));
      }
    }
  }

  Future<void> _submit() async {
    final amount = parseSmartAmount(amountController.text);
    if (fromAccountId == null ||
        toAccountId == null ||
        fromAccountId == toAccountId) {
      showError(context, 'Chọn hai tài khoản khác nhau.');
      return;
    }
    if (amount <= 0) {
      showError(context, 'Nhập số tiền hợp lệ.');
      return;
    }
    setState(() => submitting = true);
    try {
      await widget.api.request('POST', '/transfers', {
        'fromAccountId': fromAccountId,
        'toAccountId': toAccountId,
        'amount': amount.toStringAsFixed(0),
        'currency': 'VND',
        'note': noteController.text.trim(),
        'occurredAt': DateTime.now().toUtc().toIso8601String(),
      });
      if (!mounted) return;
      Navigator.pop(context);
      widget.onSuccess();
    } catch (error) {
      if (mounted) showError(context, presentableError(error));
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    amountController.dispose();
    noteController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.lg,
        FinoraSpace.sm,
        FinoraSpace.lg,
        FinoraSpace.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surfaceElevated,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: MainAxisSize.min,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: const BoxDecoration(
                  color: FinoraColors.borderStrong,
                  borderRadius: FinoraRadius.full,
                ),
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            const Text('Chuyển tiền', style: FinoraTypography.h3),
            const SizedBox(height: FinoraSpace.xs),
            Text(
              'Chuyển tiền nội bộ giữa các tài khoản của bạn.',
              style: FinoraTypography.bodySmall.copyWith(
                color: FinoraColors.textSecondary,
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            if (loading)
              const Center(
                child: CircularProgressIndicator(color: FinoraColors.primary),
              )
            else if (accounts.length < 2)
              FinoraEmptyState(
                title: 'Cần ít nhất hai tài khoản',
                message: 'Tạo thêm tài khoản trước khi chuyển tiền nội bộ.',
                icon: Icons.account_balance_wallet_outlined,
              )
            else ...[
              DropdownButtonFormField<String>(
                initialValue: fromAccountId,
                decoration: const InputDecoration(labelText: 'Từ tài khoản'),
                items: accounts
                    .map(
                      (a) => DropdownMenuItem(
                        value: a['id']?.toString(),
                        child: Text(a['name']?.toString() ?? 'Tài khoản'),
                      ),
                    )
                    .toList(),
                onChanged: submitting
                    ? null
                    : (value) => setState(() => fromAccountId = value),
              ),
              const SizedBox(height: FinoraSpace.sm),
              DropdownButtonFormField<String>(
                initialValue: toAccountId,
                decoration: const InputDecoration(labelText: 'Đến tài khoản'),
                items: accounts
                    .map(
                      (a) => DropdownMenuItem(
                        value: a['id']?.toString(),
                        child: Text(a['name']?.toString() ?? 'Tài khoản'),
                      ),
                    )
                    .toList(),
                onChanged: submitting
                    ? null
                    : (value) => setState(() => toAccountId = value),
              ),
              const SizedBox(height: FinoraSpace.sm),
              TextField(
                controller: amountController,
                keyboardType: TextInputType.text,
                decoration: const InputDecoration(
                  labelText: 'Số tiền',
                  hintText: 'Ví dụ: 500k hoặc 1tr',
                  prefixIcon: Icon(Icons.payments_outlined),
                ),
              ),
              const SizedBox(height: FinoraSpace.sm),
              TextField(
                controller: noteController,
                maxLines: 2,
                decoration: const InputDecoration(
                  labelText: 'Ghi chú',
                  prefixIcon: Icon(Icons.notes_rounded),
                ),
              ),
              const SizedBox(height: FinoraSpace.lg),
              FilledButton.icon(
                onPressed: submitting ? null : _submit,
                icon: submitting
                    ? const SizedBox.square(
                        dimension: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.swap_horiz_rounded),
                label: const Text('Xác nhận chuyển tiền'),
              ),
            ],
          ],
        ),
      ),
    ),
  );
}

double parseSmartAmount(String rawInput, {String currency = 'VND'}) {
  return currency.toUpperCase() == 'VND'
      ? parseVietnameseMoneyInput(rawInput)
      : (double.tryParse(rawInput.trim()) ?? 0);
}

String formatCurrency(double amount, {String currency = 'VND'}) {
  final intVal = amount.round();
  final formattedStr = intVal.toString().replaceAllMapped(
    RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'),
    (Match m) => '${m[1]}.',
  );
  return '$formattedStr $currency';
}

class BudgetPage extends StatefulWidget {
  const BudgetPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<BudgetPage> createState() => _BudgetPageState();
}

class _BudgetPageState extends State<BudgetPage> {
  final period = TextEditingController(
        text:
            '${DateTime.now().year}-${DateTime.now().month.toString().padLeft(2, '0')}',
      ),
      category = TextEditingController(),
      limit = TextEditingController();
  dynamic data;
  String? err;
  bool loading = false;
  Future<void> load() async {
    setState(() => loading = true);
    try {
      data = await widget.api.request('GET', '/budgets/${period.text}');
      err = null;
    } catch (e) {
      err = presentableError(e);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> save() async {
    final parsedLimit = parseSmartAmount(limit.text);
    if (parsedLimit <= 0) {
      showError(context, 'Nhập hạn mức hợp lệ, ví dụ 30tr.');
      return;
    }
    try {
      await widget.api.request('PUT', '/budgets/${period.text}', {
        'categoryId': category.text,
        'limit': parsedLimit.toStringAsFixed(0),
      });
      load();
    } catch (e) {
      if (mounted) showError(context, presentableError(e));
    }
  }

  @override
  void dispose() {
    period.dispose();
    category.dispose();
    limit.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Ngân sách',
    child: ListView(
      children: [
        const _ScreenIntro(
          'Đặt giới hạn thông minh để kế hoạch tài chính luôn đúng hướng.',
        ),
        FinoraSurface(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Tra cứu kỳ ngân sách',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  fontSize: 16,
                ),
              ),
              const SizedBox(height: 14),
              _CustomGlassTextField(
                controller: period,
                labelText: 'Kỳ ngân sách (YYYY-MM)',
                icon: Icons.calendar_month_rounded,
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Color(0xfff7d070)),
                  foregroundColor: const Color(0xfff7d070),
                ),
                onPressed: loading ? null : load,
                icon: const Icon(Icons.search_rounded),
                label: const Text('Xem ngân sách'),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        FinoraSurface(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Thiết lập hạn mức',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  fontSize: 16,
                ),
              ),
              const SizedBox(height: 14),
              _CustomGlassTextField(
                controller: category,
                labelText: 'Mã danh mục',
                icon: Icons.sell_outlined,
              ),
              const SizedBox(height: 12),
              _CustomGlassTextField(
                controller: limit,
                keyboardType: TextInputType.text,
                labelText: 'Hạn mức (vd: 30tr)',
                icon: Icons.account_balance_wallet_outlined,
              ),
              const SizedBox(height: 14),
              FilledButton.icon(
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xfff7d070),
                  foregroundColor: const Color(0xff200733),
                ),
                onPressed: save,
                icon: const Icon(Icons.check_rounded),
                label: const Text(
                  'Lưu hạn mức',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
        ),
        if (err != null) ErrorBox(err!),
        if (data != null) _buildBudgetResults(),
      ],
    ),
  );

  Widget _buildBudgetResults() {
    final response = data is Map
        ? Map<String, dynamic>.from(data as Map)
        : <String, dynamic>{};
    final rows = (response['rows'] as List?) ?? const [];
    return Padding(
      padding: const EdgeInsets.only(top: FinoraSpace.md),
      child: FinoraSurface(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Ngân sách ${response['period'] ?? period.text}',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: FinoraSpace.sm),
            if (rows.isEmpty)
              const Text(
                'Chưa có hạn mức cho kỳ này.',
                style: TextStyle(color: Colors.white70),
              )
            else
              ...rows.map((raw) {
                final row = Map<String, dynamic>.from(raw as Map);
                final limitValue =
                    double.tryParse(row['limit']?.toString() ?? '') ?? 0;
                final spentValue =
                    double.tryParse(row['spent']?.toString() ?? '') ?? 0;
                final progress = limitValue > 0
                    ? (spentValue / limitValue).clamp(0.0, 1.0)
                    : 0.0;
                final isOver = limitValue > 0 && spentValue > limitValue;
                final currency = row['currency']?.toString() ?? 'VND';
                return Padding(
                  padding: const EdgeInsets.only(bottom: FinoraSpace.md),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              'Danh mục ${row['categoryId'] ?? 'Chưa phân loại'}',
                              style: const TextStyle(
                                color: Colors.white,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                          Text(
                            '${formatCurrency(spentValue, currency: currency)} / ${formatCurrency(limitValue, currency: currency)}',
                            style: TextStyle(
                              color: isOver
                                  ? FinoraColors.accentAmber
                                  : FinoraColors.goldLight,
                              fontWeight: FontWeight.w800,
                              fontSize: 12,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: FinoraSpace.xs),
                      LinearProgressIndicator(
                        value: progress,
                        minHeight: 7,
                        borderRadius: const BorderRadius.all(
                          Radius.circular(99),
                        ),
                        backgroundColor: Colors.white.withValues(alpha: 0.14),
                        valueColor: AlwaysStoppedAnimation(
                          isOver ? FinoraColors.danger : FinoraColors.success,
                        ),
                      ),
                    ],
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }
}

class SePayConnectionPage extends StatefulWidget {
  const SePayConnectionPage({super.key, required this.api});
  final ApiClient api;

  @override
  State<SePayConnectionPage> createState() => _SePayConnectionPageState();
}

class _SePayConnectionPageState extends State<SePayConnectionPage> {
  List<dynamic> _accounts = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
    });
    try {
      final data = await widget.api.request('GET', '/me/sepay') as Map;
      _accounts = (data['bankAccounts'] as List?) ?? const [];
      _error = null;
    } catch (error) {
      _error = presentableError(error);
    }
    if (mounted) {
      setState(() {
        _loading = false;
      });
    }
  }

  Future<void> _link() async {
    try {
      final data =
          await widget.api.request('POST', '/me/sepay/link-session') as Map;
      final url = data['hosted_link_url']?.toString() ?? '';
      if (url.isEmpty) {
        throw Exception('SePay không trả về đường dẫn liên kết.');
      }
      if (!mounted) {
        return;
      }
      final completion = await Navigator.of(context).push<Map<String, dynamic>>(
        MaterialPageRoute(builder: (_) => _HostedLinkPage(url: url)),
      );
      final accountNumber =
          completion?['account_number']?.toString() ??
          completion?['accountNumber']?.toString() ??
          '';
      if (accountNumber.isNotEmpty) {
        await widget.api.request('POST', '/me/sepay/bank-accounts/sync', {
          'accountNumber': accountNumber,
        });
      }
      await _load();
    } catch (error) {
      if (mounted) showError(context, presentableError(error));
    }
  }

  Future<void> _mapAccount(Map account) async {
    try {
      final accounts = await widget.api.request('GET', '/accounts') as List;
      if (!mounted) return;
      final selected = await showModalBottomSheet<Map>(
        context: context,
        builder: (context) => SafeArea(
          child: ListView(
            shrinkWrap: true,
            children: [
              const ListTile(
                title: Text(
                  'Chọn tài khoản Finora',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              const Padding(
                padding: EdgeInsets.fromLTRB(16, 0, 16, 12),
                child: Text(
                  'Nếu không gian làm việc có nhiều thành viên, giao dịch ngân hàng sẽ được chia sẻ theo quyền của từng thành viên.',
                  style: TextStyle(color: Color(0xff64748b), fontSize: 12),
                ),
              ),
              ...accounts.map(
                (item) => ListTile(
                  leading: const Icon(Icons.account_balance_wallet_outlined),
                  title: Text(item['name']?.toString() ?? 'Tài khoản'),
                  subtitle: Text(item['currency']?.toString() ?? 'VND'),
                  onTap: () => Navigator.pop(
                    context,
                    Map<String, dynamic>.from(item as Map),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
      if (selected == null) return;
      await widget.api.request(
        'POST',
        '/me/sepay/bank-accounts/${account['id']}/map',
        {'accountId': selected['id']},
      );
      await _load();
    } catch (error) {
      if (mounted) showError(context, presentableError(error));
    }
  }

  Future<void> _unlink(Map account) async {
    final accepted = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Ngắt liên kết?'),
        content: const Text(
          'Giao dịch đã ghi nhận vẫn được giữ lại. Bạn có thể liên kết lại sau.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Ngắt liên kết'),
          ),
        ],
      ),
    );
    if (accepted != true) return;
    try {
      await widget.api.request(
        'POST',
        '/me/sepay/bank-accounts/${account['id']}/unlink',
      );
      await _load();
    } catch (error) {
      if (mounted) showError(context, presentableError(error));
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Liên kết SePay')),
    body: _loading
        ? const Center(child: CircularProgressIndicator())
        : RefreshIndicator(
            onRefresh: _load,
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: const Color(0xffeff6ff),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: const Row(
                    children: [
                      Icon(Icons.shield_outlined, color: Color(0xff2563eb)),
                      SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Finora chỉ có quyền đọc giao dịch và số dư. Chúng tôi không lưu mật khẩu hoặc OTP của bạn.',
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                FilledButton.icon(
                  onPressed: _link,
                  icon: const Icon(Icons.add_link),
                  label: const Text('Liên kết tài khoản ngân hàng'),
                ),
                if (_error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: ErrorBox(_error!),
                  ),
                const SizedBox(height: 20),
                const Text(
                  'TÀI KHOẢN ĐÃ LIÊN KẾT',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Color(0xff475569),
                  ),
                ),
                const SizedBox(height: 8),
                if (_accounts.isEmpty)
                  const EmptyState(
                    'Chưa có tài khoản ngân hàng nào được liên kết.',
                  ),
                ..._accounts.map((raw) {
                  final account = Map<String, dynamic>.from(raw as Map);
                  final mapping = account['mapping'];
                  final canIn = account['supportsIn'] == true;
                  final canOut = account['supportsOut'] == true;
                  return Card(
                    child: Padding(
                      padding: const EdgeInsets.all(14),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(
                                Icons.account_balance,
                                color: Color(0xff2563eb),
                              ),
                              const SizedBox(width: 10),
                              Expanded(
                                child: Text(
                                  account['bankName']?.toString().isNotEmpty ==
                                          true
                                      ? account['bankName'].toString()
                                      : account['bankCode']?.toString() ??
                                            'Ngân hàng',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                              Text(_connectionStatusLabel(account['status'])),
                            ],
                          ),
                          const SizedBox(height: 8),
                          Text(
                            account['accountNumberMasked']?.toString() ??
                                '••••',
                          ),
                          const SizedBox(height: 10),
                          Wrap(
                            spacing: 8,
                            children: [
                              Chip(
                                label: Text(
                                  canIn ? 'Tiền vào' : 'Không hỗ trợ tiền vào',
                                ),
                              ),
                              Chip(
                                label: Text(
                                  canOut ? 'Tiền ra' : 'Không hỗ trợ tiền ra',
                                ),
                              ),
                            ],
                          ),
                          const Divider(),
                          if (mapping == null)
                            Align(
                              alignment: Alignment.centerRight,
                              child: TextButton.icon(
                                onPressed: () => _mapAccount(account),
                                icon: const Icon(Icons.link),
                                label: const Text('Gán tài khoản Finora'),
                              ),
                            )
                          else
                            Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    'Đã gán vào tài khoản Finora',
                                    style: TextStyle(
                                      color: Colors.green.shade700,
                                    ),
                                  ),
                                ),
                                TextButton(
                                  onPressed: () => _unlink(account),
                                  child: const Text('Ngắt liên kết'),
                                ),
                              ],
                            ),
                        ],
                      ),
                    ),
                  );
                }),
              ],
            ),
          ),
  );
}

class _HostedLinkPage extends StatefulWidget {
  const _HostedLinkPage({required this.url});
  final String url;
  @override
  State<_HostedLinkPage> createState() => _HostedLinkPageState();
}

class _HostedLinkPageState extends State<_HostedLinkPage> {
  late final WebViewController _controller;
  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..addJavaScriptChannel(
        'BankHubEvents',
        onMessageReceived: (message) {
          try {
            final event = jsonDecode(message.message) as Map;
            final name = event['event']?.toString() ?? '';
            final metadata = Map<String, dynamic>.from(
              (event['metadata'] as Map?) ?? const {},
            );
            if (name == 'FINISHED_BANK_ACCOUNT_LINK') {
              Navigator.of(context).pop(metadata);
            } else if (name == 'BANKHUB_CLOSE_LINK' ||
                name == 'FINISHED_BANK_ACCOUNT_UNLINK') {
              Navigator.of(context).pop();
            } else if (name == 'BANKHUB_TOKEN_EXPIRED' ||
                name == 'BANKHUB_SESSION_EXPIRED') {
              showError(context, 'Phiên liên kết SePay đã hết hạn.');
              Navigator.of(context).pop();
            }
          } catch (_) {
            // Ignore malformed cross-window messages; they are not provider data.
          }
        },
      )
      ..setNavigationDelegate(
        NavigationDelegate(
          onPageFinished: (_) => _controller.runJavaScript('''
            (function () {
              if (window.__finoraBankHubListener) return;
              window.__finoraBankHubListener = true;
              window.addEventListener('message', function (event) {
                var host = '';
                try { host = new URL(event.origin).hostname; } catch (_) { return; }
                if (host !== 'bankhub.sepay.vn' && !host.endsWith('.sepay.vn')) return;
                var data = event.data;
                if (typeof data === 'string') { try { data = JSON.parse(data); } catch (_) { return; } }
                if (data && data.event) BankHubEvents.postMessage(JSON.stringify(data));
              });
            })();
          '''),
        ),
      )
      ..loadRequest(Uri.parse(widget.url));
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Kết nối ngân hàng')),
    body: WebViewWidget(controller: _controller),
  );
}

class BankPage extends StatefulWidget {
  const BankPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<BankPage> createState() => _BankPageState();
}

class _BankPageState extends State<BankPage> {
  List feed = [];
  String? err;
  bool loading = true;
  String state = 'needs_review';
  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      final data =
          await widget.api.request('GET', '/me/bank-feed?state=$state') as Map;
      feed = (data['items'] as List?) ?? const [];
      err = null;
    } catch (e) {
      err = presentableError(e);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _action(
    String id,
    String action, [
    Map<String, dynamic>? payload,
  ]) async {
    try {
      await widget.api.request('POST', '/bank-feed/$id/$action', payload);
      await load();
    } catch (e) {
      if (mounted) showError(context, presentableError(e));
    }
  }

  Future<void> _edit(Map item) async {
    final result = await showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _BankFeedEditSheet(api: widget.api, item: item),
    );
    if (result != null) await _action(item['id'].toString(), 'correct', result);
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Cần kiểm tra',
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Rà soát giao dịch trước khi Finora ghi nhận vào sổ.',
              ),
              if (err != null) ErrorBox(err!),
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  children: [
                    for (final tab in const [
                      ('needs_review', 'Cần kiểm tra'),
                      ('ai_tagged', 'AI đã gắn'),
                      ('confirmed', 'Đã xác nhận'),
                      ('ignored', 'Bỏ qua'),
                    ])
                      Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: ChoiceChip(
                          label: Text(tab.$2),
                          selected: state == tab.$1,
                          onSelected: (_) {
                            setState(() => state = tab.$1);
                            load();
                          },
                        ),
                      ),
                  ],
                ),
              ),
              const SizedBox(height: 14),
              ...feed.map((raw) {
                final item = Map<String, dynamic>.from(raw as Map);
                final inbound = item['direction'] == 'in';
                final suggestions = (item['suggestions'] as List?) ?? const [];
                final suggestion = suggestions.isEmpty
                    ? null
                    : Map<String, dynamic>.from(suggestions.first as Map);
                return Card(
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              inbound ? Icons.south_west : Icons.north_east,
                              color: inbound ? Colors.green : Colors.red,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                inbound ? 'Thu' : 'Chi',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: inbound ? Colors.green : Colors.red,
                                ),
                              ),
                            ),
                            Text(
                              '${item['amount'] ?? ''} ${item['currency'] ?? 'VND'}',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(_redactBankContent(item['description'])),
                        Text(
                          'Ngân hàng · ${item['occurredAt']?.toString().replaceFirst('T', ' ').split('.').first ?? ''}',
                          style: const TextStyle(
                            color: Color(0xff64748b),
                            fontSize: 12,
                          ),
                        ),
                        if (suggestion != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 10),
                            child: Text(
                              '${suggestion['suggestedName'] ?? 'Gợi ý AI'} · ${suggestion['reason'] ?? ''} · ${_confidenceLabel(suggestion['confidence'])}',
                              style: const TextStyle(
                                color: Color(0xff2563eb),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        if (suggestion == null)
                          const Padding(
                            padding: EdgeInsets.only(top: 10),
                            child: Text(
                              'Chưa phân loại',
                              style: TextStyle(
                                color: Color(0xff64748b),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        if (state == 'needs_review' || state == 'ai_tagged')
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              TextButton(
                                onPressed: () =>
                                    _action(item['id'].toString(), 'ignore'),
                                child: const Text('Bỏ qua'),
                              ),
                              TextButton(
                                onPressed: () => _edit(item),
                                child: const Text('Sửa'),
                              ),
                              FilledButton(
                                onPressed: () =>
                                    _action(item['id'].toString(), 'confirm'),
                                child: const Text('Xác nhận'),
                              ),
                            ],
                          ),
                      ],
                    ),
                  ),
                );
              }),
              if (feed.isEmpty)
                const EmptyState('Không có giao dịch trong mục này.'),
            ],
          ),
  );
}

String _redactBankContent(dynamic value) {
  final text = value?.toString().trim() ?? '';
  if (text.isEmpty) return 'Giao dịch ngân hàng';
  final masked = text.replaceAll(RegExp(r'\b\d{6,}\b'), '••••');
  final runes = masked.runes.toList();
  return runes.length <= 90
      ? masked
      : '${String.fromCharCodes(runes.take(90))}…';
}

String _confidenceLabel(dynamic value) {
  final confidence = value is num
      ? value.toDouble()
      : double.tryParse(value?.toString() ?? '') ?? 0;
  if (confidence >= 90) return 'Khớp cao ${confidence.round()}%';
  if (confidence > 0) return 'Cần xác nhận ${confidence.round()}%';
  return 'Chưa phân loại';
}

String _connectionStatusLabel(dynamic value) {
  return switch (value?.toString()) {
    'active' => 'Đang kết nối',
    'pending' => 'Đang chờ',
    'error' => 'Cần kiểm tra',
    'revoked' => 'Đã ngắt kết nối',
    null || '' => 'Chưa xác định',
    _ => 'Trạng thái: ${value.toString()}',
  };
}

class _BankFeedEditSheet extends StatefulWidget {
  const _BankFeedEditSheet({required this.api, required this.item});
  final ApiClient api;
  final Map item;
  @override
  State<_BankFeedEditSheet> createState() => _BankFeedEditSheetState();
}

class _BankFeedEditSheetState extends State<_BankFeedEditSheet> {
  late final TextEditingController _name = TextEditingController(
    text: widget.item['description']?.toString() ?? '',
  );
  final _category = TextEditingController();
  final _note = TextEditingController();
  bool _remember = false;
  List<dynamic> _accounts = const [];
  String? _accountId;
  late String _type = widget.item['direction'] == 'in' ? 'income' : 'expense';
  @override
  void initState() {
    super.initState();
    final mapping = widget.item['mapping'];
    _accountId = mapping is Map ? mapping['accountId']?.toString() : null;
    _loadAccounts();
  }

  Future<void> _loadAccounts() async {
    try {
      final data = await widget.api.request('GET', '/accounts');
      if (mounted) setState(() => _accounts = (data as List?) ?? const []);
    } catch (_) {
      // The mapped account remains the server-side fallback if this optional
      // picker cannot be loaded (for example, offline while editing).
    }
  }

  @override
  void dispose() {
    _name.dispose();
    _category.dispose();
    _note.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => FinoraSheet(
    title: 'Sửa giao dịch',
    subtitle: 'Dữ liệu gốc từ ngân hàng không thể thay đổi.',
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        DropdownButtonFormField<String>(
          initialValue: _type,
          items: const [
            DropdownMenuItem(value: 'income', child: Text('Thu')),
            DropdownMenuItem(value: 'expense', child: Text('Chi')),
          ],
          onChanged: (value) => setState(() => _type = value!),
          decoration: const InputDecoration(labelText: 'Loại'),
        ),
        TextField(
          controller: _name,
          decoration: const InputDecoration(labelText: 'Tên giao dịch'),
        ),
        TextField(
          controller: _category,
          decoration: const InputDecoration(
            labelText: 'Mã danh mục (không bắt buộc)',
          ),
        ),
        DropdownButtonFormField<String>(
          initialValue:
              _accounts.any((item) => item['id']?.toString() == _accountId)
              ? _accountId
              : null,
          items: _accounts
              .map(
                (raw) => DropdownMenuItem<String>(
                  value: raw['id']?.toString(),
                  child: Text(raw['name']?.toString() ?? 'Tài khoản Finora'),
                ),
              )
              .toList(),
          onChanged: (value) => setState(() => _accountId = value),
          decoration: const InputDecoration(labelText: 'Tài khoản Finora'),
        ),
        TextField(
          controller: _note,
          decoration: const InputDecoration(labelText: 'Ghi chú'),
        ),
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          value: _remember,
          onChanged: (value) => setState(() => _remember = value),
          title: const Text('Ghi nhớ lựa chọn này cho các lần sau'),
        ),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: const Color(0xfff8fafc),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            'Dữ liệu gốc từ ngân hàng\n${widget.item['description'] ?? ''}\n${widget.item['amount'] ?? ''} ${widget.item['currency'] ?? ''}',
            style: const TextStyle(fontSize: 12, color: Color(0xff475569)),
          ),
        ),
        const SizedBox(height: 16),
        FilledButton(
          onPressed: () => Navigator.pop(context, {
            'name': _name.text.trim(),
            'categoryId': _category.text.trim(),
            'accountId': _accountId ?? '',
            'note': _note.text.trim(),
            'type': _type,
            'rememberChoice': _remember,
          }),
          child: const Text('Lưu và xác nhận'),
        ),
      ],
    ),
  );
}

class FinoraSheet extends StatelessWidget {
  const FinoraSheet({
    super.key,
    required this.title,
    required this.subtitle,
    required this.child,
  });
  final String title, subtitle;
  final Widget child;
  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.fromLTRB(
      16,
      0,
      16,
      16 + MediaQuery.of(context).viewInsets.bottom,
    ),
    child: Container(
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 22),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: Colors.white, width: 1.5),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.22),
            blurRadius: 36,
            offset: const Offset(0, -10),
          ),
        ],
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 44,
                height: 5,
                decoration: BoxDecoration(
                  color: const Color(0xffcbd5e1),
                  borderRadius: BorderRadius.circular(9),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Text(
              title,
              style: const TextStyle(
                color: Color(0xff0f172a),
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 5),
            Text(
              subtitle,
              style: const TextStyle(
                color: Color(0xff64748b),
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 20),
            child,
          ],
        ),
      ),
    ),
  );
}

class _AccountFormSheet extends StatefulWidget {
  const _AccountFormSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_AccountFormSheet> createState() => _AccountFormSheetState();
}

class _AccountFormSheetState extends State<_AccountFormSheet> {
  String selectedType = 'cash';
  final nameController = TextEditingController(text: 'Ví Tiền mặt');
  final balanceController = TextEditingController();
  bool submitting = false;

  final accountTypes = const [
    {
      'key': 'cash',
      'title': 'Tiền mặt',
      'icon': Icons.payments_rounded,
      'color': Color(0xff4ade80),
      'defaultName': 'Ví Tiền mặt',
    },
    {
      'key': 'bank',
      'title': 'Ngân hàng',
      'icon': Icons.account_balance_rounded,
      'color': Color(0xff38bdf8),
      'defaultName': 'Tài khoản Ngân hàng',
    },
    {
      'key': 'gold',
      'title': 'Vàng',
      'icon': Icons.monetization_on_rounded,
      'color': Color(0xfffbbf24),
      'defaultName': 'Vàng SJC 9999',
    },
    {
      'key': 'real_estate',
      'title': 'Đất / BĐS',
      'icon': Icons.home_work_rounded,
      'color': Color(0xfff43f5e),
      'defaultName': 'Bất động sản / Đất',
    },
  ];

  Color _accountTypeColor(String key) => switch (key) {
    'cash' => const Color(0xff15803d),
    'bank' => const Color(0xff0369a1),
    'gold' => const Color(0xffa16207),
    _ => const Color(0xffbe123c),
  };

  void _onSelectType(Map<String, dynamic> item) {
    setState(() {
      final oldDefault =
          accountTypes.firstWhere(
                (x) => x['key'] == selectedType,
                orElse: () => accountTypes.first,
              )['defaultName']
              as String;

      selectedType = item['key'] as String;

      if (nameController.text.trim().isEmpty ||
          nameController.text == oldDefault) {
        nameController.text = item['defaultName'] as String;
      }
    });
  }

  Future<void> _submit() async {
    setState(() => submitting = true);
    try {
      final selectedItem = accountTypes.firstWhere(
        (x) => x['key'] == selectedType,
        orElse: () => accountTypes.first,
      );
      final finalName = nameController.text.trim().isNotEmpty
          ? nameController.text.trim()
          : selectedItem['defaultName'] as String;

      final payload = <String, dynamic>{
        'name': finalName,
        'type': selectedType,
        'currency': 'VND',
      };

      if (balanceController.text.trim().isNotEmpty) {
        final balance = parseSmartAmount(balanceController.text);
        if (balance <= 0) {
          showError(context, 'Nhập số dư hợp lệ, ví dụ 30tr.');
          return;
        }
        payload['initialBalance'] = balance.toStringAsFixed(0);
      }

      await widget.api.request('POST', '/accounts', payload);
      if (mounted) widget.onSuccess();
    } catch (e) {
      if (mounted) showError(context, presentableError(e));
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    nameController.dispose();
    balanceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FinoraSheet(
      title: 'Thêm tài khoản',
      subtitle:
          'Chọn loại tài khoản & số dư. Thông tin khác sẽ được tự động nội suy.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'LOẠI TÀI KHOẢN / TÀI SẢN',
            style: TextStyle(
              color: Color(0xff92400e),
              fontSize: 11,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.1,
            ),
          ),
          const SizedBox(height: 12),
          GridView.count(
            crossAxisCount: 2,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            mainAxisSpacing: 10,
            crossAxisSpacing: 10,
            childAspectRatio: 2.3,
            children: accountTypes.map((item) {
              final isSelected = selectedType == item['key'];
              final color = item['color'] as Color;
              final foreground = _accountTypeColor(item['key'] as String);
              return InkWell(
                onTap: () => _onSelectType(item),
                borderRadius: BorderRadius.circular(18),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? color.withValues(alpha: 0.2)
                        : const Color(0xfff1f5f9),
                    borderRadius: BorderRadius.circular(18),
                    border: Border.all(
                      color: isSelected ? foreground : const Color(0xff94a3b8),
                      width: isSelected ? 2.2 : 1.4,
                    ),
                    boxShadow: isSelected
                        ? [
                            BoxShadow(
                              color: foreground.withValues(alpha: 0.18),
                              blurRadius: 10,
                            ),
                          ]
                        : null,
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: color.withValues(alpha: 0.18),
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          item['icon'] as IconData,
                          color: foreground,
                          size: 18,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          item['title'] as String,
                          style: TextStyle(
                            color: const Color(0xff0f172a),
                            fontWeight: isSelected
                                ? FontWeight.bold
                                : FontWeight.w600,
                            fontSize: 13,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (isSelected)
                        Icon(
                          Icons.check_circle_rounded,
                          color: foreground,
                          size: 16,
                        ),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 16),
          _CustomGlassTextField(
            controller: balanceController,
            keyboardType: TextInputType.text,
            labelText: 'Số dư / Giá trị ban đầu (vd: 30tr)',
            icon: Icons.payments_outlined,
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: nameController,
            labelText: 'Tên tài khoản (Tự động gợi ý)',
            icon: Icons.edit_note_rounded,
          ),
          const SizedBox(height: 18),
          _AnimatedGoldButton(
            busy: submitting,
            label: 'Lưu tài khoản',
            onTap: _submit,
          ),
        ],
      ),
    );
  }
}

class FinoraSurface extends StatelessWidget {
  const FinoraSurface({super.key, required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: FinoraColors.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: FinoraColors.border, width: 1),
        boxShadow: FinoraElevation.card,
      ),
      child: child,
    );
  }
}

class _ScreenIntro extends StatelessWidget {
  const _ScreenIntro(this.text);
  final String text;
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        text,
        style: TextStyle(
          color: FinoraColors.textSecondary,
          height: 1.35,
          fontSize: 12.5,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

class SectionTitle extends StatelessWidget {
  const SectionTitle(this.text, {super.key, required this.icon});
  final String text;
  final IconData icon;
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: FinoraColors.primarySoft,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: FinoraColors.primary, size: 16),
          ),
          const SizedBox(width: 8),
          Text(
            text,
            style: TextStyle(
              fontWeight: FontWeight.w800,
              color: FinoraColors.textPrimary,
              fontSize: 14.5,
            ),
          ),
        ],
      ),
    );
  }
}

class FinoraListTile extends StatelessWidget {
  const FinoraListTile({
    super.key,
    required this.icon,
    required this.title,
    required this.subtitle,
    this.amount,
    this.badge,
    this.iconColor,
    this.onDelete,
  });
  final IconData icon;
  final String title, subtitle;
  final String? amount, badge;
  final Color? iconColor;
  final VoidCallback? onDelete;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: FinoraSurface(
        child: Row(
          children: [
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                color: (iconColor ?? FinoraColors.primary).withValues(
                  alpha: 0.12,
                ),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                icon,
                color: iconColor ?? FinoraColors.primary,
                size: 18,
              ),
            ),
            const SizedBox(width: 11),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: FinoraColors.textPrimary,
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w500,
                      color: FinoraColors.textSecondary,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            if (amount != null)
              Text(
                amount!,
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  color: iconColor ?? FinoraColors.primary,
                  fontSize: 13,
                ),
              ),
            if (badge != null)
              Container(
                margin: const EdgeInsets.only(left: 8),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
                decoration: BoxDecoration(
                  color: (iconColor ?? FinoraColors.primary).withValues(
                    alpha: 0.14,
                  ),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  badge!,
                  style: TextStyle(
                    color: iconColor ?? FinoraColors.primary,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            if (onDelete != null)
              Padding(
                padding: const EdgeInsets.only(left: 6),
                child: IconButton(
                  onPressed: onDelete,
                  tooltip: 'Xóa tài khoản',
                  icon: const Icon(
                    Icons.delete_outline_rounded,
                    color: Color(0xfff87171),
                    size: 20,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

IconData _iconForTitle(String title) => switch (title) {
  'Tài khoản' => Icons.account_balance_rounded,
  'Khoản vay' => Icons.request_quote_rounded,
  'Tài sản' => Icons.inventory_2_rounded,
  'Bất động sản' => Icons.home_work_rounded,
  'Danh mục' => Icons.category_rounded,
  'Dự báo' => Icons.auto_graph_rounded,
  'Trợ lý AI' => Icons.smart_toy_rounded,
  _ => Icons.bolt_rounded,
};

class ErrorBox extends StatelessWidget {
  const ErrorBox(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext c) => Semantics(
    liveRegion: true,
    container: true,
    label: presentableError(text),
    child: Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: const Color(0x14f04438),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: FinoraColors.danger.withValues(alpha: 0.35),
          ),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.info_outline_rounded, color: FinoraColors.danger),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                presentableError(text),
                style: const TextStyle(
                  color: FinoraColors.textPrimary,
                  height: 1.35,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class EmptyState extends StatelessWidget {
  const EmptyState(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext c) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 46, horizontal: 24),
    child: Center(
      child: Column(
        children: [
          Container(
            width: 62,
            height: 62,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: FinoraColors.primarySoft,
            ),
            child: const Icon(
              Icons.inbox_rounded,
              color: FinoraColors.primary,
              size: 29,
            ),
          ),
          const SizedBox(height: 14),
          Text(
            text,
            textAlign: TextAlign.center,
            style: FinoraTypography.title,
          ),
          const SizedBox(height: 5),
          Text(
            'Dữ liệu mới sẽ xuất hiện tại đây.',
            textAlign: TextAlign.center,
            style: TextStyle(color: FinoraColors.textSecondary, fontSize: 13),
          ),
        ],
      ),
    ),
  );
}

String presentableError(Object error) {
  final raw = error.toString().trim();
  if (raw.isEmpty) return 'Đã xảy ra lỗi. Vui lòng thử lại.';
  if (raw.contains('SocketException') || raw.contains('Failed host lookup')) {
    return 'Không thể kết nối máy chủ. Kiểm tra mạng rồi thử lại.';
  }
  return raw.replaceFirst(RegExp(r'^(Exception|ApiException):\s*'), '').trim();
}

void showError(BuildContext c, String e) =>
    ScaffoldMessenger.of(c).showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: const Color(0xff200733),
        content: Text(
          presentableError(e),
          style: const TextStyle(color: Colors.white),
        ),
      ),
    );
