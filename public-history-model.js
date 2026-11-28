export function compareHistoryEntries(a, b) {
    return Number(b.committedAt !== undefined) - Number(a.committedAt !== undefined)
        || (b.committedAt ?? b.generatedAt).localeCompare(a.committedAt ?? a.generatedAt)
        || a.sha.localeCompare(b.sha);
}
