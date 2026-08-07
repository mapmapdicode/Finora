import { Pipe, PipeTransform } from '@angular/core';

/** Formats monetary values in the full Vietnamese accounting style. */
@Pipe({ name: 'vndMoney', standalone: true })
export class VndMoneyPipe implements PipeTransform {
  transform(value: string | number | null | undefined, currency = 'VND', showPositiveSign = false): string {
    const amount = Number.parseFloat(String(value ?? 0));
    const normalized = Number.isFinite(amount) ? amount : 0;
    const sign = showPositiveSign && normalized > 0 ? '+' : '';
    const unit = (currency || 'VND').trim().toUpperCase() || 'VND';
    const formatted = new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 0 }).format(normalized);

    return `${sign}${formatted} ${unit}`;
  }
}
