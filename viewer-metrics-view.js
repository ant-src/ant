import { columnMenu, tableColumns } from './viewer-columns.js';
import { navigationButtons, pressSegment, segmented } from './viewer-controls.js';
import { count, element, inside, scrollableTable } from './viewer-dom.js';
import { directorySummary, fileDetails } from './viewer-file-aside.js';
import { formatQuery, matches, parseQuery } from './viewer-filter.js';
import { pageSize } from './viewer-inspection.js';
import { renderLanguages } from './viewer-language-bar.js';
import { breadcrumbs, splitter } from './viewer-layout.js';
import { findingIndex } from './viewer-locations.js';
import { labelOf } from './viewer-metrics.js';
import { queryField } from './viewer-query-field.js';
import { suggest } from './viewer-suggest.js';
import { markSort, metricsRow, sortableTable, unreadRow } from './viewer-table-cells.js';
import { componentRows, entryRange, filteredFolderRows, folderRows, foldUnread, isGroup, listSummary, notInComponent, sortRows } from './viewer-table-rows.js';
import { aggregate, componentIndex, unreadFolders } from './viewer-totals.js';
const groupChoices = [['folders', 'Folders'], ['files', 'Files'], ['components', 'Components']];
function tableNote(columns) {
    const effort = labelOf('effort');
    return 'A dash marks an unavailable measurement. Totals include measured files.'
        + (columns.some(column => column.label === effort) ? ` ${effort} is assigned to each finding’s primary file.` : '');
}
function componentHint(hint) {
    hint.hidden = false;
    hint.replaceChildren('No components are defined for this repository. Group paths into components in ', element('code', '', '.repometry/components.json'), ' to compare them here.');
}
/** The Metrics page: a query, a grouping and columns over the folder's files, their languages, and the totals beside. */
export function metricsView(details, navigation, settings, directory, selectedFile) {
    const view = element('div', 'fx-view');
    const { metrics } = details;
    const columns = tableColumns();
    const scopeFiles = metrics.files.filter(entry => inside(entry.path, directory));
    const components = details.components ?? [];
    const facts = { components: componentIndex(components), findings: findingIndex(details.findings ?? []) };
    const state = {
        clauses: parseQuery(navigation.query ?? ''), text: '', page: 0,
        group: isGroup(navigation.group) ? navigation.group : settings.mode === 'tree' ? 'folders' : 'files',
    };
    // Every link out of the view carries the query and the grouping, so a file opened from a selection keeps it.
    const linkTo = (options = {}) => navigation.href('metrics', { ...options, query: formatQuery(state.clauses), group: state.group });
    const trail = element('div');
    view.append(trail);
    const split = element('div', 'fx-split');
    const main = element('div', 'fx-main');
    split.append(main, splitter(split));
    view.append(split);
    const restart = () => {
        state.page = 0;
        remember();
        refresh();
    };
    const field = queryField(state.clauses, {
        suggestions: typed => suggest(typed, scopeFiles, facts, state.clauses),
        change: (next, typed) => {
            state.clauses = next;
            state.text = typed;
            restart();
        },
    });
    const grouping = segmented('Group by', groupChoices, state.group, value => {
        state.group = value;
        settings.mode = state.group === 'folders' ? 'tree' : 'list';
        restart();
    });
    const resort = () => {
        state.page = 0;
        refresh();
    };
    const { menu, isVisible } = columnMenu(columns, settings, { totals: aggregate(directory, scopeFiles), files: scopeFiles }, resort);
    const tools = element('div', 'fx-tools fx-metrics-tools');
    tools.append(field.element, grouping, menu);
    const languages = element('div', 'fx-languages');
    const hint = element('p', 'fx-hint');
    hint.hidden = true;
    const result = element('p', 'fx-note fx-results');
    result.setAttribute('aria-live', 'polite');
    const table = sortableTable(columns, settings, resort);
    const pages = navigationButtons('File pages', () => { state.page--; refresh(); }, () => { state.page++; refresh(); });
    // The count of what the list holds stands with the page buttons, under the table it counts.
    const pager = element('div', 'fx-pager');
    pager.append(result, pages.section);
    const tableWrap = scrollableTable(table.table, 'File and directory metrics');
    const note = element('p', 'fx-note');
    main.append(tools, languages, hint, tableWrap, pager, note);
    let aside = element('aside');
    split.append(aside);
    /** The URL follows the query, without a new history entry per keystroke. */
    function remember() {
        history.replaceState(history.state, '', navigation.href('metrics', {
            directory, ...(selectedFile ? { file: selectedFile.path } : {}), query: formatQuery(state.clauses), group: state.group
        }));
    }
    function selected(...except) {
        const kept = state.clauses.filter(clause => !except.includes(clause.kind));
        return scopeFiles.filter(entry => matches(entry, kept, facts, state.text));
    }
    function groupedRows(files, filtering) {
        hint.hidden = true;
        if (state.group === 'files')
            return files.map(entry => ({ entry, kind: 'file' }));
        if (state.group === 'folders')
            return filtering ? filteredFolderRows(files, directory) : folderRows(metrics.directories, files, directory);
        if (components.length)
            return componentRows(components, files, filtering);
        componentHint(hint);
        return [];
    }
    /** A component's name opens its files as a list, filtered by it. */
    function opener(row) {
        if (row.kind !== 'component' || row.entry.path === notInComponent)
            return undefined;
        return () => {
            state.group = 'files';
            settings.mode = 'list';
            pressSegment(grouping, 'Files');
            field.set([...state.clauses.filter(clause => clause.kind !== 'component'), { kind: 'component', value: row.entry.path }]);
        };
    }
    function toggleFold() {
        settings.unanalysed = !settings.unanalysed;
        refresh();
    }
    /** Only the side panel is swapped: moving the list column would take focus out of the query field. */
    function showAside(files, filtering) {
        const next = selectedFile
            ? fileDetails(details, selectedFile, { ...navigation, href: (target, options) => target === 'metrics' ? linkTo(options) : navigation.href(target, options) })
            : directorySummary(aggregate(directory, files), files, filtering ? `${count(files.length)} of ${count(scopeFiles.length)} files selected` : undefined);
        aside.replaceWith(next);
        aside = next;
    }
    function refresh() {
        const files = selected();
        const filtering = state.clauses.length > 0 || state.text !== '';
        trail.replaceChildren(breadcrumbs('metrics', { href: (_, options) => linkTo(options) }, directory));
        renderLanguages(languages, selected('lang', 'analysed'), state.clauses, next => field.set(next));
        const sort = columns[settings.sort] ?? columns[0];
        const rows = sortRows(groupedRows(files, filtering), settings.sort === 0, sort.value, settings.ascending);
        markSort(table.headers, isVisible, settings);
        table.headers[0].querySelector('button').textContent = state.group === 'components' ? 'Component' : 'Path';
        const keep = state.clauses.some(clause => clause.kind === 'analysed' || clause.kind === 'lockfile');
        const { display, folded } = foldUnread(rows, unreadFolders(scopeFiles), keep, settings.unanalysed);
        const first = state.page * pageSize;
        const visible = display.slice(first, first + pageSize);
        const visibleColumns = columns.slice(1).filter((_, index) => isVisible(index + 1));
        const context = { group: state.group, columns: visibleColumns, linkTo, selected: selectedFile?.path };
        table.body.replaceChildren(...visible.map(row => row.kind === 'unread'
            ? unreadRow(folded, visibleColumns.length + 1, settings.unanalysed, toggleFold) : metricsRow(row, context, opener(row))));
        result.textContent = listSummary(state.group, { rows: rows.length, display: display.length, files: files.length, scope: scopeFiles.length,
            range: entryRange(display, first, visible.length) }, filtering, !hint.hidden);
        tableWrap.hidden = rows.length === 0;
        pager.hidden = !result.textContent;
        pages.section.hidden = display.length <= pageSize;
        pages.update(state.page, Math.ceil(display.length / pageSize));
        note.hidden = rows.length === 0;
        note.textContent = tableNote(visibleColumns);
        showAside(files, filtering);
    }
    refresh();
    return view;
}
