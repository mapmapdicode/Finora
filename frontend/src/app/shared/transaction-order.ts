import { Transaction } from './models';

/** Returns a new list ordered by the transaction time, newest first. */
export function sortTransactionsNewestFirst(transactions: Transaction[]): Transaction[] {
  return [...transactions].sort((left, right) => {
    const timeDifference = new Date(right.occurredAt).getTime() - new Date(left.occurredAt).getTime();
    if (timeDifference !== 0) return timeDifference;
    return right.id.localeCompare(left.id);
  });
}
