import { betterOf, difference, formatMetric, pluginSpec } from './viewer-metrics.js';
import { commitTime, trendValue } from './viewer-trend-data.js';
const dateTime = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });
const shortDate = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' });
/** When the commit was made, or else when it was inspected. */
function whenOf(report) {
    const time = commitTime(report);
    if (time !== null)
        return dateTime.format(time);
    const inspectedAt = Date.parse(report.generatedAt);
    return Number.isFinite(inspectedAt) ? `Inspected ${dateTime.format(inspectedAt)}` : 'Date unavailable';
}
function pointNotes(report, boundary) {
    const notes = report.status === 'partial' ? ['Partial inspection'] : [];
    if (boundary)
        notes.push(boundary === 'changed' ? 'Inspection profile changed' : 'Profile comparison unavailable');
    return notes;
}
function point(report, reading, selected, link) {
    const { value, shown, change, boundary } = reading;
    const time = commitTime(report);
    const at = time ?? Date.parse(report.generatedAt);
    const when = whenOf(report);
    const moved = change && change.direction !== 'same' ? `, ${change.text}` : '';
    return {
        value, time, boundary,
        label: Number.isFinite(at) ? shortDate.format(at) : report.sha.slice(0, 7),
        title: `${report.sha.slice(0, 7)}, ${when}: ${shown}${moved}`,
        tip: { heading: when, commit: report.sha.slice(0, 7), ...(report.subject ? { subject: report.subject } : {}), value: shown, change,
            notes: pointNotes(report, boundary) },
        href: link(report.sha),
        current: report.sha === selected,
        partial: report.status === 'partial',
    };
}
/** The inspection just before this one in the whole history, which the change is measured from. */
export function previousOf(data, report) {
    return data.reports[data.reports.indexOf(report) + 1];
}
/** A value's change from the inspection before; none across a profile change or a missing measurement. */
function readingOf(data, report, metric) {
    const value = trendValue(data, report, metric);
    const previous = previousOf(data, report);
    const before = previous ? trendValue(data, previous, metric) : null;
    const boundary = data.context?.[report.sha]?.boundary ?? null;
    const comparable = boundary !== 'changed' && value !== null && before !== null;
    return { value, boundary, shown: value === null ? 'No measurement' : formatMetric(value, metric),
        change: comparable ? difference(value, before, metric) : null };
}
export function seriesPoints(data, shown, metric, selected, link) {
    const points = shown.map(report => point(report, readingOf(data, report, metric), selected, link));
    const { best, worst } = plantFlags(points.map(entry => entry.value), metric);
    return points.map((entry, index) => {
        if (index === best)
            return { ...entry, flag: 'best' };
        return index === worst ? { ...entry, flag: 'worst' } : entry;
    });
}
/**
 * Where the green and the red flag stand: the best and the worst value of the series, each at the inspection that
 * first reached it, the way a flag is planted where a summit was reached. Neutral metrics get no flags, and neither
 * does a series with fewer than three measurements or one that reads as unchanged from its best to its worst.
 *
 * The extremes are the raw minimum and maximum, as the line draws them: a tolerant "beats" is not transitive, and a
 * first value within rounding of both ends used to hide them. A value reaches an extreme when it equals it up to
 * floating-point noise, so a flag never stands beside a point the line shows lower or higher.
 */
export function plantFlags(values, metric) {
    const none = { best: null, worst: null };
    const better = betterOf(metric);
    if (better === 'none')
        return none;
    const measured = values.flatMap((entry, index) => entry === null ? [] : [index]);
    if (measured.length < 3)
        return none;
    const raw = measured.map(index => values[index]);
    const lowest = Math.min(...raw);
    const highest = Math.max(...raw);
    const [bestValue, worstValue] = better === 'lower' ? [lowest, highest] : [highest, lowest];
    if (difference(bestValue, worstValue, metric).direction === 'same')
        return none;
    const reached = (target) => measured.find(index => Math.abs(values[index] - target) <= 1e-9 * Math.max(1, Math.abs(target)));
    const best = reached(bestValue);
    const worst = reached(worstValue);
    return best === worst ? none : { best, worst };
}
/** Only what the chart cannot say by itself; empty when there is nothing to add. */
export function trendLegend(points) {
    const parts = [
        points.some(entry => entry.partial && entry.value !== null) ? 'Dashed dots mark partial inspections' : '',
        points.some(entry => entry.boundary === 'changed') ? 'Dashed lines mark profile changes' : '',
        points.some(entry => entry.boundary === 'unavailable') ? 'Dotted lines mark missing profile information' : '',
    ].filter(Boolean);
    return parts.length ? `${parts.join('. ')}.` : '';
}
/** Remediation time ticks in whole hours once the series reaches two hours. */
export function tickUnit(values, metric) {
    const minutes = metric === 'effort' || pluginSpec(metric)?.descriptor.unit === 'minutes';
    return minutes && values.some(value => Math.abs(value ?? 0) >= 120) ? 60 : 1;
}
export function signedDuration(minutes) {
    if (Math.round(minutes) === 0)
        return '0 min';
    return `${minutes > 0 ? '+' : '−'}${formatMetric(Math.abs(minutes), 'effort')}`;
}
/** Why an inspection has no change to show: it is the first, the profile changed, or the one before was not measured. */
function noChange(previous, boundary) {
    if (!previous)
        return 'First inspection';
    return boundary === 'changed' ? 'Profile changed' : 'Not measured before';
}
/**
 * Remediation time as changes, one bar per inspection. The first shown inspection compares with the one just before
 * the window, when there is one. Flags mark the best and worst total; the bars show changes, so they carry none.
 */
export function debtChanges(data, shown, selected, link) {
    return seriesPoints(data, shown, 'effort', selected, link).map((entry, index) => {
        const report = shown[index];
        const previous = previousOf(data, report);
        const before = previous ? trendValue(data, previous, 'effort') : null;
        const change = entry.boundary !== 'changed' && entry.value !== null && before !== null ? entry.value - before : null;
        const total = entry.value === null ? 'No measurement' : `${formatMetric(entry.value, 'effort')} in total`;
        const shownChange = change === null ? noChange(previous, entry.boundary) : signedDuration(change);
        return { ...entry, value: change, flag: undefined,
            title: `${report.sha.slice(0, 7)}, ${entry.tip.heading}: ${shownChange}`,
            tip: { ...entry.tip, value: shownChange, change: null, notes: [total, ...entry.tip.notes] } };
    });
}
