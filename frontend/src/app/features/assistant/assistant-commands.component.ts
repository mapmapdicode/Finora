import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { AssistantCommand } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-assistant-commands',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TranslatePipe, IconComponent],
  templateUrl: './assistant-commands.component.html',
})
export class AssistantCommandsComponent implements OnInit {
  commands: AssistantCommand[] = [];
  statusMessage = '';
  form: FormGroup;
  selectedCommandId: string | null = null;
  approvalInputs: Record<string, string> = {};
  loading = false;

  constructor(
    private api: ApiService,
    private fb: FormBuilder,
    public auth: AuthService,
  ) {
    this.form = this.fb.group({
      command: ['', Validators.required],
      plan: [''],
    });
  }

  ngOnInit() {
    this.reload();
  }

  reload() {
    this.api.listAssistantCommands().subscribe({
      next: (items) => {
        this.commands = items;
      },
      error: () => {
        this.statusMessage = 'Không thể tải danh sách assistant commands.';
      },
    });
  }

  submitCommand() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    this.loading = true;
    this.api
      .createAssistantCommand({
        command: this.form.value.command || '',
        plan: this.form.value.plan || undefined,
      })
      .subscribe({
        next: (item) => {
          this.loading = false;
          this.approvalInputs[item.id] = item.approvalId || '';
          this.statusMessage = item.status === 'awaiting_approval'
            ? 'Đã gửi command và nhận token phê duyệt.'
            : 'Đã gửi command cho assistant.';
          this.form.reset({ command: '', plan: '' });
          this.reload();
        },
        error: () => {
          this.loading = false;
          this.statusMessage = 'Gửi command thất bại.';
        },
      });
  }

  refreshOne(id: string) {
    this.api.getAssistantCommand(id).subscribe({
      next: (item) => {
        const idx = this.commands.findIndex((row) => row.id === id);
        if (idx >= 0) {
          this.commands[idx] = item;
          this.commands = [...this.commands];
        }
      },
      error: () => {
        this.statusMessage = 'Không thể làm mới command.';
      },
    });
  }

  approve(id: string) {
    if (!this.auth.canMutate) return;
    const item = this.commands.find((row) => row.id === id);
    const approvalId = item?.approvalId?.trim() || this.approvalInputs[id]?.trim() || '';
    this.api.approveAssistantCommand(id, approvalId || undefined).subscribe({
      next: () => {
        this.statusMessage = 'Đã phê duyệt.';
        this.refreshOne(id);
      },
      error: () => {
        this.statusMessage = 'Phê duyệt command thất bại.';
      },
    });
  }

  cancel(id: string) {
    if (!this.auth.canMutate) return;
    this.api.cancelAssistantCommand(id).subscribe({
      next: () => {
        this.statusMessage = 'Đã huỷ.';
        this.refreshOne(id);
      },
      error: () => {
        this.statusMessage = 'Huỷ command thất bại.';
      },
    });
  }

  canApprove(item: AssistantCommand) {
    return item.status === 'awaiting_approval';
  }

  requiresApproval(item: AssistantCommand) {
    return item.status === 'awaiting_approval' && !item.approvalId && !this.approvalInputs[item.id]?.trim();
  }

  inspect(item: AssistantCommand) {
    this.selectedCommandId = item.id;
  }

  setApprovalInput(id: string, value: string) {
    this.approvalInputs[id] = value;
  }
}
