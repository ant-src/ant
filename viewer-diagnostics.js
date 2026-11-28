import { choice, choices, navigationButtons, searchField } from './viewer-controls.js';
import { contains, element, link, quantity } from './viewer-dom.js';
import { fileHref } from './viewer-file-figures.js';
import { pageSize } from './viewer-inspection.js';
const diagnosticStates = { failed: 'Failed', blocked: 'Blocked', partial: 'Partial' };
function diagnosticItem(row, navigation, files) {
    const item = element('li');
    const heading = element('div', 'fx-event-heading');
    const state = element('span', 'fx-state', diagnosticStates[row.state]);
    state.dataset.state = row.state;
    heading.append(element('code', 'fx-path', row.operation), state);
    item.append(heading, element('p', '', row.message));
    if (row.path)
        item.append(files.has(row.path) ? link(fileHref(navigation, row.path), row.path, 'fx-inline-link') : element('code', 'fx-path', row.path));
    return item;
}
/** The Diagnostics page: what the analysis could not do, filtered by a search, an operation and a state. */
export function diagnosticsView(details, navigation) {
    const { diagnostics, metrics } = details;
    const view = element('section', 'fx-view panel fx-events');
    const path = navigation.file;
    const scoped = path ? diagnostics.filter(row => row.path === path) : diagnostics;
    if (path) {
        const scope = element('div', 'fx-file-heading');
        scope.append(element('code', 'fx-path', path), link(navigation.href('diagnostics'), 'All diagnostics', 'fx-inline-link'));
        view.append(scope);
    }
    if (!scoped.length) {
        view.append(element('p', 'fx-note', path ? 'No analysis diagnostics for this file.' : 'No analysis diagnostics for this commit.'));
        return view;
    }
    const search = searchField('Search diagnostics', 'Operation, path or message');
    const operation = choice('Operation', choices('All operations', scoped.map(row => row.operation)), '');
    const state = choice('State', choices('All states', scoped.map(row => row.state), diagnosticStates), '');
    const tools = element('div', 'fx-tools');
    tools.append(search.label, operation.label, state.label);
    const result = element('p', 'fx-note fx-results');
    result.setAttribute('aria-live', 'polite');
    const events = element('ul', 'fx-event-list');
    const nav = navigationButtons('Diagnostic pages', () => { page--; render(); }, () => { page++; render(); });
    view.append(tools, result, events, nav.section);
    const files = new Set(metrics.files.map(entry => entry.path));
    let page = 0;
    const picked = (select, value) => !select.value || select.value === value;
    function render() {
        const query = search.input.value.trim().toLowerCase();
        const matches = scoped.filter(row => picked(state.select, row.state) && picked(operation.select, row.operation)
            && contains(query, [row.operation, row.path, row.message]));
        result.textContent = matches.length ? quantity(matches.length, 'diagnostic') : 'No diagnostics match these filters.';
        events.replaceChildren(...matches.slice(page * pageSize, (page + 1) * pageSize).map(row => diagnosticItem(row, navigation, files)));
        nav.section.hidden = matches.length <= pageSize;
        nav.update(page, Math.ceil(matches.length / pageSize));
    }
    const reset = () => { page = 0; render(); };
    search.input.addEventListener('input', reset);
    operation.select.addEventListener('change', reset);
    state.select.addEventListener('change', reset);
    render();
    return view;
}
