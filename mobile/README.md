# Finora Mobile

Ứng dụng Flutter iOS/Android dùng chung backend với `frontend/`.

```bash
cd mobile
flutter run
```

Mặc định app chọn API theo nền tảng:

- Android emulator: `http://10.0.2.2:8080`
- iOS simulator/macOS: `http://127.0.0.1:8080`

Với thiết bị thật hoặc khi backend chạy máy khác, chỉ định địa chỉ máy chạy backend:

```bash
flutter run --dart-define=API_BASE=http://192.168.1.10:8080
```

Ứng dụng bao gồm: xác thực, workspace, dashboard tài sản ròng, tài khoản, giao dịch, khoản vay, tài sản, bất động sản, ngân sách, dự báo, danh mục, SePay/ngân hàng, tự động hóa, trợ lý và audit log.

## Kiến trúc

Mã nguồn dùng cấu trúc feature-first, với composition root ở `lib/app/` và
hạ tầng dùng chung ở `lib/core/`. Hướng dẫn quy ước và lộ trình tách các màn
hình hiện hữu nằm tại [doc/architecture.md](doc/architecture.md).

```bash
dart format lib test
flutter analyze
flutter test
```
