import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-forbidden',
  standalone: true,
  imports: [RouterLink, IconComponent],
  templateUrl: './forbidden.component.html',
})
export class ForbiddenComponent {}
