// A list that may run to tens of thousands of rows, as a profile's rules and patterns do in a monorepo.
import { multiChoice, navigationButtons, searchField } from './viewer-controls.js';
import { count, element } from './viewer-dom.js';
const tableSize = 50;
const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' });
/** Orders by a column: numbers by value, anything else as people read it. */
function byColumn(sort, ascending) {
    const direction = ascending ? 1 : -1;
    return (a, b) => {
        const left = sort(a);
        const right = sort(b);
        return direction * (typeof left === 'number' && typeof right === 'number' ? left - right : collator.compare(String(left), String(right)));
    };
}
/** The count beside the page buttons, as in Metrics: the page's range, how many the search kept, or how many there are. */
export function tableSummary(noun, shown, total, first, visible) {
    const [one, many] = noun;
    if (shown === 0)
        return `No ${many} match.`;
    if (shown > tableSize)
        return `${count(first + 1)}–${count(first + visible)} of ${count(shown)} ${many}`;
    if (shown !== total)
        return `${count(shown)} of ${count(total)} ${many}`;
    return `${count(shown)} ${shown === 1 ? one : many}`;
}
/** The facet's values with how many rows each has, by name. */
function facetOptions(rows, key) {
    const counts = new Map();
    for (const row of rows)
        counts.set(key(row), (counts.get(key(row)) ?? 0) + 1);
    return [...counts].sort((a, b) => collator.compare(a[0], b[0])).map(([value, total]) => [value, value, total]);
}
function tableRow(row, columns) {
    const line = element('tr');
    for (const column of columns) {
        const cell = element('td', column.numeric ? 'numeric' : '');
        cell.append(column.cell(row));
        line.append(cell);
    }
    return line;
}
/**
 * A list searched, narrowed by one facet, sorted by any column, and paged, with the count beside the page buttons as
 * in Metrics. Only the page in view is built.
 */
export function pagedTable(spec) {
    const box = element('div', 'profile-table');
    if (!spec.rows.length) {
        box.append(element('p', 'detail', spec.empty));
        return box;
    }
    const haystacks = spec.rows.map(row => spec.search(row).toLowerCase());
    const state = { page: 0, sortBy: -1, ascending: true };
    const reset = () => { state.page = 0; render(); };
    const tools = element('div', 'fx-tools profile-table-tools');
    const search = searchField(`Search ${spec.noun[1]}`, `Search ${spec.noun[1]}`);
    search.label.classList.add('fx-field-bare');
    search.input.addEventListener('input', reset);
    const facet = spec.facet && multiChoice(spec.facet, facetOptions(spec.rows, spec.facet.key), reset);
    // A short list needs no controls: it is read at a glance.
    if (spec.controls || spec.rows.length > 10)
        tools.append(search.label);
    if (facet && !facet.element.hidden)
        tools.append(facet.element);
    const table = element('table', 'fx-table profile-list');
    table.setAttribute('aria-label', spec.label);
    const head = element('tr');
    const headers = spec.columns.map((column, index) => {
        const cell = element('th', column.numeric ? 'numeric' : '');
        cell.scope = 'col';
        const button = element('button', 'fx-sort', column.label);
        button.type = 'button';
        button.addEventListener('click', () => {
            state.ascending = state.sortBy === index ? !state.ascending : !column.numeric;
            state.sortBy = index;
            reset();
        });
        cell.append(button);
        head.append(cell);
        return cell;
    });
    const thead = element('thead');
    thead.append(head);
    const body = element('tbody');
    table.append(thead, body);
    const frame = element('div', 'table-scroll fx-table-scroll');
    frame.append(table);
    const pager = element('div', 'fx-pager');
    const result = element('p', 'fx-note fx-results');
    result.setAttribute('aria-live', 'polite');
    const buttons = navigationButtons(`${spec.label} pages`, () => { state.page -= 1; render(); }, () => { state.page += 1; render(); });
    pager.append(result, buttons.section);
    if (tools.childElementCount)
        box.append(tools);
    box.append(frame, pager);
    function selected() {
        const words = search.input.value.trim().toLowerCase().split(/\s+/).filter(Boolean);
        const chosen = facet ? facet.values : new Set();
        const inFacet = (row) => chosen.size === 0 || chosen.has(spec.facet.key(row));
        return spec.rows.filter((row, index) => inFacet(row) && words.every(word => haystacks[index].includes(word)));
    }
    function render() {
        const shown = selected();
        if (state.sortBy >= 0)
            shown.sort(byColumn(spec.columns[state.sortBy].sort, state.ascending));
        for (const [index, cell] of headers.entries()) {
            if (index === state.sortBy)
                cell.setAttribute('aria-sort', state.ascending ? 'ascending' : 'descending');
            else
                cell.removeAttribute('aria-sort');
        }
        const pages = Math.max(1, Math.ceil(shown.length / tableSize));
        state.page = Math.min(state.page, pages - 1);
        const first = state.page * tableSize;
        const visible = shown.slice(first, first + tableSize);
        body.replaceChildren(...visible.map(row => tableRow(row, spec.columns)));
        frame.hidden = shown.length === 0;
        result.textContent = tableSummary(spec.noun, shown.length, spec.rows.length, first, visible.length);
        buttons.section.hidden = shown.length <= tableSize;
        buttons.update(state.page, pages);
    }
    render();
    return box;
}
