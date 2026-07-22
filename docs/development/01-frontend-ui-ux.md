# 01 — Kế hoạch Frontend và UI/UX

## Stack và cấu trúc Angular

- Angular hiện hành, standalone components, strict template/type checking, RxJS cho stream HTTP và signals cho local/UI state.
- UI kit nội bộ theo design token CSS: color, typography, spacing 4 px, radius, elevation, focus ring; không hard-code màu/spacing trong feature component.
- Typed API client sinh từ OpenAPI. DTO không dùng trực tiếp làm view model: mapper chuẩn hóa decimal string, timezone và display status.
- State server dùng service/query cache theo key; state form/dialog/tab dùng component store/signals. Không dùng global store cho draft ngắn hạn.

```text
src/app/
  core/          auth, API client, interceptors, config, error handling
  shared/        money/date/status UI, chart, table, confirm dialog, form controls
  layout/        app shell, top bar, side nav, mobile nav
  features/
    dashboard/ accounts/ cash-flow/ loans/ assets/ forecast/
    bank-feed/ settings/ assistant/
  routes.ts
```

## Information architecture

| Route | Mục tiêu chính | Primary action |
|---|---|---|
| `/overview` | Net worth và thay đổi có giải thích | Cập nhật tài sản / thêm giao dịch |
| `/cash-flow` | Giao dịch, dòng tiền, category | Thêm giao dịch |
| `/accounts` | Số dư và health từng account | Kết nối ngân hàng |
| `/loans` | Dư gốc/lãi/lịch thu | Thêm khoản vay |
| `/assets` | Property và tài sản thủ công | Thêm tài sản / valuation |
| `/forecast` | Kịch bản, cash floor, goals | Tạo kịch bản |
| `/inbox` | Việc đến hạn, bank review, stale valuation | Xử lý item |
| `/settings` | Workspace, consent, integration, audit/export | Quản lý kết nối |

Desktop dùng sidebar có label; mobile dùng bottom navigation: Tổng quan, Dòng tiền, Inbox, Thêm, Khác. `Thêm` mở action sheet gồm thu, chi, chuyển tiền, valuation, loan payment; không nhét tất cả form vào FAB trực tiếp.

## Design system và quy tắc trình bày tài chính

| Token/Component | Quy tắc |
|---|---|
| `Money` | Decimal từ server; `vi-VN`, currency rõ, không làm tròn để submit; hiển thị `1,17 tỷ` kèm số đầy đủ ở tooltip/detail |
| `AsOf` | Mọi card net worth/forecast/balance có “Tính đến …” và timezone khi cần |
| `DataQualityBadge` | `Đã đối soát`, `Cần đối soát`, `Dữ liệu cũ`, `Không có số dư`; không chỉ dùng màu |
| `StatusBadge` | Text + icon + màu; mapping status tập trung, không tự viết trong màn hình |
| `MoneyInput` | Nhập số rõ currency, validate > 0, cho phép paste, không dùng `number` JS để tính |
| `EvidencePanel` | Dùng cho income auto-classify/forecast; cho xem rule, confidence, nguồn event |
| `DestructiveConfirm` | Tên đối tượng + hậu quả + lý do bắt buộc cho void/revoke/write-off |

Màu xanh/đỏ biểu thị hướng tăng/giảm nhưng không mặc định “tốt/xấu”: thu gốc loan tăng cash nhưng không tăng net worth. Luôn hiển thị label attribution.

## Màn hình chi tiết và logic interaction

### 1. Onboarding và workspace rỗng

**Mục tiêu:** đưa user đến dashboard hữu ích trong dưới 3 phút mà không ép tạo loan/property.

1. Chọn base currency → tạo workspace/portfolio mặc định.
2. Chọn “Thêm số dư đầu kỳ”, “Nhập CSV” hoặc “Kết nối ngân hàng”; cho phép bỏ qua.
3. Nếu thêm account: name, type, currency, opening balance/effective date; preview “net worth sẽ tăng … vì cash”.
4. Dashboard rỗng hiển thị checklist có thể dismiss, không chặn thao tác.

Validation inline, giữ form draft local khi network lỗi. Sau success chuyển về account/detail, toast chứa link undo chỉ cho action có compensation an toàn.

### 2. Tổng quan tài sản

**Bố cục desktop:** hero net worth → data quality bar → breakdown + attribution hai cột → 30-day agenda/inbox → trend chart/table. Mobile xếp hero, quality, agenda trước chart.

- Date selector: hôm nay, cuối tháng trước, custom as-of. Khi đổi date, loading chỉ trong vùng query; giữ data cũ có “Đang cập nhật”.
- Hero click mở drill-down sheet: `assets - liabilities = net worth`, mỗi dòng link tới nguồn.
- Breakdown card có absolute value, % và valuation/last update. Không tạo pie chart nếu dưới 2 phần tử.
- Chart có keyboard tooltip và bảng thay thế; nút “xem công thức” mở attribution drawer.
- Nếu snapshot stale: hiển thị `calculatedAt` + nút refresh có rate-limit; không tự spam API.

