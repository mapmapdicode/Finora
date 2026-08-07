import { VndMoneyPipe } from './vnd-money.pipe';

describe('VndMoneyPipe', () => {
  const pipe = new VndMoneyPipe();

  it('formats a VND amount in full with Vietnamese separators', () => {
    expect(pipe.transform('2533804110')).toBe('2.533.804.110 VND');
  });

  it('keeps negative amounts and can mark a positive amount explicitly', () => {
    expect(pipe.transform(-1250000)).toBe('-1.250.000 VND');
    expect(pipe.transform(1250000, 'VND', true)).toBe('+1.250.000 VND');
  });

  it('falls back safely for absent or invalid values', () => {
    expect(pipe.transform(undefined)).toBe('0 VND');
    expect(pipe.transform('invalid')).toBe('0 VND');
  });
});
