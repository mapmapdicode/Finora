import { Transaction } from './models';
import { sortTransactionsNewestFirst } from './transaction-order';

describe('sortTransactionsNewestFirst', () => {
  it('puts newer transactions first without changing the source list', () => {
    const transactions = [
      { id: 'old', occurredAt: '2026-08-05T08:00:00.000Z' },
      { id: 'new', occurredAt: '2026-08-07T08:00:00.000Z' },
      { id: 'middle', occurredAt: '2026-08-06T08:00:00.000Z' },
    ] as Transaction[];

    const result = sortTransactionsNewestFirst(transactions);

    expect(result.map((transaction) => transaction.id)).toEqual(['new', 'middle', 'old']);
    expect(transactions.map((transaction) => transaction.id)).toEqual(['old', 'new', 'middle']);
  });
});
