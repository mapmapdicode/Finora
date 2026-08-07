import { CommonModule } from '@angular/common';
import { Component, ElementRef, EventEmitter, HostListener, Input, Output } from '@angular/core';

export interface SelectMenuOption {
  value: string;
  label: string;
  disabled?: boolean;
}

@Component({
  selector: 'app-select-menu',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="select-menu">
      <button
        type="button"
        class="select-menu__trigger"
        [class.is-open]="isOpen"
        [disabled]="disabled"
        [attr.aria-label]="ariaLabel || placeholder"
        aria-haspopup="listbox"
        [attr.aria-expanded]="isOpen"
        (click)="toggle()"
        (keydown)="onTriggerKeydown($event)"
      >
        <span class="select-menu__value" [class.is-placeholder]="!selectedOption">{{ selectedOption?.label || placeholder }}</span>
        <span class="material-symbols-outlined select-menu__chevron" [class.is-open]="isOpen">expand_more</span>
      </button>

      <div
        *ngIf="isOpen"
        class="select-menu__panel"
        [class.select-menu__panel--top]="opensUpward"
        role="listbox"
        [attr.aria-label]="ariaLabel || placeholder"
      >
        <button
          *ngFor="let option of options; let index = index"
          type="button"
          class="select-menu__option"
          [class.is-selected]="option.value === value"
          [class.is-active]="index === activeIndex"
          [disabled]="option.disabled"
          role="option"
          [attr.aria-selected]="option.value === value"
          (click)="select(option)"
          (mouseenter)="activeIndex = index"
        >
          <span>{{ option.label }}</span>
          <span *ngIf="option.value === value" class="material-symbols-outlined select-menu__check">check</span>
        </button>
      </div>
    </div>
  `,
  styles: [`
    :host { display: block; min-width: 0; }
    .select-menu { position: relative; width: 100%; }
    .select-menu__trigger {
      display: flex; width: 100%; min-height: 46px; align-items: center; justify-content: space-between; gap: 12px;
      padding: 0 14px 0 16px; border: 1px solid var(--border-strong); border-radius: 12px;
      background: var(--bg-surface); color: var(--text-main); font: 600 14px var(--w-font-sans); text-align: left;
      box-sizing: border-box; cursor: pointer; transition: border-color .15s ease, box-shadow .15s ease, background-color .15s ease;
    }
    .select-menu__trigger:hover:not(:disabled) { border-color: var(--outline); }
    .select-menu__trigger:focus-visible, .select-menu__trigger.is-open { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(67, 1, 106, .15); outline: 0; }
    .select-menu__trigger:disabled { cursor: not-allowed; color: var(--outline); background: var(--bg-subtle); border-color: var(--border-subtle); }
    .select-menu__value { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .select-menu__value.is-placeholder { color: var(--outline); font-weight: 500; }
    .select-menu__chevron { flex: none; color: var(--accent-primary); font-size: 20px; transition: transform .16s ease; }
    .select-menu__chevron.is-open { transform: rotate(180deg); }
    .select-menu__panel {
      position: absolute; z-index: 60; right: 0; left: 0; max-height: 240px; margin-top: 6px; overflow-y: auto;
      padding: 6px; border: 1px solid rgba(207, 194, 210, .8); border-radius: 12px; background: var(--bg-surface);
      box-shadow: 0 8px 24px rgba(43, 1, 77, .12); animation: select-menu-enter .16s ease-out both;
    }
    .select-menu__panel--top { top: auto; bottom: calc(100% + 6px); margin-top: 0; animation-name: select-menu-enter-top; }
    .select-menu__option {
      display: flex; width: 100%; min-height: 48px; align-items: center; justify-content: space-between; gap: 12px;
      padding: 10px 12px; border: 0; border-radius: 9px; background: transparent; color: var(--text-main);
      font: 600 14px/1.35 var(--w-font-sans); text-align: left; cursor: pointer;
    }
    .select-menu__option:hover:not(:disabled), .select-menu__option.is-active { background: #fae8ff; color: #43016a; }
    .select-menu__option.is-selected { background: #f3daff; color: #43016a; font-weight: 750; }
    .select-menu__option:focus-visible { outline: 3px solid var(--secondary-fixed-dim); outline-offset: -2px; }
    .select-menu__option:disabled { color: var(--outline); cursor: not-allowed; }
    .select-menu__check { flex: none; color: #43016a; font-size: 19px; font-variation-settings: 'FILL' 1; }
    @keyframes select-menu-enter { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }
    @keyframes select-menu-enter-top { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
    @media (prefers-reduced-motion: reduce) { .select-menu__panel { animation: none; } }
  `],
})
export class SelectMenuComponent {
  @Input() options: SelectMenuOption[] = [];
  @Input() value = '';
  @Input() placeholder = 'Chọn một mục';
  @Input() ariaLabel = '';
  @Input() disabled = false;
  @Output() valueChange = new EventEmitter<string>();

  isOpen = false;
  opensUpward = false;
  activeIndex = -1;

  constructor(private host: ElementRef<HTMLElement>) {}

  get selectedOption(): SelectMenuOption | undefined {
    return this.options.find((option) => option.value === this.value);
  }

  toggle() {
    if (this.disabled) return;
    if (this.isOpen) {
      this.close();
      return;
    }
    this.open();
  }

  select(option: SelectMenuOption) {
    if (option.disabled) return;
    this.valueChange.emit(option.value);
    this.close();
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: Event) {
    if (!this.host.nativeElement.contains(event.target as Node)) {
      this.close();
    }
  }

  @HostListener('window:resize')
  onResize() {
    if (this.isOpen) this.updatePlacement();
  }

  onTriggerKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      this.close();
      return;
    }
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      if (!this.isOpen) this.open();
      this.moveActive(event.key === 'ArrowDown' ? 1 : -1);
      return;
    }
    if (event.key === 'Enter' || event.key === ' ') {
      if (!this.isOpen) return;
      event.preventDefault();
      const option = this.options[this.activeIndex];
      if (option) this.select(option);
    }
  }

  private open() {
    this.isOpen = true;
    this.activeIndex = Math.max(0, this.options.findIndex((option) => option.value === this.value));
    queueMicrotask(() => this.updatePlacement());
  }

  private close() {
    this.isOpen = false;
    this.activeIndex = -1;
  }

  private moveActive(step: number) {
    if (!this.options.length) return;
    let index = this.activeIndex;
    for (const _option of this.options) {
      index = (index + step + this.options.length) % this.options.length;
      if (!this.options[index].disabled) {
        this.activeIndex = index;
        return;
      }
    }
  }

  private updatePlacement() {
    const trigger = this.host.nativeElement.querySelector('.select-menu__trigger');
    if (!trigger) return;
    const rect = trigger.getBoundingClientRect();
    const menuHeight = Math.min(240, Math.max(60, this.options.length * 48 + 12));
    this.opensUpward = window.innerHeight - rect.bottom < menuHeight + 8 && rect.top > window.innerHeight - rect.bottom;
  }
}