### 3. Cash flow, accounts và transaction form

**List:** filter account/category/type/status/date, search debounced 300 ms, cursor pagination, total/selected sum. URL giữ filter để share/bookmark nội bộ.

**Form thu/chi:** account, amount, occurred-at, category, note, attachment, recurring toggle. Amount + account là bắt buộc; category bắt buộc với income/expense. Submit disabled khi request đang chạy nhưng có `Idempotency-Key` client-generated cho retry an toàn.

**Transfer:** chọn nguồn/đích, amount, date, note; UI chặn cùng account/currency mismatch trước, server vẫn là nơi quyết định. Preview giải thích “Tổng cash và net worth không đổi”.

**Account detail:** ledger timeline, current/reconciled balance, provider health, last sync, connect/revoke. Số provider unavailable phải hiện “Không hỗ trợ”, không hiện 0.

### 4. Loan UI

Wizard bốn bước: (1) loại và đối tác, (2) gốc/lãi/day-count, (3) lịch/điều khoản, (4) preview + confirm. Mỗi bước lưu draft local; không tạo loan thật tới confirm.

Loan detail gồm header status, principal/outstanding/accrued/received riêng, schedule table, payment timeline, attachment và audit tab.

Payment split dialog lấy amount từ transaction/candidate, cho user phân principal/interest/fee/waived. Client kiểm `sum(parts) = amount`, không cho principal > outstanding; server trả preview chính thức/từ chối. Với SePay candidate, banner nêu source/evidence và không cho bấm “ghi thu nhập”.

### 5. Asset/property và valuation

List có asset class, ownership tag, current valuation, effective date, stale badge. Detail timeline không dùng editable current value; nút “Cập nhật định giá” mở form amount/effective date/source/note/attachment.

Form cho preview attribution: “Không tạo cash; thay đổi net worth dự kiến …”. Effective date tương lai chỉ có ở scenario flow, không ở actual valuation.

### 6. Forecast, goals và inbox

Scenario editor là side panel: horizon, assumptions, event overrides, cash floor. `Run` tạo async job; UI poll theo backoff hoặc nhận SSE sau này. Khi trạng thái running, hiển thị input còn được đọc nhưng không sửa result version.

Result có chart + table + assumption/evidence drawer. Goal progress phải nói rõ scenario được dùng và ngày tính. Inbox item luôn có severity, reason, source, due date và action; resolve action không tự thay sổ cái nếu item chỉ là reminder.

### 7. SePay bank-feed UX

**Connect:** giải thích dữ liệu đọc, bank capability, consent, quyền revoke; redirect Hosted Link/OAuth. Callback hiển thị success/fail và connection health, không lộ token.

**Imported transaction inbox:** tabs `Cần xem`, `Đã tự ghi`, `Đã match`, `Bỏ qua`; row có direction, amount, account, content đã mask, confidence và result. Detail drawer hiển thị raw fields theo quyền, evidence, rule version, ledger link và audit.

**Auto out:** row “Đã tự ghi chi” kèm category; `Sửa` mở reclassify form. **Auto in:** chỉ row confidence ≥ threshold mới là “Đã ghi thu”; row khác ở Cần xem với action Income/Transfer/Loan payment/Asset sale/Ignore/Split.

**Rule builder:** điều kiện dễ hiểu (account, tiền vào/ra, chứa nội dung, amount range, lịch lặp), preview 10 giao dịch gần nhất trước khi enable, priority, scope account/workspace. Không có toggle “Mọi tiền vào là thu nhập”.

## State, request và lỗi

| Tình huống | Hành vi UI |
|---|---|
| 401 | Clear session, quay login, giữ return URL an toàn |
| 403 | Trang/CTA bị ẩn theo permission, nhưng API error hiện “Bạn không có quyền” |
| 409 idempotency/conflict | Tải transaction đã tạo hoặc yêu cầu refresh, không tự submit lại vô hạn |
| 422 validation | Map field error cạnh control; focus lỗi đầu tiên |
| 429 | Hiện retry time; disable CTA tạm thời |
| Offline | Banner; form draft local; không giả vờ đã lưu |
| Async job fail | Result cũ không bị xóa; nêu trace/reference + Retry nếu policy cho phép |

## Accessibility và UI verification

- Contrast WCAG AA; focus visible; không trap focus sai trong modal/drawer.
- Form có label thực, error `aria-describedby`, status live region; chart có data table.
- Test breakpoint 360, 768, 1024, 1440 px; locale Vietnamese và amount lớn/âm/0/unavailable.
- Visual regression cho dashboard, transaction form, loan payment split, bank review và destructive dialogs.
