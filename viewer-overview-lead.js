import { byId, parentPath } from './viewer-dom.js';
import { difference, formatMetric, labelOf, metricValue } from './viewer-metrics.js';
import { securityLink, smellsLink, uncoveredLink } from './viewer-overview-cards.js';
import { href } from './viewer-route.js';
import { commitTime } from './viewer-trend-data.js';
const dateTime = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });
/**
 * Each quality area brings its own figures under the chart, and a second series drawn in grey behind the first:
 * maintainability pairs the debt ratio with the remediation time, coverage pairs the ratio with the uncovered lines.
 */
const areas = {
    debt: ['debt', 'effort', 'duplication', 'complexity'],
    coverage: ['coverage', 'coverable-lines', 'covered-lines', 'uncovered-lines'],
    security: ['security', 'security-high', 'security-medium', 'security-low'],
    complexity: ['complexity', 'functions', 'cognitive', 'code-lines'],
};
const emptyAreas = {
    coverage: 'No coverage report for this commit.',
    security: 'No security scanner report for this commit.',
};
/** The figures that count findings lead to them; uncovered lines lead to the files that hold them. */
const findingLinks = {
    effort: smellsLink,
    duplication: sha => href({ page: 'issues', sha, category: 'duplication' }),
    'uncovered-lines': uncoveredLink,
    security: sha => securityLink(sha),
    'security-high': sha => securityLink(sha, 'high'),
    'security-medium': sha => securityLink(sha, 'medium'),
    'security-low': sha => securityLink(sha, 'low'),
};
/** The most complex function's line, where the details name a function of that complexity. */
function functionLink(value, sha, details) {
    const top = details?.functions.find(item => item.complexity === value);
    if (!top)
        return undefined;
    return href({ page: 'metrics', sha, file: top.path, directory: parentPath(top.path), ...top.line ? { line: top.line } : {} });
}
/**
 * Where a lead figure leads, when it counts things the report lists: remediation time to the code smells it prices,
 * duplication to the duplicated code, uncovered lines to the files that hold them, security counts to the findings,
 * and the most complex function to its line. Ratios and sizes lead nowhere.
 */
function figureLink(metric, value, sha, details) {
    if (metric === 'complexity')
        return functionLink(value, sha, details);
    return findingLinks[metric]?.(sha);
}
/** What a change is measured from: the commit's time where it has one, else the inspection's. */
function sinceWhen(previous) {
    const time = previous ? commitTime(previous) : null;
    if (time !== null)
        return `the commit of ${dateTime.format(time)}`;
    const inspectedAt = Date.parse(previous?.generatedAt ?? '');
    return Number.isFinite(inspectedAt) ? `the inspection of ${dateTime.format(inspectedAt)}` : 'the previous inspection';
}
/** A figure shows its change only when it moved, and only against an inspection under the same profile. */
function leadFigures(selected, previous, area, history) {
    const read = (report, metric) => metricValue(report, metric, history.quality[report.sha], history.context?.[report.sha]);
    const comparable = previous !== undefined && history.context?.[selected.sha]?.boundary !== 'changed';
    return (areas[area] ?? areas.debt).map(metric => {
        const value = read(selected, metric);
        const prior = comparable ? read(previous, metric) : null;
        return { metric, value, change: value !== null && prior !== null ? difference(value, prior, metric) : null };
    });
}
function figureValue(figure, sha, details) {
    const shown = document.createElement('dd');
    shown.className = 'figure-value';
    const { metric, value } = figure;
    if (value === null) {
        shown.textContent = '—';
        return shown;
    }
    const destination = value > 0 ? figureLink(metric, value, sha, details) : undefined;
    if (!destination) {
        shown.textContent = formatMetric(value, metric);
        return shown;
    }
    const anchor = document.createElement('a');
    anchor.href = destination;
    anchor.textContent = formatMetric(value, metric);
    shown.append(anchor);
    return shown;
}
function figureDelta(figure, since) {
    const text = figure.change && figure.change.direction !== 'same' ? figure.change.text : '';
    const delta = document.createElement('dd');
    delta.className = 'figure-delta';
    delta.textContent = text;
    delta.dataset.direction = figure.change?.direction ?? 'none';
    if (text)
        delta.title = since;
    return delta;
}
/** The figures under the Overview's chart for the selected area, each with its change since the inspection before. */
export function renderLead(reports, selected, history, area, details) {
    const previous = reports[reports.indexOf(selected) + 1];
    const figures = leadFigures(selected, previous, area, history);
    // An area with nothing measured at this commit says so once, instead of a row of dashes.
    if (figures.every(entry => entry.value === null)) {
        const empty = document.createElement('p');
        empty.className = 'lead-empty';
        empty.textContent = emptyAreas[area] ?? 'Not measured for this commit.';
        byId('lead').replaceChildren(empty);
        return;
    }
    const since = previous ? `Since ${sinceWhen(previous)} (${previous.sha.slice(0, 7)})` : '';
    byId('lead').replaceChildren(...figures.map(figure => {
        const node = document.createElement('div');
        node.className = 'figure';
        const label = document.createElement('dt');
        label.textContent = labelOf(figure.metric);
        node.append(label, figureValue(figure, selected.sha, details), figureDelta(figure, since));
        return node;
    }));
}
