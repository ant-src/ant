import { renderSunburst } from './viewer-code-map.js';
import { byId, count, notice, parentPath, quantity } from './viewer-dom.js';
import { hotspotScore, scorePeak } from './viewer-file-figures.js';
import { securityCategories } from './viewer-files.js';
import { band, bandLimits } from './viewer-map-scale.js';
import { mapStub } from './viewer-map-shape.js';
import { securityGrades, securityRisks } from './viewer-metrics.js';
import { href } from './viewer-route.js';
const titles = {
    complexity: 'Maintainability hotspots',
    cognitive: 'Cognitive complexity hotspots',
    duplication: 'Duplication hotspots',
    missedLines: 'Coverage hotspots',
};
/** What the Overview ranks for each quality area: coverage by uncovered lines, the rest by complexity. */
function hotspotOf(area) {
    if (area === 'coverage')
        return 'missedLines';
    return area === 'cognitive' || area === 'duplication' ? area : 'complexity';
}
/** The ranking's heading and link show only when there is something to rank. */
function listHeading(visible) {
    byId('overview-hotspots-title').parentElement.hidden = !visible;
}
/** Nothing measured: the grey map stands alone; the quality area's tab already says what is missing. */
function unranked() {
    listHeading(false);
    byId('overview-hotspot-list').replaceChildren();
}
function allLink(target, visible) {
    const all = document.querySelector('#all-hotspots');
    all.href = target;
    all.hidden = !visible;
}
/** The compact map, drawn again in place when its shape changes. */
function overviewMap(details, sha, options) {
    const chart = byId('overview-sunburst');
    const map = () => chart.replaceChildren(renderSunburst(details, {
        ...options, compact: true, href: (page, link) => href({ page, sha, ...link }), reshape: map,
    }));
    map();
}
/** A ranked file: its name over its folder, linked, with its score in a chip. A file in the root has no folder to name. */
function hotspotRow(path, target, chip) {
    const link = document.createElement('a');
    link.href = target;
    const label = document.createElement('span');
    label.className = 'hotspot-name';
    const name = document.createElement('strong');
    name.textContent = path.slice(path.lastIndexOf('/') + 1);
    label.append(name);
    const directory = parentPath(path);
    if (directory) {
        const parent = document.createElement('small');
        parent.textContent = directory;
        label.append(parent);
    }
    link.append(label, chip);
    const row = document.createElement('li');
    row.append(link);
    return row;
}
function hotspotList(rows) {
    const list = document.createElement('ol');
    list.className = 'overview-hotspot-list';
    list.append(...rows);
    return list;
}
function rankedFiles(details, metric) {
    const byPath = new Map(details.metrics.files.map(file => [file.path, file]));
    return (details.metrics.hotspots[metric] ?? []).slice(0, 5).map(path => {
        const entry = byPath.get(path);
        return { path, value: entry ? hotspotScore(entry, metric) : null, entry };
    });
}
/** The Overview's hotspot column for the selected quality area: the compact map and the top five files. */
export function renderOverviewHotspots(details, sha, area, scanned) {
    const aside = byId('overview-hotspots');
    aside.removeAttribute('aria-busy');
    aside.classList.remove('overview-stale');
    listHeading(true);
    if (area === 'security') {
        renderSecurityHotspots(details, sha, scanned);
        return;
    }
    const metric = hotspotOf(area);
    byId('overview-hotspots-title').textContent = titles[metric];
    overviewMap(details, sha, { metric });
    allLink(href({ page: 'hotspots', sha }), true);
    const top = rankedFiles(details, metric);
    const peak = scorePeak(top.map(item => item.value));
    const rows = top.map(({ path, value, entry }) => {
        const score = document.createElement('span');
        score.className = 'chip';
        score.textContent = count(value);
        if (value !== null && entry)
            score.dataset.band = String(band(entry, metric, value, peak, bandLimits(details, metric)));
        return hotspotRow(path, href({ page: 'metrics', sha, file: path, directory: parentPath(path) }), score);
    });
    const output = byId('overview-hotspot-list');
    if (rows.length)
        output.replaceChildren(hotspotList(rows));
    else if (metric === 'missedLines' && !details.metrics.files.some(file => file.lineCoverage))
        unranked();
    else
        notice(output, 'No files in this ranking.');
}
/** Each file's worst finding, in the steps of the security mark; until findings carry a critical level, high is the top. */
const levelSteps = { low: 1, unknown: 1, medium: 2, high: 3 };
function securityFindings(details) {
    const sensitive = new Set(securityCategories);
    return (details.findings ?? []).filter((finding) => sensitive.has(finding.category) && Boolean(finding.path));
}
/**
 * Security has no measure of its own on the map: the map shows sizes in grey, and the list ranks files by their
 * scanner findings. Without a scanner report there is nothing to rank, and the grey map says so.
 */
function renderSecurityHotspots(details, sha, scanned) {
    byId('overview-hotspots-title').textContent = 'Security hotspots';
    const findings = securityFindings(details);
    const levels = new Map();
    const counts = new Map();
    for (const finding of findings) {
        levels.set(finding.path, Math.max(levels.get(finding.path) ?? 0, levelSteps[finding.level]));
        counts.set(finding.path, (counts.get(finding.path) ?? 0) + 1);
    }
    overviewMap(details, sha, { metric: 'complexity', ...scanned ? { levels } : { muted: true } });
    allLink(href({ page: 'issues', sha }), scanned);
    if (!scanned) {
        unranked();
        return;
    }
    // The worst level first, then the most findings; the chip names the level, as the map colours it.
    const top = [...levels].sort((a, b) => b[1] - a[1] || counts.get(b[0]) - counts.get(a[0]) || a[0].localeCompare(b[0])).slice(0, 5);
    const output = byId('overview-hotspot-list');
    if (!top.length) {
        notice(output, 'No security findings.');
        return;
    }
    output.replaceChildren(hotspotList(top.map(([path, level]) => {
        const score = document.createElement('span');
        score.className = 'chip chip-level';
        score.dataset.level = String(level);
        score.textContent = securityGrades[level];
        score.title = `${securityRisks[level]} · ${quantity(counts.get(path), 'finding')}`;
        return hotspotRow(path, href({ page: 'issues', sha, query: path }), score);
    })));
}
/**
 * The hotspot column while the inspection loads. A map and ranking already on screen stay, dimmed, until the next ones
 * replace them; on a first visit placeholders of the same size hold the column, so nothing beside it moves.
 */
export function overviewStub(sha) {
    const chart = byId('overview-sunburst');
    const output = byId('overview-hotspot-list');
    allLink(href({ page: 'hotspots', sha }), true);
    const aside = byId('overview-hotspots');
    aside.setAttribute('aria-busy', 'true');
    listHeading(true);
    if (chart.querySelector('.fx-sunburst:not(.fx-map-stub)') && output.querySelector('.overview-hotspot-list')) {
        aside.classList.add('overview-stale');
        return;
    }
    chart.replaceChildren(mapStub());
    const list = document.createElement('ol');
    list.className = 'overview-hotspot-list overview-hotspot-stub';
    for (let row = 0; row < 5; row++) {
        const item = document.createElement('li');
        item.innerHTML = '<span><span class="stub-line"></span><span class="stub-line stub-line-short"></span></span><span class="stub-chip"></span>';
        list.append(item);
    }
    output.replaceChildren(list);
}
