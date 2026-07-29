/** Converts Vietnamese shorthand to a canonical full VND integer for APIs/DB. */
export function normalizeVndAmount(input: unknown): string {
  const raw = String(input ?? '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '')
    .replace(/vnd|đ/g, '');
  const match = raw.match(/^([0-9.,]+)(k|tr|t|m|trieu|triệu|ty|tỷ|b)?$/);
  if (!match) return '';
  const value = parseLocalizedNumber(match[1]);
  if (!Number.isFinite(value) || value < 0) return '';
  const multiplier: Record<string, number> = {
    k: 1_000,
    tr: 1_000_000,
    t: 1_000_000,
    m: 1_000_000,
    trieu: 1_000_000,
    triệu: 1_000_000,
    ty: 1_000_000_000,
    tỷ: 1_000_000_000,
    b: 1_000_000_000,
  };
  return String(Math.round(value * (multiplier[match[2] || ''] || 1)));
}

function parseLocalizedNumber(raw: string): number {
  const dots = (raw.match(/\./g) || []).length;
  const commas = (raw.match(/,/g) || []).length;
  if (dots && commas) {
    const at = Math.max(raw.lastIndexOf('.'), raw.lastIndexOf(','));
    return Number(`${raw.slice(0, at).replace(/[.,]/g, '')}.${raw.slice(at + 1)}`);
  }
  const separator = dots ? '.' : commas ? ',' : '';
  if (!separator) return Number(raw);
  const lastGroup = raw.slice(raw.lastIndexOf(separator) + 1);
  return Number((dots + commas > 1 || lastGroup.length === 3)
    ? raw.replaceAll(separator, '')
    : raw.replace(separator, '.'));
}
