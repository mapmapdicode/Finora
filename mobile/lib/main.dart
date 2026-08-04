import 'package:flutter/widgets.dart';
import 'package:mobile/app/app.dart';
import 'package:mobile/features/loans/data/services/loan_reminder_service.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await LoanReminderService.instance.initialize();
  runApp(const FinoraApp());
}
