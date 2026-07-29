# Finora mobile redesign specification

This specification is implemented without changing a route, endpoint, request payload, or domain rule. It is the handoff source for the Flutter UI and for a Figma file.

## Foundation

| Token | Value / rule |
| --- | --- |
| Canvas | `#FAFAFC`, light mode only |
| Primary action | `#6D5DF6`; pressed state is 4% darker and scales to 96% |
| Supporting purple | `#8B7BFF`; soft selection / icon well `#F3F0FF` |
| Surface / border | `#FFFFFF` / `#ECEAFB` |
| Copy | `#232323` primary, `#707070` secondary |
| Status | success `#36C275`, warning `#F5A623`, error `#F04438` |
| Type | Inter/SF Pro fallback: 32/24/20/18/16/14/12; tabular figures for money |
| Grid | 8 pt; page inset 24; card gap 16; control min-height 48; icon 20–24 |
| Shape | cards and fields 16, sheets 24 top, pills 999 |
| Elevation | `0 5 18 #6D5DF6 5%`; only elevated cards, sheets, and floating action |

Maple is atmospheric only: the supplied maple artwork sits at 4–8% opacity behind the login and workspace canvas. It is never placed beneath copy, fields, money, or charts.

## Screen inventory and Figma-ready handoff

| Screen / state | Structure and components | Interaction, motion, responsive and accessibility |
| --- | --- | --- |
| Splash / sign in / register | Maple backdrop → compact brand rail → single white auth card → primary CTA → capability rail. Fields have labels above the input, visibility control, inline error, language and notification buttons. | Auth card fades/slides 12 px on entry; CTA scales/ripples. Autofocus first field; labels remain visible; 48 px targets; errors are announced. On tablet constrain to 400 px; on small phones content scrolls above the safe area. |
| Home shell / navigation | Frosted white top bar, subtle maple canvas, adaptive navigation. Mobile: five-item bottom nav with raised primary create button. Desktop/tablet: left rail. | Selected destination uses a purple icon and 12 px label. Tap transitions are 180 ms fade/slide. Labels are never icon-only; selected state is semantic. |
| Dashboard / statistics / charts | Page label → 24/32 title → purple net-worth hero → horizontally scrollable metric cards → activity/chart sections. | Balance conceal toggle, refresh, progressive skeletons. Hero is an information block rather than a decorative gradient. Metrics become two columns on larger widths and one compact scroll row on phones. |
| Accounts, assets, real estate, portfolio, bank, budget, forecast, automation, audit | Shared resource frame: eyebrow, title, one clear primary Add action, contextual intro, filter/list rows, empty/error states. Rows use a 36–44 px icon well, title, concise metadata, amount/status badge and swipe-to-delete where supported. | Lists preserve existing GET/POST/DELETE behavior. Add/edit is a spring bottom sheet with drag handle and sticky submit. Loading uses neutral skeletons; retry/error and destructive confirmation are explicit. 24 px inset collapses to 16 only below 360 px. |
| Transactions, transfer and create flows | Transaction list has clear income/expense status, readable signed amount, date and category. Create uses grouped fields, segmented choices and one bottom primary CTA. | Category/type selection has a 2 px purple selection ring and check. Keyboard-safe sheet, text labels plus icon/color redundancy, numeric keypad for amounts, success snackbar. |
| Loans / borrower / collection | Summary hero → prominent “Quy đổi lãi suất” entry → borrower cards → borrower detail → loan detail and collection/create sheets. The converter accepts đồng/đầu triệu/ngày or % per day/month/year, exposes quick 2–4 nghìn chips, equivalent rates, formula, and 50/100/200 million examples. | Cards elevate on press then push to details. The converter is a keyboard-safe bottom sheet and can pass its calculated annual rate plus đồng/triệu/day directly into the existing create-loan fields. Money uses tabular numerals and never relies on color for collection state. Detail uses 24 px page inset, 16 px grouping rhythm, and accessible amount labels. |
| Personal information / profile / settings / permissions | Profile header with initials/avatar, grouped settings cells, account/security group, display preference selection and sign out action. | Each setting has title, supporting copy, trailing state and 48 px hit area. Destructive choices are isolated and confirmed. Light appearance is intentional; no dark-mode control is exposed. |
| Notifications / search / empty / error / loading / success | Notification rows pair a quiet leading icon, timestamp and unread marker. Search uses a persistent field with clear control. Empty/error state contains a pale purple illustration well, explanation, and one recovery CTA. | Unread status uses dot plus text/semantics. Error and success snackbar are floating, high contrast, and announced. Skeletons preserve final layout to prevent shifts. |
| AI assistant / history / knowledge-base mode | Header with New chat, mode pills, conversation history, welcome illustration and suggested prompts; messages are distinct purple user and white assistant cards; composer is fixed above safe area. | Existing `/assistant/commands` GET/POST calls remain unchanged. Send shows an in-context “thinking” indicator; reply actions include copy and feedback. Composer supports multiline and Send action. On narrow screens bubbles reserve 48 px opposite margin; all controls have labels/tooltips. |

## Component specification

`Button / primary`: 48 high, 16 radius, purple fill, white 14/700 label; press 0.96 scale. `Button / secondary`: white with purple border and label. `Card`: white, 16 radius, 1 px `#ECEAFB`, 16 padding, 16 gap. `Field`: pale purple-tinted fill, persistent label, purple 1.5 px focus border. `Bottom sheet`: white, 24 top radius, 40 px handle, 24 horizontal inset. `Dialog`: white, 24 radius, clear cancel and destructive action. `Badge`: 12/700, status color at 12% fill plus solid text. `Toast`: floating rounded 16, concise text, respects safe areas.

## Content and accessibility rules

Keep one primary action per view, shorten metadata before truncating titles, and use descriptions rather than color alone for financial gain/loss, success, warning, or error. Maintain a 4.5:1 text contrast minimum, 48 px touch targets, logical reading order, focus-visible fields, and `Semantics` labels for icon-only controls. All money formats use tabular figures; loading and network feedback are announced without blocking navigation.
