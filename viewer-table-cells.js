import { basename, count, element, link, parentPath } from './viewer-dom.js';
import { duplicationContext } from './viewer-file-figures.js';
import { isLockfile, languageColor, languageOf } from './viewer-languages.js';
import { partialNote } from './viewer-totals.js';
const icons = { file: 'fx-file-icon', folder: 'fx-folder-icon', component: 'fx-component-icon' };
/** In the flat list the folder is context and the file name is the subject; a long folder wraps after a slash. */
function flatLabel(path) {
    const slash = path.lastIndexOf('/');
    const label = element('span', 'fx-entry-label');
    if (slash >= 0) {
        const folder = element('span', 'fx-entry-folder');
        for (const part of path.slice(0, slash + 1).split(/(?<=\/)/))
            folder.append(part, element('wbr'));
        label.append(folder);
    }
    label.append(element('span', 'fx-entry-name', path.slice(slash + 1)));
    return label;
}
/** A component opens as a filter where it can; the files no component holds are only a label. */
function componentAnchor(path, open) {
    if (!open)
        return element('span', 'fx-entry fx-entry-muted', path);
    return Object.assign(element('button', 'fx-entry fx-entry-button', path), { type: 'button', onclick: open });
}
function entryAnchor(row, context, flat, open) {
    const { entry, kind } = row;
    if (kind === 'component')
        return componentAnchor(entry.path, open);
    const target = kind === 'folder' ? { directory: entry.path } : { directory: parentPath(entry.path), file: entry.path };
    const anchor = link(context.linkTo(target), flat ? '' : basename(entry.path), 'fx-entry');
    if (flat)
        anchor.append(flatLabel(entry.path));
    return anchor;
}
function entryIcon(row) {
    const icon = element('span', icons[row.kind]);
    icon.setAttribute('aria-hidden', 'true');
    if (row.kind === 'file')
        icon.style.color = isLockfile(row.entry.path) ? 'var(--lockfile)' : languageColor(languageOf(row.entry));
    return icon;
}
function cell(entry, column, files) {
    const result = element('td', 'numeric', column.display(entry));
    if (column.value(entry) === 0)
        result.classList.add('fx-zero');
    if (column.label === 'Duplication')
        result.title = duplicationContext(entry) ?? '';
    const partial = files && column.present ? partialNote(files, column.present) : undefined;
    if (partial) {
        result.classList.add('fx-partial');
        result.title = partial;
    }
    return result;
}
/** One row of the Metrics table: its name, with what it is and how many files it holds, then its figures. */
export function metricsRow(row, context, open) {
    const { entry, kind } = row;
    const result = element('tr');
    const name = element('th', 'fx-path-column');
    name.scope = 'row';
    const anchor = entryAnchor(row, context, kind === 'file' && context.group !== 'folders', open);
    anchor.prepend(entryIcon(row));
    if (kind !== 'component')
        anchor.setAttribute('title', entry.path);
    if (entry.path === context.selected) {
        anchor.setAttribute('aria-current', 'true');
        result.dataset.selected = '';
    }
    name.append(anchor);
    if (kind === 'file' && isLockfile(entry.path))
        name.append(element('span', 'fx-entry-tag', 'Lockfile'));
    if (kind !== 'file' && 'files' in entry)
        name.append(element('span', 'fx-entry-count', count(entry.files)));
    result.append(name, ...context.columns.map(column => cell(entry, column, row.files)));
    return result;
}
function counted(total, one, many) {
    return total ? `${count(total)} ${total === 1 ? one : many}` : '';
}
/** Names a list as people say it: "a, b and c". */
export function spokenList(parts) {
    return [parts.slice(0, -1).join(', '), parts.at(-1)].filter(Boolean).join(' and ');
}
/** The fold at the end of the list: how many entries no plugin read, opening and closing in place. */
export function unreadRow(entries, span, expanded, toggle) {
    const lockfiles = entries.filter(row => row.kind === 'file' && isLockfile(row.entry.path)).length;
    const files = entries.filter(row => row.kind === 'file').length - lockfiles;
    const folders = entries.length - files - lockfiles;
    const parts = [counted(lockfiles, 'lockfile', 'lockfiles'), counted(files, 'file', 'files'), counted(folders, 'folder', 'folders')]
        .filter(Boolean);
    const row = element('tr', 'fx-unread');
    const holder = element('td');
    holder.colSpan = span;
    const button = element('button', 'fx-unread-toggle', `${spokenList(parts)} ignored`);
    button.type = 'button';
    button.setAttribute('aria-expanded', String(expanded));
    button.addEventListener('click', toggle);
    holder.append(button);
    row.append(holder);
    return row;
}
/** The table with a sort button per column; a click on the sorted column reverses it. */
export function sortableTable(columns, settings, sorted) {
    const table = element('table', 'fx-table');
    const header = element('tr');
    const thead = element('thead');
    const body = element('tbody');
    const headers = columns.map(({ label }, index) => {
        const th = element('th', index ? 'numeric' : '');
        th.scope = 'col';
        const sort = element('button', 'fx-sort', label);
        sort.type = 'button';
        sort.addEventListener('click', () => {
            settings.ascending = settings.sort === index ? !settings.ascending : index === 0 || (columns[index].ascending ?? false);
            settings.sort = index;
            sorted();
        });
        th.append(sort);
        header.append(th);
        return th;
    });
    thead.append(header);
    table.append(thead, body);
    return { table, headers, body };
}
/** Shows the chosen columns' headers and marks the one the table is sorted by. */
export function markSort(headers, isVisible, settings) {
    headers.forEach((th, index) => {
        th.hidden = !isVisible(index);
        if (index === settings.sort)
            th.setAttribute('aria-sort', settings.ascending ? 'ascending' : 'descending');
        else
            th.removeAttribute('aria-sort');
    });
}
