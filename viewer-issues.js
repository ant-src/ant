import { choices, multiChoice, navigationButtons, searchField, segmented } from './viewer-controls.js';
import { basename, collator, contains, count, element, fact, link, quantity } from './viewer-dom.js';
import { pageSize } from './viewer-inspection.js';
import { effortOf, issueCategories, issueGroup, issueLevels, levelOrder } from './viewer-issue-group.js';
import { splitter } from './viewer-layout.js';
import { atPath, locationPaths } from './viewer-locations.js';
import { duration, labelOf } from './viewer-metrics.js';
const severities = ['high', 'medium', 'low', 'unknown'];
/** How often each value occurs: the tally a multiple choice offers. */
function tally(values, labels) {
    return choices('', values, labels).slice(1).map(([value, name]) => [value, name, values.filter(item => item === value).length]);
}
/** Whether a multiple choice lets the value through: none ticked means all. */
function allowed(values, value) {
    return values.size === 0 || values.has(value);
}
function searchable(item) {
    return [item.rule, item.message, item.path, ...(item.related?.map(location => location.path) ?? [])];
}
/** The findings grouped by rule: the most severe first, then the most frequent, then by name. */
function byRule(findings) {
    const groups = new Map();
    for (const item of findings) {
        const key = JSON.stringify([item.tool, item.rule, item.category, item.level]);
        const group = groups.get(key);
        if (group)
            group.push(item);
        else
            groups.set(key, [item]);
    }
    return [...groups.values()].sort((a, b) => levelOrder[a[0].level] - levelOrder[b[0].level]
        || b.length - a.length || collator.compare(a[0].rule, b[0].rule));
}
function summaryFacts(findings, rules) {
    const facts = element('dl', 'fx-kv');
    facts.append(fact('Findings', count(findings.length)), fact('Rules', count(rules)));
    const effort = effortOf(findings);
    if (effort !== undefined)
        facts.append(fact(labelOf('effort'), duration(effort)));
    // One severity would only repeat the total.
    const levels = severities.map(value => [value, findings.filter(item => item.level === value).length]).filter(([, total]) => total > 0);
    if (levels.length > 1)
        facts.append(...levels.map(([value, total]) => fact(issueLevels[value], count(total))));
    return facts;
}
function asideSection(title, content) {
    const section = element('section', 'fx-aside-section');
    section.append(element('h3', '', title), content);
    return section;
}
function categorySection(findings) {
    const categories = new Map();
    for (const item of findings)
        categories.set(item.category, (categories.get(item.category) ?? 0) + 1);
    if (categories.size < 2)
        return [];
    const rows = element('dl', 'fx-kv');
    rows.append(...[...categories].sort((a, b) => b[1] - a[1]).map(([name, total]) => fact(issueCategories[name] ?? name, count(total))));
    return [asideSection('By category', rows)];
}
function placesSection(findings, navigation, files) {
    const places = new Map();
    for (const item of findings)
        for (const location of locationPaths(item))
            places.set(location, (places.get(location) ?? 0) + 1);
    if (!places.size)
        return [];
    const top = element('ol', 'fx-function-list fx-top-files');
    for (const [file, total] of [...places].sort((a, b) => b[1] - a[1] || collator.compare(a[0], b[0])).slice(0, 5)) {
        const row = element('li');
        const name = files.has(file) ? link(navigation.href('issues', { file }), basename(file), 'fx-top-file') : element('span', 'fx-top-file', basename(file));
        name.title = file;
        row.append(name, element('strong', 'fx-function-score', count(total)));
        top.append(row);
    }
    return [asideSection('Most affected files', top)];
}
/** Totals for the findings the filters select: how many, how much work, of what kind, and where they gather. */
function issueSummary(findings, rules, navigation, files) {
    const panel = element('aside', 'fx-aside');
    panel.setAttribute('aria-label', 'Findings summary');
    const heading = element('div', 'fx-aside-heading');
    heading.append(element('h2', '', 'Findings'));
    panel.append(heading, summaryFacts(findings, rules), ...categorySection(findings), ...placesSection(findings, navigation, files));
    return panel;
}
/** A file's findings name the file, with the way back to all of them. */
function fileScope(path, navigation) {
    const scope = element('div', 'fx-file-heading');
    scope.append(element('code', 'fx-path', path), link(navigation.href('issues'), 'All findings', 'fx-inline-link'));
    return scope;
}
/** The Findings page: the inspection's findings grouped by rule, filtered by a search, a severity, tools and categories. */
export function issuesView(details, navigation) {
    const { findings, metrics } = details;
    const view = element('div', 'fx-view fx-events');
    if (findings === undefined) {
        view.append(element('p', 'fx-note', 'Findings are unavailable for this inspection. Reinspect this commit to collect them.'));
        return view;
    }
    const path = navigation.file;
    const scoped = path ? findings.filter(item => atPath(item, path)) : findings;
    if (path)
        view.append(fileScope(path, navigation));
    if (!scoped.length) {
        view.append(element('p', 'fx-note', path ? 'No findings recorded for this file.' : 'No findings recorded for this commit.'));
        return view;
    }
    // One line of controls, as on Metrics: a search, the severities as a switch, and selects that appear only where they
    // can narrow something. The counts live in the sidebar, which follows the filters.
    const search = searchField('Search findings', 'Search rules, messages and paths');
    search.label.classList.add('fx-field-bare');
    // A link may arrive with a search, as the Profile's rules do: the list opens narrowed to it.
    search.input.value = navigation.query ?? '';
    const present = severities.filter(value => scoped.some(item => item.level === value));
    // A link may name a severity and categories, as the Overview's figures do: the list opens filtered, the controls show it.
    let level = present.find(value => value === navigation.level) ?? '';
    const severity = segmented('Severity', [['', 'All'], ...present.map(value => [value, issueLevels[value]])], level, value => {
        level = value;
        reset();
    });
    const tools = multiChoice({ text: 'Tool', all: 'All tools', plural: 'tools' }, tally(scoped.map(item => item.tool)), reset);
    const category = multiChoice({ text: 'Category', all: 'All categories', plural: 'categories' }, tally(scoped.map(item => item.category), issueCategories), reset, navigation.category?.split(','));
    const toolbar = element('div', 'fx-tools fx-metrics-tools');
    toolbar.append(search.label);
    if (present.length > 1)
        toolbar.append(severity);
    toolbar.append(category.element, tools.element);
    const status = element('p', 'fx-note fx-results');
    status.setAttribute('aria-live', 'polite');
    const list = element('div', 'fx-issue-list');
    const nav = navigationButtons('Finding pages', () => { page--; render(); }, () => { page++; render(); });
    const split = element('div', 'fx-split');
    const main = element('div', 'fx-main');
    main.append(toolbar, status, list, nav.section);
    let aside = element('aside');
    split.append(main, splitter(split), aside);
    view.append(split);
    const files = new Set(metrics.files.map(entry => entry.path));
    let page = 0;
    const passes = (item, query) => allowed(tools.values, item.tool) && allowed(category.values, item.category)
        && (!level || item.level === level) && contains(query, searchable(item));
    function render() {
        const query = search.input.value.trim().toLowerCase();
        const chosen = scoped.filter(item => passes(item, query));
        const sorted = byRule(chosen);
        list.replaceChildren(...sorted.slice(page * pageSize, (page + 1) * pageSize).map(group => issueGroup(group, navigation, files)));
        // The sidebar carries the counts; on screen this line speaks only when nothing matches, and it tells a screen
        // reader what the filters left.
        status.textContent = chosen.length === 0 ? 'No findings match these filters.' : `${quantity(chosen.length, 'finding')} in ${quantity(sorted.length, 'rule')}`;
        status.classList.toggle('visually-hidden', chosen.length > 0);
        nav.section.hidden = sorted.length <= pageSize;
        nav.update(page, Math.ceil(sorted.length / pageSize));
        const next = issueSummary(chosen, sorted.length, navigation, files);
        aside.replaceWith(next);
        aside = next;
    }
    function reset() {
        page = 0;
        remember();
        render();
    }
    /** The URL follows the filters, without a new history entry per change, so the view can be shared and reloaded. */
    function remember() {
        history.replaceState(history.state, '', navigation.href('issues', {
            ...(path ? { file: path } : {}), query: search.input.value.trim(), level,
            category: [...category.values].join(','),
        }));
    }
    search.input.addEventListener('input', reset);
    render();
    return view;
}
