import { count, inside, parentPath } from './viewer-dom.js';
import { compareEntries } from './viewer-file-figures.js';
import { pageSize } from './viewer-inspection.js';
import { aggregate } from './viewer-totals.js';
/** The component row for the files no component claims: it cannot be opened as a filter. */
export const notInComponent = 'Not in a component';
export function isGroup(value) {
    return value === 'folders' || value === 'files' || value === 'components';
}
const folderRow = (entry) => ({ entry, kind: 'folder' });
const fileRow = (entry) => ({ entry, kind: 'file' });
/** Components, each totalled over the chosen files it holds, and a last row for the chosen files none holds. */
export function componentRows(components, files, filtering) {
    const chosen = new Map(files.map(entry => [entry.path, entry]));
    const rows = components.map(({ component, paths }) => ({
        entry: aggregate(component.name, paths.flatMap(path => chosen.get(path) ?? [])), kind: 'component',
        files: paths.flatMap(path => chosen.get(path) ?? [])
    }))
        .filter(row => !filtering || row.entry.files > 0);
    const assigned = new Set(components.flatMap(({ paths }) => paths));
    const rest = files.filter(entry => !assigned.has(entry.path));
    if (rest.length)
        rows.push({ entry: aggregate(notInComponent, rest), kind: 'component', files: rest });
    return rows;
}
/** A folder's own listing: the folders right under it with their recorded totals, then its files. */
export function folderRows(directories, files, directory) {
    const under = (entry) => ({ ...folderRow(entry), files: files.filter(file => inside(file.path, entry.path)) });
    return [...directories.filter(entry => entry.path && parentPath(entry.path) === directory).map(under),
        ...files.filter(entry => parentPath(entry.path) === directory).map(fileRow)];
}
/** Filtered folders hold only what passes the query, and their totals count only that. */
export function filteredFolderRows(files, directory) {
    const below = new Map();
    const here = [];
    const prefix = directory ? `${directory}/` : '';
    for (const entry of files) {
        const rest = entry.path.slice(prefix.length);
        const slash = rest.indexOf('/');
        if (slash < 0)
            here.push(entry);
        else {
            const folder = prefix + rest.slice(0, slash);
            below.set(folder, [...below.get(folder) ?? [], entry]);
        }
    }
    return [...[...below].map(([path, entries]) => ({ ...folderRow(aggregate(path, entries)), files: entries })), ...here.map(fileRow)];
}
/** Sorts in place by a column; sorted by path, folders come before files. */
export function sortRows(rows, byPath, value, ascending) {
    return rows.sort((a, b) => {
        const files = Number(a.kind === 'file') - Number(b.kind === 'file');
        return (byPath ? files : 0) || compareEntries(a.entry, b.entry, value, ascending);
    });
}
/**
 * The code comes first, whatever the sort: files no plugin read, and folders holding only such files, fold into one
 * row at the end, open or closed. A query that asks for them (`keep`), or a list of nothing else, shows them as they are.
 */
export function foldUnread(rows, unreadPaths, keep, expanded) {
    const unread = (row) => row.kind === 'file' ? row.entry.analysed === false : row.kind === 'folder' && unreadPaths.has(row.entry.path);
    const folded = keep ? [] : rows.filter(unread);
    const read = rows.filter(row => !folded.includes(row));
    if (!folded.length || !read.length)
        return { display: rows, folded };
    return { display: [...read, { kind: 'unread' }, ...expanded ? folded : []], folded };
}
/**
 * The entries a page shows, 1-based, counting rows and not the fold of ignored files, which takes a place in the list
 * but is no entry of its own. None where the page holds only the fold.
 */
export function entryRange(display, first, shown) {
    const entries = (from, to) => display.slice(from, to).filter(row => row.kind !== 'unread').length;
    const before = entries(0, first);
    const here = entries(first, first + shown);
    return here ? [before + 1, before + here] : null;
}
const nouns = { components: ['component', 'components'], folders: ['entry', 'entries'],
    files: ['file', 'files'] };
/**
 * The line under the table. A filtered file list says how many of the folder's files it kept; other lists name the
 * selection only once. An empty list that a hint explains says nothing more.
 */
export function listSummary(group, counts, filtering, hinted) {
    if (counts.rows === 0)
        return hinted ? '' : 'No files match this selection.';
    const [one, many] = nouns[group];
    const paged = counts.display > pageSize;
    const listsFiles = group === 'files' && counts.rows === counts.files;
    const of = filtering && (paged || !listsFiles) ? ` · ${count(counts.files)} of ${count(counts.scope)} files selected` : '';
    const total = `${count(counts.rows)} ${many}${of}`;
    if (paged)
        return counts.range ? `${count(counts.range[0])}–${count(counts.range[1])} of ${total}` : total;
    if (filtering && listsFiles)
        return `${count(counts.rows)} of ${count(counts.scope)} files`;
    return `${count(counts.rows)} ${counts.rows === 1 ? one : many}${of}`;
}
