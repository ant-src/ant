import { renderSunburst } from './viewer-code-map.js';
import { choice, navigationButtons } from './viewer-controls.js';
import { basename, collator, count, element, inside, link, parentPath } from './viewer-dom.js';
import { fileHref, hotspotScore, scorePeak } from './viewer-file-figures.js';
import { localValue, measured } from './viewer-inspection.js';
import { breadcrumbs, liftToTitle } from './viewer-layout.js';
import { band, bandLimits, rankTitle } from './viewer-map-scale.js';
import { bandScale, compareValues, firstSortAscending } from './viewer-metric-format.js';
import { labelOf, pluginMetrics, pluginSpec } from './viewer-metrics.js';
const maxRanked = 10;
const hotspotChoices = [
    ['complexity', labelOf('complexity')],
    ['cognitive', labelOf('cognitive')],
    ['missedLines', 'Uncovered lines'],
    ['duplication', 'Duplicated lines'],
    ['churn', 'Git churn'],
];
/**
 * What Hotspots ranks by: Repometry's rankings, then every other plugin's metric this inspection ran that has a worse
 * end to rank by. The window is part of what churn means, so it is in the metric's name rather than in a sentence.
 */
function rankChoices(days) {
    const churn = days ? `Changes in ${count(days)} days` : undefined;
    return [
        ...hotspotChoices.map(([value, name]) => [value, value === 'churn' ? churn ?? name : name]),
        ...pluginMetrics().filter(spec => spec.hotspot && measured(spec.descriptor)).map(spec => [spec.key, spec.label]),
    ];
}
/** What makes the churn count incomplete, when anything does: missing history, a shallow checkout, a capped scan. */
function churnCaveat(window) {
    if (!window)
        return 'Git history is unavailable for this inspection.';
    return [window.shallow ? 'The checkout is shallow, so older changes are missing.' : '',
        window.capReached ? `Only the latest ${count(window.maxCommits)} commits were scanned.` : ''].filter(Boolean).join(' ');
}
/**
 * Files by another plugin's metric, worst first. Where lower is better a zero is nothing to fix and stays off the list,
 * as with Repometry's own rankings; where higher is better every measured file ranks.
 */
function pluginRanking(files, spec) {
    const order = compareValues(firstSortAscending(spec));
    const ranks = (value) => value !== null && (spec.descriptor.better === 'higher' || value > 0);
    return files.flatMap(entry => {
        const value = localValue(entry, spec);
        return ranks(value) ? [{ entry, value }] : [];
    }).sort((a, b) => order(a.value, b.value) || collator.compare(a.entry.path, b.entry.path)).map(({ entry }) => entry);
}
function ranked(details, files, metric) {
    const plugin = pluginSpec(metric);
    if (plugin)
        return pluginRanking(files, plugin);
    const byPath = new Map(files.map(entry => [entry.path, entry]));
    return (details.metrics.hotspots[metric] ?? []).flatMap(path => byPath.get(path) ?? []);
}
function rankRow(entry, metric, navigation, peak, limits) {
    const row = element('li');
    const title = link(fileHref(navigation, entry.path), basename(entry.path), 'fx-rank-link');
    title.append(element('span', 'fx-rank-path', parentPath(entry.path) || '/'));
    const value = hotspotScore(entry, metric);
    const mark = element('span', 'fx-rank-mark');
    // An unmeasured file has no bar and no band: a zero-length bar would read as a measured zero.
    mark.style.setProperty('--length', `${(value ?? 0) / peak * 100}%`);
    if (value !== null)
        mark.dataset.band = String(band(entry, metric, value, peak, limits));
    mark.setAttribute('aria-hidden', 'true');
    row.append(title, mark, element('strong', 'fx-rank-score', count(value)));
    return row;
}
function rankFiles(entries, metric, navigation, hasMeasures, limits) {
    const panel = element('section', 'fx-rank');
    // The bars speak the map's colours: the same bands, the same palette.
    const plugin = pluginSpec(metric);
    panel.dataset.palette = plugin ? bandScale(plugin) : metric === 'churn' ? 'neutral' : 'heat';
    const heading = rankTitle(metric);
    panel.append(element('h2', '', `${heading} by file`));
    if (!entries.length) {
        panel.append(element('p', 'fx-note', hasMeasures
            ? `No files with ${heading.toLowerCase()} above zero.`
            : `No measurements for ${heading.toLowerCase()} in this folder.`));
        return panel;
    }
    const list = element('ol', 'fx-ranking');
    const pages = navigationButtons('Ranking pages', () => { page--; render(); }, () => { page++; render(); });
    pages.section.hidden = entries.length <= maxRanked;
    // The longest bar is the highest value, whichever end of the scale the list starts from.
    const peak = scorePeak(entries.map(entry => hotspotScore(entry, metric)));
    let page = 0;
    function render() {
        const offset = page * maxRanked;
        list.start = offset + 1;
        list.replaceChildren(...entries.slice(offset, offset + maxRanked).map(entry => rankRow(entry, metric, navigation, peak, limits)));
        pages.update(page, Math.ceil(entries.length / maxRanked));
    }
    panel.append(list, pages.section);
    render();
    return panel;
}
/** The Hotspots page: the folder's code map coloured by a metric, and its files ranked by it beside. */
export function hotspotView(details, navigation, settings, directory) {
    const view = element('div', 'fx-view');
    const choices = rankChoices(details.metrics.churnWindow?.lookbackDays);
    const linked = choices.find(([value]) => value === navigation.rank)?.[0];
    if (linked)
        settings.hotspot = linked;
    if (!choices.some(([value]) => value === settings.hotspot))
        settings.hotspot = 'complexity';
    const selector = choice('Rank by', choices, settings.hotspot);
    selector.label.classList.add('fx-field-bare');
    const tools = element('div', 'fx-tools fx-hotspot-tools');
    tools.append(breadcrumbs('hotspots', navigation, directory));
    const content = element('div', 'fx-hotspot-content');
    const provenance = element('p', 'fx-note fx-churn-provenance');
    view.append(tools, provenance, content);
    selector.select.addEventListener('change', () => {
        settings.hotspot = choices.find(([value]) => value === selector.select.value)?.[0] ?? 'complexity';
        history.replaceState(history.state, '', navigation.href('hotspots', { directory, rank: settings.hotspot }));
        render();
    });
    const files = details.metrics.files.filter(entry => inside(entry.path, directory));
    function render() {
        const metric = settings.hotspot;
        const caveat = metric === 'churn' ? churnCaveat(details.metrics.churnWindow) : '';
        provenance.hidden = !caveat;
        provenance.textContent = caveat;
        const hasMeasures = files.some(entry => hotspotScore(entry, metric) !== null);
        const mapSection = renderSunburst(details, { metric, directory, href: navigation.href, reshape: () => render(), controls: selector.label });
        content.replaceChildren(mapSection, rankFiles(ranked(details, files, metric), metric, navigation, hasMeasures, bandLimits(details, metric)));
    }
    // The ranking is the page's side column: its heading stands on the page title's baseline, as a sidebar's does.
    liftToTitle(content, ':scope > .fx-rank h2');
    render();
    return view;
}
