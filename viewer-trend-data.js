import { isMetric, labelOf, metricLabels, metricValue, pluginMetrics } from './viewer-metrics.js';
/** The views the viewer draws by hand rather than from one metric's values. */
const ownViews = new Set(['debt-changes', 'diff-coverage']);
const ownLabels = {
    debt: 'Technical debt', 'debt-changes': 'Technical debt changes', 'diff-coverage': 'Changed-line coverage',
};
export function viewLabel(view) {
    return ownLabels[view] ?? labelOf(view);
}
/** The metric a view reads at a commit; the views drawn by hand read none. */
export function viewMetric(view) {
    return ownViews.has(view) ? null : view;
}
const builtinGroups = [
    ['Maintainability', ['debt', 'debt-changes', 'duplication', 'cognitive', 'complexity', 'smells']],
    // Branch and function coverage come from the registry, where the coverage report has them.
    ['Coverage', ['coverage', 'branch-coverage', 'function-coverage', 'diff-coverage']],
    ['Security', ['security']],
    ['Size', ['code-lines', 'files', 'functions', 'classes', 'comments']],
];
/** A view the registry names, not one the viewer draws by hand. */
function pluginKey(view) {
    return !ownViews.has(view) && !(view in metricLabels);
}
/**
 * The series list: Repometry's own groups, with other plugins' metrics joining a group they name or one of their own.
 * A registry metric the groups already place keeps its place, and is left out where this history has no such metric.
 */
export function groups() {
    const placed = new Set(builtinGroups.flatMap(([, views]) => views));
    const trended = pluginMetrics().filter(spec => spec.trend);
    const keys = new Set(trended.map(spec => spec.key));
    const merged = new Map(builtinGroups.map(([group, views]) => [group, views.filter(view => !pluginKey(view) || keys.has(view))]));
    for (const spec of trended.filter(spec => !placed.has(spec.key))) {
        merged.set(spec.group, [...merged.get(spec.group) ?? [], spec.key]);
    }
    return [...merged];
}
export function isTrendView(value) {
    if (value === 'debt-changes' || value === 'diff-coverage')
        return true;
    return isMetric(value) && groups().some(([, views]) => views.includes(value));
}
export function isPeriod(value) {
    return value === '30' || value === '100' || value === 'all';
}
/** A metric's value at a commit, read with the commit's own quality summary and context. */
export function trendValue(data, report, metric) {
    return metricValue(report, metric, data.quality[report.sha], data.context?.[report.sha]);
}
/**
 * Views that have data in this history, in display order: debt changes need remediation times, and changed-line
 * coverage needs saved pull request reviews.
 */
export function availableViews(data) {
    const has = (metric) => data.reports.some(report => trendValue(data, report, metric) !== null);
    const shown = (view) => view === 'diff-coverage' ? data.reviews > 0 : has(view === 'debt-changes' ? 'effort' : view);
    return groups().flatMap(([, views]) => views).filter(shown);
}
/** The whole curve by default; only a very long history opens on its last hundred inspections. */
export function defaultPeriod(total) {
    return total <= 1000 ? 'all' : '100';
}
const windows = new Map();
/** The inspections shown for a period, oldest first, always including the selected one. */
export function windowOf(reports, selected, period) {
    const limit = period === 'all' ? reports.length : Number(period);
    const { start, end } = stableWindow(reports, reports.indexOf(selected), limit, 'trends');
    return reports.slice(start, end).reverse();
}
/** Whether the slice shown last still fits: the same size over the same history, and the commit inside it. */
function keeps(previous, index, shape) {
    if (!previous || previous.size !== shape.size || previous.total !== shape.total)
        return false;
    return previous.head === shape.head && index >= previous.start && index < previous.end;
}
/**
 * The slice of history a chart shows, newest first by index. Picking a commit inside the current slice keeps it, so
 * the axes and scale stay where the eye left them; only a commit outside it moves the slice, which then centres on
 * that commit so its neighbours on both sides are in view. The newest commits are the default slice.
 */
export function stableWindow(reports, index, size, key) {
    const shape = { size, total: reports.length, head: reports[0]?.sha };
    const previous = windows.get(key);
    let result;
    if (keeps(previous, index, shape))
        result = { start: previous.start, end: previous.end };
    else if (index < size)
        result = { start: 0, end: Math.min(size, shape.total) };
    else {
        const end = Math.min(shape.total, Math.max(0, index - Math.floor(size / 2)) + size);
        result = { start: Math.max(0, end - size), end };
    }
    windows.set(key, { ...result, ...shape });
    return result;
}
/** Commit time of an inspection, or null for exports made before commit times were recorded. */
export function commitTime(report) {
    const time = report.committedAt === undefined ? NaN : Date.parse(report.committedAt);
    return Number.isFinite(time) ? time : null;
}
/**
 * The inspections a time chart can place: when the history has commit times, one without (a root commit has no churn
 * window to read it from) cannot be put on the time axis, so it stays out of the chart and in the table.
 */
export function charted(shown) {
    const timed = shown.filter(report => commitTime(report) !== null);
    return timed.length ? timed : [...shown];
}
