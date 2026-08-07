import { FormBuilder } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { LoanListComponent } from './loan-list.component';

describe('LoanListComponent payment modes', () => {
  let component: LoanListComponent;

  beforeEach(() => {
    component = new LoanListComponent(
      jasmine.createSpyObj<ApiService>('ApiService', ['getLoans']),
      new FormBuilder(),
      { canMutate: true } as AuthService,
    );
  });

  it('clears the unused amount when choosing interest-only or principal-only', () => {
    component.paymentForm.patchValue({ interestAmount: '50.000', principalAmount: '200.000' });

    component.setPaymentMode('interest_only');
    expect(component.paymentForm.value.interestAmount).toBe('50.000');
    expect(component.paymentForm.value.principalAmount).toBe('0');

    component.setPaymentMode('principal_only');
    expect(component.paymentForm.value.interestAmount).toBe('0');
    expect(component.paymentForm.value.principalAmount).toBe('0');
  });

  it('fills two separate fields only for a full settlement', () => {
    component.selectedLoanForPayment = 'loan-1';
    component.activePaymentLoan = { id: 'loan-1', principalBalance: '200000' } as any;
    component.accrualsMap['loan-1'] = { totalAccruedInterest: '50000' } as any;

    component.setQuickPaymentAmount('full');

    expect(component.paymentMode).toBe('both');
    expect(component.paymentForm.value.interestAmount).toBe('50.000');
    expect(component.paymentForm.value.principalAmount).toBe('200.000');
  });
});
