import 'dart:convert';

import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_timezone/flutter_timezone.dart';
import 'package:mobile/features/loans/domain/entities/loan_reminder.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:timezone/data/latest.dart' as tz;
import 'package:timezone/timezone.dart' as tz;

/// Stores one reminder per loan on the device and schedules its local alert.
/// The reminder is intentionally device-local: it works before server-side push
/// notifications are available and keeps personal reminder preferences private.
class LoanReminderService {
  LoanReminderService._();

  static final instance = LoanReminderService._();
  static const _storageKey = 'loan_reminders_v1';
  static const _channelId = 'loan_due_reminders';
  static const _channelName = 'Nhắc lịch khoản vay';

  final FlutterLocalNotificationsPlugin _notifications =
      FlutterLocalNotificationsPlugin();
  bool _initialized = false;

  Future<void> initialize() async {
    if (_initialized) return;
    tz.initializeTimeZones();
    try {
      final deviceTimeZone = await FlutterTimezone.getLocalTimezone();
      tz.setLocalLocation(tz.getLocation(deviceTimeZone));
    } catch (_) {
      // The timezone package falls back safely when a platform has no timezone
      // identifier available (for example, some test environments).
    }
    await _notifications.initialize(
      const InitializationSettings(
        android: AndroidInitializationSettings('@mipmap/ic_launcher'),
        iOS: DarwinInitializationSettings(
          requestAlertPermission: false,
          requestBadgePermission: false,
          requestSoundPermission: false,
        ),
      ),
    );
    _initialized = true;
  }

  Future<bool> requestPermission() async {
    await initialize();
    final android = _notifications
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >();
    final androidGranted = await android?.requestNotificationsPermission();
    if (androidGranted != false &&
        await android?.canScheduleExactNotifications() == false) {
      await android?.requestExactAlarmsPermission();
    }
    final ios = _notifications
        .resolvePlatformSpecificImplementation<
          IOSFlutterLocalNotificationsPlugin
        >();
    final iosGranted = await ios?.requestPermissions(
      alert: true,
      badge: true,
      sound: true,
    );
    return androidGranted ?? iosGranted ?? true;
  }

  /// Sends an immediate alert so a person can verify device-level settings
  /// before relying on a future collection reminder.
  Future<void> showTestNotification(String borrower) async {
    await initialize();
    await _notifications.show(
      0x3ffffffe,
      'Finora đang hoạt động',
      'Bạn sẽ nhận được nhắc lịch thu cho $borrower tại thời điểm đã chọn.',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          _channelId,
          _channelName,
          channelDescription: 'Nhắc lịch thu cho các khoản vay',
          importance: Importance.high,
          priority: Priority.high,
        ),
        iOS: DarwinNotificationDetails(),
      ),
    );
  }

  Future<LoanReminder?> read(String loanId) async {
    final values = await _readAll();
    return values[loanId];
  }

  Future<void> save(LoanReminder reminder) async {
    await initialize();
    final reminders = await _readAll();
    reminders[reminder.loanId] = reminder;
    await _writeAll(reminders);
    await _notifications.cancel(_notificationId(reminder.loanId));
    if (!reminder.enabled || !reminder.scheduledAt.isAfter(DateTime.now())) {
      return;
    }
    final android = _notifications
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >();
    final canScheduleExactly =
        await android?.canScheduleExactNotifications() ?? false;
    await _notifications.zonedSchedule(
      _notificationId(reminder.loanId),
      'Nhắc thu khoản vay',
      'Đến lịch kiểm tra khoản vay của ${reminder.borrower}.',
      tz.TZDateTime.from(reminder.scheduledAt, tz.local),
      const NotificationDetails(
        android: AndroidNotificationDetails(
          _channelId,
          _channelName,
          channelDescription: 'Nhắc lịch thu cho các khoản vay',
          importance: Importance.high,
          priority: Priority.high,
        ),
        iOS: DarwinNotificationDetails(),
      ),
      // Android can defer inexact alarms by several minutes in Doze mode.
      // Exact mode is used after the user grants the special alarm permission.
      androidScheduleMode: canScheduleExactly
          ? AndroidScheduleMode.exactAllowWhileIdle
          : AndroidScheduleMode.inexactAllowWhileIdle,
    );
  }

  Future<void> remove(String loanId) async {
    final reminders = await _readAll();
    reminders.remove(loanId);
    await _writeAll(reminders);
    await _notifications.cancel(_notificationId(loanId));
  }

  Future<Map<String, LoanReminder>> _readAll() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_storageKey);
    if (raw == null) return {};
    try {
      final values = Map<String, dynamic>.from(jsonDecode(raw) as Map);
      return values.map(
        (id, value) => MapEntry(
          id,
          LoanReminder.fromJson(Map<String, dynamic>.from(value as Map)),
        ),
      );
    } catch (_) {
      return {};
    }
  }

  Future<void> _writeAll(Map<String, LoanReminder> reminders) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
      _storageKey,
      jsonEncode(
        reminders.map((id, reminder) => MapEntry(id, reminder.toJson())),
      ),
    );
  }

  int _notificationId(String loanId) {
    var hash = 0;
    for (final unit in loanId.codeUnits) {
      hash = (hash * 31 + unit) & 0x3fffffff;
    }
    return hash;
  }
}
