import { element } from './viewer-dom.js';
const shown = 50;
const day = new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short', year: 'numeric' });
export const dayTime = new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
export function commitIndex(entries) {
    return entries.map(entry => {
        const time = when(entry);
        const tags = (entry.tags ?? []).map(tag => tag.toLowerCase());
        const text = [entry.subject ?? '', ...tags, entry.author?.name ?? '', entry.author?.email ?? '',
            time === null ? '' : new Date(time).toISOString().slice(0, 10), time === null ? '' : day.format(new Date(time))]
            .join('\n').toLowerCase();
        return { entry, sha: entry.sha.toLowerCase(), tags, text };
    });
}
/**
 * Commits matching every word of the query, in history order, with exact tags and hash prefixes first. A hash
 * matches from its start only, so "a1" does not match every commit that merely contains it.
 */
export function searchCommits(index, query, limit = shown) {
    const words = query.trim().toLowerCase().split(/\s+/).filter(Boolean);
    const exact = [];
    const rest = [];
    let total = 0;
    for (const [position, item] of index.entries()) {
        if (!words.every(word => item.sha.startsWith(word) || item.text.includes(word)))
            continue;
        total += 1;
        const strong = words.length > 0 && words.some(word => item.tags.includes(word) || (word.length >= 4 && item.sha.startsWith(word)));
        const bucket = strong ? exact : rest;
        if (exact.length + rest.length < limit || strong)
            bucket.push({ entry: item.entry, index: position });
    }
    return { matches: [...exact, ...rest].slice(0, limit), total };
}
/**
 * How a commit's author reads in the list: the name, with the email only when the search found the email and not the
 * name. A history without emails carries the name alone, and then it is the name alone; nothing reads as missing.
 */
export function authorLabel(author, words) {
    if (!author?.name)
        return null;
    return author.email && hit(author.email, words) && !hit(author.name, words) ? `${author.name} <${author.email}>` : author.name;
}
export function hit(text, words) {
    const lower = text.toLowerCase();
    return words.some(word => lower.includes(word));
}
export function when(entry) {
    const time = Date.parse(entry.committedAt ?? entry.generatedAt);
    return Number.isFinite(time) ? time : null;
}
/** Where the typed words stand in the text, in order; a hash is matched only from its start. */
function wordRanges(lower, words, prefixOnly) {
    const ranges = [];
    for (const word of words) {
        if (!prefixOnly) {
            for (let at = lower.indexOf(word); at >= 0; at = lower.indexOf(word, at + word.length))
                ranges.push([at, at + word.length]);
        }
        else if (lower.startsWith(word))
            ranges.push([0, word.length]);
    }
    return ranges.sort((a, b) => a[0] - b[0]);
}
/** Text with the typed words marked; a hash is marked only from its start, as it is matched. */
export function marked(tag, className, text, words, prefixOnly = false) {
    const node = element(tag, className);
    let cursor = 0;
    for (const [start, end] of wordRanges(text.toLowerCase(), words, prefixOnly)) {
        if (end <= cursor)
            continue;
        const from = Math.max(start, cursor);
        if (from > cursor)
            node.append(text.slice(cursor, from));
        node.append(element('mark', '', text.slice(from, end)));
        cursor = end;
    }
    if (cursor < text.length)
        node.append(text.slice(cursor));
    return node;
}
