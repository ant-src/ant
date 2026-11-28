import { renderBarChart } from './viewer-chart.js';
import { element } from './viewer-dom.js';
import { splitter } from './viewer-layout.js';
import { definitionOf } from './viewer-metrics.js';
import { availableViews, charted, defaultPeriod, viewLabel, windowOf } from './viewer-trend-data.js';
import { periodControl, viewNavigation } from './viewer-trend-nav.js';
import { debtChanges, seriesPoints, signedDuration, tickUnit, trendLegend } from './viewer-trend-points.js';
import { renderSeries } from './viewer-trend-series.js';
const definitions = {
    'debt-changes': 'Change in estimated remediation time from the previous inspection.',
    'diff-coverage': 'Coverage of added lines in saved pull request reviews, against each review threshold.',
};
/** The series the page opens on: the one asked for, else technical debt, else the first this history has. */
function chosenView(views, requested) {
    if (requested && views.includes(requested))
        return requested;
    return views.includes('debt') ? 'debt' : views[0] ?? 'code-lines';
}
/** Remediation time as one bar per inspection's change; returns the legend under the chart. */
function drawDebtChanges({ data, selected, shown, frame, link }) {
    const bars = debtChanges(data, shown, selected.sha, link);
    const chart = element('div', 'chart');
    frame.append(chart);
    renderBarChart(chart, bars, { format: signedDuration, compact: data.compact, upIsWorse: true, tickUnit: tickUnit(bars.map(bar => bar.value), 'effort') });
    return trendLegend(bars);
}
/**
 * The same chart as the Overview: the series, its companion in grey, and a key that switches either line. The title
 * repeats the item chosen in the sidebar; the key under the chart names each line. Returns the legend under the chart.
 */
function drawSeries({ data, selected, shown, view, frame, link }) {
    const chart = element('div', 'chart');
    const key = element('p', 'trend-key');
    frame.append(chart, key);
    const metric = view;
    if (!renderSeries({ chart, key, data, shown, metric, selected: selected.sha, link }, { compact: data.compact })) {
        chart.append(element('p', 'detail empty', 'No measurements for this series in the shown inspections.'));
    }
    return trendLegend(seriesPoints(data, shown, metric, selected.sha, link));
}
/** Renders the whole page; returns the slot for the changed-line coverage history when that view is chosen. */
export function renderTrendsPage(container, data, selected, requested, requestedPeriod) {
    const views = availableViews(data);
    const view = chosenView(views, requested);
    const period = requestedPeriod ?? defaultPeriod(data.reports.length);
    const shown = charted(windowOf(data.reports, selected, period));
    const link = (sha) => data.link(sha, view, period);
    // The chart is the page; the series list is its sidebar, behind the same draggable separator as everywhere.
    const layout = element('div', 'trends-layout fx-split');
    const main = element('div', 'trend-main');
    layout.append(main, splitter(layout), viewNavigation(data, views, view, selected, period));
    const head = element('div', 'trend-head');
    const heading = element('div');
    heading.append(element('h2', 'trend-headline', viewLabel(view)));
    head.append(heading);
    // The title stands above the frame; the frame holds the period, the chart, its key, and what the series means.
    const frame = element('div', 'trend-frame panel');
    const tools = element('div', 'trend-frame-tools');
    tools.append(periodControl(data, selected.sha, view, period));
    frame.append(tools);
    main.append(head, frame);
    const page = { data, selected, shown, view, frame, link };
    const slot = view === 'diff-coverage' ? element('div', 'trend-slot') : null;
    if (slot)
        frame.append(slot);
    const legend = slot ? '' : view === 'debt-changes' ? drawDebtChanges(page) : drawSeries(page);
    frame.append(element('p', 'detail', definitions[view] ?? definitionOf(view)));
    if (legend)
        frame.append(element('p', 'detail legend', legend));
    container.replaceChildren(layout);
    return slot;
}
