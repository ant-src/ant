import { basename, count, element, fact, link, parentPath, quantity } from './viewer-dom.js';
import { coverageContext, coverageRatio, coverageText, duplicationContext, duplicationText, effortText } from './viewer-file-figures.js';
import { isLockfile, lockfilesLabel } from './viewer-languages.js';
import { tableColumns } from './viewer-columns.js';
import { partialNote } from './viewer-totals.js';
import { coverageKinds, localValue, measured, measuredFileMetrics } from './viewer-inspection.js';
import { atPath } from './viewer-locations.js';
import { format } from './viewer-metric-format.js';
import { labelOf, size } from './viewer-metrics.js';
import { lineSnippet } from './viewer-source-view.js';
import { highlighter } from './viewer-tokenizer.js';
/** A fact where the inspection took the measurement; a measurement it did not take leaves no row behind. */
function factIf(shown, label, value, context) {
    return shown ? [fact(label, value, context)] : [];
}
function panel(label, title, className = 'fx-aside') {
    const aside = element('aside', className);
    aside.setAttribute('aria-label', label);
    const heading = element('div', 'fx-aside-heading');
    heading.append(element('h2', '', title));
    return { panel: aside, heading };
}
/**
 * The totals of a folder, or of the files a query selected in it, with how the selection reads against its folder. A
 * total some of the files a plugin read do not carry says in how many it was measured.
 */
