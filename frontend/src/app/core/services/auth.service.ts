import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { environment } from '../../../environments/environment';
import { AuthResponse, Role, Workspace } from '../../shared/models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private tokenKey = 'wealthos.token';
  private workspaceKey = 'wealthos.workspace';
  private workspaceRoleKey = 'wealthos.workspaceRole';
  private workspaceRolesKey = 'wealthos.workspaceRoles';

  token$ = new BehaviorSubject<string | null>(this.token);

  constructor(private http: HttpClient) {}

  get token() {
    return localStorage.getItem(this.tokenKey);
  }

  setToken(token: string) {
    localStorage.setItem(this.tokenKey, token);
    this.token$.next(token);
  }

  syncWorkspaceRoles(workspaces: Workspace[]) {
    const roleMap = this.readWorkspaceRoles();
    for (const workspace of workspaces) {
      if (!workspace.id) continue;
      if (workspace.role) {
        roleMap[workspace.id] = workspace.role;
      }
    }
    localStorage.setItem(this.workspaceRolesKey, JSON.stringify(roleMap));
    this.persistSelectedWorkspaceRole();
  }

  clearToken() {
    this.clearSessionMetadata();
    localStorage.removeItem(this.tokenKey);
  }

  login(email: string, password: string) {
    return this.http.post<AuthResponse>(`${environment.apiBase}/api/v1/auth/login`, { email, password });
  }

  register(payload: { email: string; password: string; name: string; workspaceName?: string }) {
    return this.http.post<AuthResponse>(`${environment.apiBase}/api/v1/auth/register`, payload);
  }

  persistSession(response: AuthResponse) {
    if (response.token) {
      this.setToken(response.token);
    }
    if (response.workspace?.id) {
      this.saveWorkspace(response.workspace.id);
    }
  }

  saveWorkspace(workspaceId: string) {
    localStorage.setItem(this.workspaceKey, workspaceId);
    this.persistSelectedWorkspaceRole();
  }

  get workspaceId(): string | null {
    return localStorage.getItem(this.workspaceKey);
  }

  get workspaceRole(): Role | null {
    const selectedRole = localStorage.getItem(this.workspaceRoleKey);
    if (!selectedRole) return null;
    const normalized = selectedRole.trim().toLowerCase();
    if (normalized === 'viewer' || normalized === 'editor' || normalized === 'owner') {
      return normalized as Role;
    }
    return null;
  }

  isViewerRole(): boolean {
    return this.workspaceRole === 'viewer';
  }

  isAuthenticated() {
    return this.token !== null;
  }

  get canMutate() {
    return !this.isViewerRole();
  }

  private readWorkspaceRoles(): Record<string, string> {
    const raw = localStorage.getItem(this.workspaceRolesKey);
    if (!raw) return {};
    try {
      const parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== 'object') {
        return {};
      }
      const output: Record<string, string> = {};
      for (const [key, value] of Object.entries(parsed)) {
        if (typeof key === 'string' && typeof value === 'string' && key.trim() && value.trim()) {
          output[key] = value.trim();
        }
      }
      return output;
    } catch (_error) {
      return {};
    }
  }

  private persistSelectedWorkspaceRole() {
    const roles = this.readWorkspaceRoles();
    const selected = this.workspaceId;
    if (selected && roles[selected]) {
      localStorage.setItem(this.workspaceRoleKey, roles[selected]);
    } else {
      localStorage.removeItem(this.workspaceRoleKey);
    }
  }

  clearSessionMetadata() {
    localStorage.removeItem(this.workspaceRolesKey);
    localStorage.removeItem(this.workspaceRoleKey);
    localStorage.removeItem(this.workspaceKey);
    this.token$.next(null);
  }
}
