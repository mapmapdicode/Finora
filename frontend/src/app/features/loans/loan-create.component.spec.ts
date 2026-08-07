import { Router } from '@angular/router';
import { of } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Customer } from '../../shared/models';
import { LoanCreateComponent } from './loan-create.component';

describe('LoanCreateComponent customer workflow', () => {
  const existingCustomer: Customer = { id: 'customer-1', name: 'Anh Minh', phone: '0900000000' };
  const savedCustomer: Customer = { id: 'customer-2', name: 'Chị Lan', phone: '0911111111' };

  let api: jasmine.SpyObj<ApiService>;
  let toast: jasmine.SpyObj<ToastService>;
  let component: LoanCreateComponent;

  beforeEach(() => {
    api = jasmine.createSpyObj<ApiService>('ApiService', ['getCustomers', 'createCustomer', 'getAccounts', 'createLoan']);
    toast = jasmine.createSpyObj<ToastService>('ToastService', ['show']);
    component = new LoanCreateComponent({} as Router, api, toast);
  });

  it('loads saved customers without preselecting one', () => {
    api.getCustomers.and.returnValue(of([existingCustomer]));

    component.loadCustomers();

    expect(component.customers).toEqual([existingCustomer]);
    expect(component.selectedCustomerId).toBe('');
    expect(component.customersLoading).toBeFalse();
  });

  it('persists a new customer and selects it for the loan', () => {
    component.customers = [existingCustomer];
    component.isAddingNewBorrower = true;
    component.newBorrowerName = savedCustomer.name;
    component.newBorrowerPhone = savedCustomer.phone || '';
    api.createCustomer.and.returnValue(of(savedCustomer));

    component.addQuickBorrower();

    expect(api.createCustomer).toHaveBeenCalledWith({ name: savedCustomer.name, phone: savedCustomer.phone });
    expect(component.customers[0]).toEqual(savedCustomer);
    expect(component.selectedCustomerId).toBe(savedCustomer.id);
    expect(component.isAddingNewBorrower).toBeFalse();
  });

  it('opens the quick-add form from the customer menu option', () => {
    component.onBorrowerSelectionChange('__add_new__');

    expect(component.isAddingNewBorrower).toBeTrue();
    expect(component.newBorrowerName).toBe('');
    expect(component.newBorrowerPhone).toBe('');
  });
});