export function directorySummary(entry, files, selection) {
    const { panel: aside, heading } = panel(selection ? 'Selection totals' : 'Folder totals', selection ? 'Selection' : basename(entry.path) || 'All files');
    // The breadcrumbs above name the folder; only a selection needs its words here.
    if (selection)
        heading.append(element('p', 'fx-aside-path fx-aside-selection', selection));
    const facts = element('dl', 'fx-kv');
    const quality = entry.quality;
    const analysed = files.filter(file => file.analysed !== false).length;
    const lockfiles = files.filter(file => isLockfile(file.path)).length;
    const partial = analysed !== entry.files;
    // A total is partial as its table column is: the column says which files carry the measure.
    const presence = new Map(tableColumns().map(column => [column.label, column.present]));
    const note = (label) => presence.get(label) && partialNote(files, presence.get(label));
    facts.append(
    // Every tracked file counts here; Trends counts the analysed ones, so both say which they are.
    fact('Tracked files', count(entry.files)), ...factIf(partial, labelOf('files'), count(analysed)), ...factIf(lockfiles > 0, lockfilesLabel, count(lockfiles)), ...factIf(quality?.codeLines !== undefined, labelOf('code-lines'), count(quality?.codeLines), note(labelOf('code-lines'))), 
    // Physical lines are what the language bar weighs once some files are ignored, so its total stands here too.
    ...factIf(quality?.codeLines === undefined || partial, labelOf('lines'), count(entry.lines)), ...factIf(entry.bytes !== undefined, 'Size', size(entry.bytes)), ...factIf(entry.functions !== null, labelOf('functions'), count(entry.functions)), fact(labelOf('complexity'), count(entry.complexity?.maximum), note(labelOf('complexity'))), ...factIf(quality?.cognitiveComplexity !== undefined, labelOf('cognitive'), count(quality?.cognitiveComplexity), note(labelOf('cognitive'))), ...factIf(quality?.duplication !== undefined, labelOf('duplication'), duplicationText(entry), note(labelOf('duplication'))), ...factIf(quality?.maintainability !== undefined, labelOf('effort'), effortText(quality?.maintainability?.effortMinutes), note(labelOf('effort'))), ...factIf(coverageRatio(entry) !== null, labelOf('coverage'), coverageText(entry)), 
    // Other plugins' totals follow Repometry's own, where this inspection ran them and they roll up; the counts behind
    // a ratio read through it.
    ...measuredFileMetrics().filter(spec => spec.descriptor.aggregate !== 'none' && !spec.ratioPart)
        .map(spec => fact(spec.short, format(spec, localValue(entry, spec)), note(spec.short))));
    aside.append(heading, facts);
    return aside;
}
function fileFacts(details, entry) {
    const facts = element('dl', 'fx-kv');
    const quality = entry.quality;
    const window = details.metrics.churnWindow;
    const kinds = coverageKinds().flatMap(spec => {
        const value = measured(spec.descriptor) ? localValue(entry, spec) : null;
        return value === null ? [] : [fact(spec.short, format(spec, value))];
    });
    facts.append(...factIf(quality?.codeLines !== undefined, labelOf('code-lines'), count(quality?.codeLines)), fact(labelOf('lines'), count(entry.lines)), ...factIf(entry.bytes !== undefined, 'Size', size(entry.bytes)), ...factIf(quality?.commentLines !== undefined, labelOf('comments'), count(quality?.commentLines)), fact(labelOf('functions'), count(entry.functions)), ...factIf(Boolean(quality?.classes), labelOf('classes'), count(quality?.classes)), fact(labelOf('complexity'), count(entry.complexity?.maximum)), ...factIf(quality?.cognitiveComplexity !== undefined, labelOf('cognitive'), count(quality?.cognitiveComplexity)), fact(labelOf('coverage'), coverageText(entry), coverageContext(entry)), ...kinds, ...factIf(Boolean(quality?.duplication), labelOf('duplication'), duplicationText(entry), duplicationContext(entry)), ...factIf(Boolean(quality?.maintainability), labelOf('effort'), effortText(quality?.maintainability?.effortMinutes)), ...factIf(quality?.churn !== undefined, 'Git changes', count(quality?.churn), window ? `${count(window.lookbackDays)}-day window` : undefined));
    return facts;
}
/** Links to the file's findings and diagnostics, where it has some. */
function fileLinks(details, entry, navigation) {
    const findings = details.findings?.filter(item => atPath(item, entry.path)).length ?? 0;
    const links = element('p', 'fx-aside-links');
    if (findings)
        links.append(link(navigation.href('issues', { file: entry.path }), quantity(findings, 'finding'), 'fx-inline-link'));
    if (entry.diagnosticCount)
        links.append(link(navigation.href('diagnostics', { file: entry.path }), quantity(entry.diagnosticCount, 'diagnostic'), 'fx-inline-link'));
    return links.childElementCount ? [links] : [];
}
/** The side column has room for the functions, most complex first; a long file folds the rest away. */
function functionSection(details, path, hooks) {
    const functions = details.functions.filter(record => record.path === path).sort((a, b) => b.complexity - a.complexity);
    if (!functions.length)
        return [];
    const list = (records) => hooks.reveal ? sourceFunctionList(records, hooks.reveal, hooks.text) : functionList(records);
    const section = element('section', 'fx-aside-section');
    section.append(element('h3', '', 'Functions by complexity'), list(functions.slice(0, 10)));
    if (functions.length > 10) {
        const rest = element('details', 'fx-functions');
        rest.append(element('summary', '', `${functions.length - 10} more`), list(functions.slice(10)));
        section.append(rest);
    }
    return [section];
}
/** An open file's figures, the links to what it was charged with, and its functions. */
export function fileDetails(details, entry, navigation, hooks = {}) {
    // The breadcrumbs above name the folder, so the title is the file alone.
    const { panel: aside, heading } = panel('File details', basename(entry.path), 'fx-aside fx-file');
    const close = link(navigation.href('metrics', { directory: parentPath(entry.path) }), '', 'fx-close');
    close.setAttribute('aria-label', 'Close file details');
    close.title = 'Close file details';
    heading.append(close);
    aside.append(heading, fileFacts(details, entry), ...fileLinks(details, entry, navigation), ...functionSection(details, entry.path, hooks));
    return aside;
}
/** Functions as buttons that bring their first line into view; an anonymous one is named by that line's code. */
function sourceFunctionList(functions, reveal, text) {
    const list = element('ol', 'fx-function-list fx-function-jumps');
    list.append(...functions.map(record => {
        const row = element('li');
        const anonymous = record.name === '(anonymous)' && Boolean(text) && record.line !== null;
        const snippet = anonymous ? lineSnippet(text, record.line) : null;
        const button = element('button', 'fx-function-jump');
        button.type = 'button';
        button.disabled = record.line === null;
        const name = element(snippet ? 'code' : 'span', 'fx-function-name');
        if (snippet)
            name.append(...highlighter(record.path)(snippet));
        else
            name.textContent = record.name;
        button.append(name);
        if (record.line !== null)
            button.append(element('span', 'fx-function-line', `${record.line}`));
        button.addEventListener('click', () => { if (record.line !== null)
            reveal(record.line); });
        row.append(button, element('strong', 'fx-function-score', count(record.complexity)));
        return row;
    }));
    return list;
}
function functionList(functions) {
    const list = element('ol', 'fx-function-list');
    list.append(...functions.map(record => {
        const row = element('li');
        const subject = element('div');
        subject.append(element('strong', '', record.line === null ? record.name : `${record.name} · line ${record.line}`));
        row.append(subject, element('strong', 'fx-function-score', count(record.complexity)));
        return row;
    }));
    return list;
}
