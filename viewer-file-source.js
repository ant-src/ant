import { segmented } from './viewer-controls.js';
import { copyPathButton } from './viewer-copy.js';
import { basename, count, element, externalLink, parentPath, store, stored } from './viewer-dom.js';
import { fileDetails } from './viewer-file-aside.js';
import { languageColor, languageOf } from './viewer-languages.js';
import { breadcrumbs, splitter } from './viewer-layout.js';
import { lineIn } from './viewer-locations.js';
import { isSourceLayer, lineReadings, setSourceLayer } from './viewer-line-coverage.js';
import { renderSource } from './viewer-source-view.js';
/** A file's line coverage, where the inspection's coverage report names the file. */
function fileCoverage(details, path) {
    const coverage = details.lineCoverage;
    return coverage && Object.hasOwn(coverage, path) ? coverage[path] : undefined;
}
const layerKey = 'repometry.source-layer';
/** The URL's choice of what lines show, else the one made earlier in this session, else the findings. */
function sourceLayer(requested) {
    if (isSourceLayer(requested))
        return requested;
    const remembered = stored(() => sessionStorage, layerKey);
    return isSourceLayer(remembered) ? remembered : 'issues';
}
/** The file's name with its full path a click away; a refused clipboard shows the path in the name's place, selected. */
function sourceName(path) {
    // The folder is in the breadcrumbs right above; the head names the file alone, with its full path on hover.
    const name = element('code', 'fx-source-path', basename(path));
    name.title = path;
    const named = element('span', 'fx-source-name');
    named.append(name, copyPathButton(path, () => {
        name.textContent = path;
        return name;
    }));
    return named;
}
function sourceMeta(file) {
    const language = languageOf(file);
    const dot = element('span', 'fx-language-dot');
    dot.style.background = languageColor(language);
    const meta = element('span', 'fx-source-meta');
    meta.append(dot, `${language} · ${count(file.lines)} lines`);
    return meta;
}
/** The findings that mark the file's lines, at their place in it. */
function sourceNotes(details, path) {
    return (details.findings ?? []).flatMap(finding => {
        const line = lineIn(finding, path);
        return line ? [{ line, level: finding.level, title: finding.rule, message: finding.message }] : [];
    });
}
/**
 * An open file: its source with the inspection laid on its lines, and its figures and functions beside it. The
 * functions lead into the code, and a chosen line goes into the URL so the place can be shared.
 */
export function fileView(details, navigation, file, load) {
    const view = element('div', 'fx-view');
    const keep = { query: navigation.query ?? '', group: navigation.group ?? '' };
    const nav = { ...navigation,
        href: (target, options) => target === 'metrics' ? navigation.href('metrics', { ...keep, ...options }) : navigation.href(target, options) };
    view.append(breadcrumbs('metrics', nav, parentPath(file.path)));
    const split = element('div', 'fx-split fx-split-source');
    const main = element('section', 'fx-source panel');
    main.setAttribute('aria-label', 'Source');
    const head = element('div', 'fx-source-head');
    const meta = sourceMeta(file);
    // Lines show the findings, or the tests' coverage where the file has it; without coverage the head stays as it was.
    const coverage = fileCoverage(details, file.path);
    let layer = coverage ? sourceLayer(navigation.layer) : 'issues';
    let line = navigation.line;
    const place = () => nav.href('metrics', { directory: parentPath(file.path), file: file.path, ...line ? { line } : {},
        ...layer === 'coverage' ? { layer } : {} });
    const body = element('div', 'fx-source-body');
    const toggle = coverage ? segmented('Lines show', [['issues', 'Issues'], ['coverage', 'Coverage']], layer, value => {
        layer = value;
        store(() => sessionStorage, layerKey, layer);
        setSourceLayer(body, layer);
        history.replaceState(history.state, '', place());
    }) : null;
    toggle?.classList.add('fx-source-layer');
    const tools = element('div', 'fx-source-tools');
    tools.append(...toggle ? [toggle] : [], meta);
    head.append(sourceName(file.path), tools);
    // The forge's page for the file, at the chosen line, whether or not the text opens here.
    const page = navigation.sourcePage?.(file.path, navigation.line);
    const forge = page ? externalLink(page.href, `View on ${page.provider}`, 'fx-source-forge') : null;
    if (forge)
        meta.append(' · ', forge);
    body.append(element('p', 'fx-note fx-source-wait', 'Loading source…'));
    main.append(head, body);
    let aside = fileDetails(details, file, nav);
    split.append(main, splitter(split), aside);
    view.append(split);
    const functions = details.functions.filter(record => record.path === file.path);
    const choose = (chosen) => {
        line = chosen;
        history.replaceState(history.state, '', place());
        const moved = navigation.sourcePage?.(file.path, chosen);
        if (forge && moved)
            forge.href = moved.href;
    };
    void load(file.path).then(text => {
        if (text === null) {
            // No locator, a private repository, a commit never pushed: one quiet state, and the forge's page when there is one.
            body.replaceChildren(element('p', 'fx-source-missing', 'Source not available'));
            // Nothing to shade: the head is the one a file without coverage has.
            toggle?.remove();
            return;
        }
        const reveal = renderSource(body, text, {
            path: file.path,
            functions: functions.flatMap(record => record.line === null ? [] : [{ line: record.line, complexity: record.complexity, name: record.name }]),
            notes: sourceNotes(details, file.path), line: navigation.line, choose,
            coverage: coverage && lineReadings(coverage), layer,
        });
        const next = fileDetails(details, file, nav, { text, reveal: target => { reveal(target); choose(target); } });
        aside.replaceWith(next);
        aside = next;
    });
    return view;
}
