import { byId, count, percent, quantity } from './viewer-dom.js';
import { codeSmellCategories, securityCategories } from './viewer-files.js';
import { coverageGrade, formatMetric, gradeScope, metricValue, pluginSpec, securityGrades, securityRisks } from './viewer-metrics.js';
import { href } from './viewer-route.js';
/** Coverage kinds beyond lines, by their registry keys, in the words the Overview card uses for them. */
const coverageKinds = [['branch-coverage', 'Branches'], ['function-coverage', 'Functions']];
/** A detail that counts something leads to exactly what it counts, where the report has a page for it. */
export function detailText(id, text, link) {
    if (link === undefined) {
        byId(id).textContent = text;
        return;
    }
    const anchor = document.createElement('a');
    anchor.href = link;
    anchor.textContent = text;
    byId(id).replaceChildren(anchor);
}
/** Where each Overview figure leads: the findings it counts, the files it ranks, the function it names. */
export function smellsLink(sha) {
    return href({ page: 'issues', sha, category: codeSmellCategories.join(',') });
}
export function securityLink(sha, level) {
    return href({ page: 'issues', sha, category: securityCategories.join(','), ...level ? { level } : {} });
}
export function uncoveredLink(sha) {
    return href({ page: 'hotspots', sha, rank: 'missedLines' });
}
/** A small progress ring for a ratio; an empty ring when there is no value. */
function ring(ratio) {
    const namespace = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(namespace, 'svg');
    svg.setAttribute('viewBox', '0 0 40 40');
    const track = document.createElementNS(namespace, 'circle');
    const arc = document.createElementNS(namespace, 'circle');
    for (const circle of [track, arc]) {
        circle.setAttribute('cx', '20');
        circle.setAttribute('cy', '20');
        circle.setAttribute('r', '16');
    }
    track.setAttribute('class', 'ring-track');
    arc.setAttribute('class', 'ring-arc');
    const length = 2 * Math.PI * 16;
    arc.setAttribute('stroke-dasharray', `${(ratio ?? 0) * length} ${length}`);
    // Flat ends keep the gap true to the value: a round cap would close 95.9% into a full circle and turn
    // a zero into a dot. An unknown or zero ratio shows the track alone.
    svg.append(...(ratio !== null && ratio > 0 ? [track, arc] : [track]));
    return svg;
}
/** Line coverage leads; the other kinds the coverage report carries follow on a quieter line. */
function renderCoverageKinds(report) {
    const kinds = coverageKinds.flatMap(([key, name]) => {
        const metric = key;
        const value = pluginSpec(metric) ? metricValue(report, metric, undefined) : null;
        if (value === null)
            return [];
        // Each kind keeps its number; a narrow card wraps between kinds, never inside one.
        const kind = document.createElement('span');
        kind.textContent = `${name} ${formatMetric(value, metric)}`;
        return [kind];
    });
    byId('coverage-kinds').replaceChildren(...kinds.flatMap((kind, index) => index ? [' · ', kind] : [kind]));
    byId('coverage-kinds').hidden = kinds.length === 0;
}
export function renderInspection(report) {
    const lines = report.lineCoverage;
    const coverage = lines && lines.total > 0 ? lines.covered / lines.total : null;
    if (coverage === null)
        detailText('coverage-detail', 'No coverage report');
    else
        detailText('coverage-detail', `${percent(coverage)} of ${count(lines.total)} lines`, coverage < 1 ? uncoveredLink(report.sha) : undefined);
    const mark = byId('coverage-mark');
    mark.replaceChildren(ring(coverage));
    // The ring is coloured by the grade the coverage badge gives it.
    if (coverage === null)
        delete mark.dataset.grade;
    else
        mark.dataset.grade = coverageGrade(coverage);
    renderCoverageKinds(report);
}
export function renderQuality(report, quality) {
    const maintainability = quality?.maintainability ?? null;
    const grade = byId('maintainability');
    grade.textContent = maintainability?.grade ?? '–';
    grade.dataset.grade = maintainability?.grade ?? 'none';
    if (!maintainability)
        detailText('maintainability-detail', 'Not measured');
    else
        detailText('maintainability-detail', quantity(maintainability.smells, 'code smell'), maintainability.smells > 0 ? smellsLink(report.sha) : undefined);
    const scope = byId('maintainability-scope');
    scope.textContent = gradeScope(quality) ?? '';
    scope.hidden = !scope.textContent;
}
/** The words of the five steps, as every grade has them: clear A, low B, medium C, high D, critical F. */
const severities = ['clear', 'low', 'medium', 'high', 'critical'];
/** A grade like maintainability's, by the worst finding: 4 critical, 3 high, 2 medium, 1 anything else, 0 none. */
function securityStep(security) {
    const worst = [security.critical ?? 0, security.high, security.medium].findIndex(amount => amount > 0);
    if (worst >= 0)
        return 4 - worst;
    return security.low + security.unknown > 0 ? 1 : 0;
}
function securityText(security) {
    if (!security)
        return { text: 'No scanner report', counted: false };
    const levels = [['critical', security.critical ?? 0], ['high', security.high], ['medium', security.medium], ['low', security.low],
        ['unclassified', security.unknown]].filter(([, amount]) => amount > 0);
    if (levels.length === 0)
        return { text: 'No findings', counted: false };
    return { text: levels.map(([name, amount]) => `${count(amount)} ${name}`).join(' · '), counted: true };
}
export function renderScannerContext(context, sha) {
    const security = context?.security;
    const mark = byId('security-mark');
    const step = security ? securityStep(security) : null;
    mark.textContent = step === null ? '–' : securityGrades[step];
    mark.title = step === null ? '' : `Security ${securityGrades[step]}: ${securityRisks[step].toLowerCase()}`;
    mark.dataset.severity = step === null ? 'none' : severities[step];
    // The mark carries the level; the text gives the counts by severity.
    const { text, counted } = securityText(security);
    detailText('security-detail', text, counted ? securityLink(sha) : undefined);
}
