import { NgForOf } from '@angular/common';
import { Component, HostBinding, Input } from '@angular/core';
import { FALLBACK_ICON_PATHS, ICON_PATHS } from './icon.registry';

@Component({
  selector: 'app-icon',
  standalone: true,
  imports: [NgForOf],
  template: `
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      focusable="false"
      aria-hidden="true"
    >
      <path *ngFor="let path of paths" [attr.d]="path" />
    </svg>
  `,
  styles: [`
    :host {
      display: inline-flex;
      width: 1em;
      height: 1em;
      flex: 0 0 auto;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      line-height: 1;
      vertical-align: -0.125em;
    }

    svg {
      display: block;
      width: 100%;
      height: 100%;
      overflow: visible;
    }
  `],
})
export class IconComponent {
  @Input() name = '';
  @Input() label = '';

  @HostBinding('class.ld-icon') readonly iconClass = true;
  @HostBinding('attr.aria-hidden') get ariaHidden(): string | null {
    return this.label ? null : 'true';
  }
  @HostBinding('attr.aria-label') get ariaLabel(): string | null {
    return this.label || null;
  }
  @HostBinding('attr.role') get role(): string | null {
    return this.label ? 'img' : null;
  }

  get paths(): readonly string[] {
    return ICON_PATHS[this.name] ?? FALLBACK_ICON_PATHS;
  }
}
