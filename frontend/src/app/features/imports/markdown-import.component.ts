import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';

@Component({
  selector: 'app-markdown-import', standalone: true, imports: [CommonModule, FormsModule],
  templateUrl: './markdown-import.component.html',
})
export class MarkdownImportComponent {
  markdown = '';
  month = new Date().toISOString().slice(0, 7);
  overwrite = false;
  loading = false;
  committing = false;
  error = '';
  success = '';
  preview: { canCommit: boolean; summary: { accounts: number; transactions: number; loans: number; payments: number }; issues: Array<{ line: number; section: string; message: string }> } | null = null;

  constructor(private api: ApiService) {}

  readFile(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith('.md')) { this.error = 'Chỉ hỗ trợ file .md.'; return; }
    const reader = new FileReader();
    reader.onload = () => { this.markdown = String(reader.result || ''); this.preview = null; this.error = ''; };
    reader.readAsText(file);
  }

  previewImport() {
    this.error = '';
    this.success = '';
    this.preview = null;
    if (!this.markdown.trim()) { this.error = 'Hãy chọn hoặc dán nội dung file Markdown.'; return; }
    this.loading = true;
    this.api.previewMarkdownImport({ markdown: this.markdown, month: this.month, overwrite: this.overwrite }).subscribe({
      next: (preview) => { this.preview = preview; this.loading = false; },
      error: (err) => { this.error = err?.error?.message || 'Không thể đọc file import.'; this.loading = false; },
    });
  }

  commitImport() {
    if (!this.preview?.canCommit || this.committing) return;
    this.error = '';
    this.success = '';
    this.committing = true;
    this.api.commitMarkdownImport({ markdown: this.markdown, month: this.month, overwrite: this.overwrite }).subscribe({
      next: ({ result }) => {
        this.committing = false;
        this.success = `Đã import: ${result.accountsCreated} tài khoản, ${result.transactionsCreated} thu/chi, ${result.loansCreated} khoản vay, ${result.paymentsCreated} thanh toán.${result.rowsSkipped ? ` Bỏ qua ${result.rowsSkipped} dòng đã có.` : ''}`;
      },
      error: (err) => { this.error = err?.error?.message || 'Không thể ghi dữ liệu import.'; this.committing = false; },
    });
  }
}
