# Finora Wealth OS Design System

## Foundation

Finora is a premium personal wealth-management product: calm, intelligent, trustworthy, and modern. Use deep violet for identity and primary action; reserve gold/amber for value, emphasis, and premium highlights. Decorative maple imagery is subordinate to data and only permitted behind a protected readable surface.

### Color roles

| Role | Flutter token | Intent |
|---|---|---|
| Primary deep / primary / soft | `primaryDeep`, `primary`, `primarySoft` | Brand anchors, primary actions, selected navigation, subtle selected surfaces |
| Gold / amber | `accentGold`, `accentAmber` | Wealth emphasis, premium badge, selected highlight; never the only semantic signal |
| Success / warning / danger / info | semantic status tokens | Status paired with text and icon |
| Background / surface / elevated / glass | semantic surface tokens | Light, spacious content hierarchy |
| Text primary / secondary / tertiary | semantic text tokens | Clear hierarchy; labels must never compete with values |
| Border / border strong | semantic border tokens | Surface separation and focus support |

### Spacing, radius, and elevation

Spacing: 4, 8, 12, 16, 20, 24, 32. Radius: 8, 12, 16, 20, 24, full. Use 16–20 for list/card surfaces and 24 only for major hero or sheet surfaces. One quiet card shadow and one floating elevation are permitted; do not stack heavy shadows.

### Typography

| Role | Size | Usage |
|---|---:|---|
| Display | 34 | Net-worth hero only |
| H1 / H2 / H3 | 28 / 22 / 18 | Screen, section, card hierarchy |
| Title / body / body small | 16 / 15 / 13 | Interactive titles and reading content |
| Caption / label | 12 / 12 | Supporting detail and field labels |
| Money | 28 default | Tabular figures, strongest visual element in finance summary |

Use tabular figures for money; always pair abbreviated money with clear currency context. Responsive values may truncate, wrap by token boundary, or reduce only within the defined type scale—never overflow.

## Component contracts

| Component | Contract |
|---|---|
| FinoraScaffold / header | Safe area, restrained maple decoration, brand, contextual actions, one screen title. |
| FinoraCard / section | One surface hierarchy; title, optional action, content. Do not card every line of text. |
| FinoraButton / icon button | 48px primary/secondary height; icon-only controls have at least 44×44 hit targets. |
| FinoraTextField | Persistent label, helper/error text, focus/disabled state; placeholder never replaces label. |
| FinoraMoneyInput | Formats numeric input, conveys currency, preserves editable final amount distinct from a suggested amount. |
| FinoraDatePicker | Human-readable value, native picker, validation and helper support. |
| FinoraMoney | Tabular money display with semantic sign and currency. |
| FinoraStatusBadge | Text + icon + status color; color is never the only indication. |
| FinoraListTile / transaction tile | Minimum 44px touch area, leading semantic icon, concise metadata, financial value aligned to the end. |
| FinoraBottomSheet / dialog | Standard grabber/title/actions; destructive action requires explicit confirmation. |
| FinoraEmptyState / skeleton / error | Explain state, show recovery action where available, never leave a blank data screen. |

## Navigation and feedback

Phone bottom navigation is Home, Accounts, Quick Action, Transactions, Profile. Quick Action opens: create transaction, new loan, collect interest/principal, transfer, and add asset. Loans, assets, investment, planning, automation, and AI remain discoverable from Home contextual sections and profile/tools navigation.

Use a timeline for loan collection history, not a database-style table. Use toast/snackbar for successful, non-blocking confirmation; use dialogs for destructive operations. Maintain keyboard avoidance and screen-reader labels on all controls.

## Chart and data rules

Charts use primary violet as the anchor and a restrained semantic palette; every chart has a textual summary. Income, expense, transfer, loan collection, and disbursement use an icon and label in addition to amount sign/color. Dashboard order: net worth → cash flow → alerts → upcoming money → recent transactions → insight.
