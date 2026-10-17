export function lastJournalId(userId: string): string | null {
  try { return localStorage.getItem(`last-journal:${userId}`); } catch { return null; }
}

export function rememberJournal(userId: string, journalId: string) {
  try { localStorage.setItem(`last-journal:${userId}`, journalId); } catch { /* Navigation works without browser storage. */ }
}
