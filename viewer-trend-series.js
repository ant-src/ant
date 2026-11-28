import { renderLineChart } from './viewer-chart.js';
import { element, store, stored } from './viewer-dom.js';
import { formatMetric, isPercent, labelOf } from './viewer-metrics.js';
import { trendValue } from './viewer-trend-data.js';
import { seriesPoints, tickUnit } from './viewer-trend-points.js';
/** The measure drawn beside a series as a grey line on its own scale: the time behind the ratio, the lines behind the share. */
export const companions = { debt: 'effort', coverage: 'uncovered-lines' };
const hiddenKey = 'repometry.hidden-series';
let hiddenSeries;
/** The lines switched off in the key, remembered for this viewer. */
function hidden() {
    if (hiddenSeries)
        return hiddenSeries;
    try {
        hiddenSeries = new Set(JSON.parse(stored(() => localStorage, hiddenKey) ?? '[]'));
    }
    catch {
        hiddenSeries = new Set();
    }
    return hiddenSeries;
}
function remember() {
    store(() => localStorage, hiddenKey, JSON.stringify([...hidden()]));
}
/** The card under the pointer names the grey series too. */
function withCompanion(points, second, pair) {
    return points.map((entry, index) => {
        const value = second[index];
        if (value == null)
            return entry;
        return { ...entry, tip: { ...entry.tip, notes: [`${labelOf(pair)}: ${formatMetric(value, pair)}`, ...entry.tip.notes] } };
    });
}
/** One entry of the key: pressed while its line shows; the one line left stays, since a chart of nothing says nothing. */
function keyItem(series, kind, on, alone, toggle) {
    const item = document.createElement('button');
    item.type = 'button';
    item.className = `trend-key-item trend-key-${kind}`;
    item.textContent = labelOf(series);
    item.setAttribute('aria-pressed', String(on));
    item.disabled = on && alone;
    item.addEventListener('click', toggle);
    return item;
}
function emptySeries(chart) {
    chart.replaceChildren(element('p', 'detail empty', 'No measurements for this series in the shown inspections.'));
}
/**
 * A series with its companion, as the Overview draws it: the series in ink on the left scale, the companion in grey on
 * the right, and a key whose entries switch each line on and off. The last line shown cannot be switched off; with one
 * left it takes the left scale. The choice is remembered for this viewer. Returns whether anything was drawn.
 */
export function renderSeries(view, options) {
    const { chart, key, data, shown, metric, selected, link } = view;
    const pair = companions[metric];
    const second = pair ? shown.map(report => trendValue(data, report, pair)) : [];
    // A companion earns its line only once it can draw one, that is, from two measurements on.
    const paired = pair !== undefined && second.filter(item => item !== null).length >= 2;
    const main = seriesPoints(data, shown, metric, selected, link);
    const hasData = main.some(entry => entry.value !== null) || (paired && second.some(item => item !== null));
    function draw() {
        const showMain = !paired || !hidden().has(metric);
        const showSecond = paired && !hidden().has(pair);
        const both = showMain && showSecond;
        const lead = showMain ? metric : pair;
        const base = showMain ? main : seriesPoints(data, shown, pair, selected, link);
        const points = both ? withCompanion(base, second, pair) : base;
        chart.classList.toggle('chart-companion-only', !showMain);
        if (points.some(entry => entry.value !== null)) {
            renderLineChart(chart, points, { format: number => formatMetric(number, lead), ratio: isPercent(lead), compact: options.compact,
                height: options.height, tickUnit: tickUnit(points.map(entry => entry.value), lead), labelCurrent: options.labelCurrent ?? true,
                secondary: both ? { values: second, format: number => formatMetric(number, pair), tickUnit: tickUnit(second, pair) } : undefined });
        }
        else
            emptySeries(chart);
        key.hidden = !paired;
        if (!paired) {
            key.replaceChildren();
            return;
        }
        key.replaceChildren(...[[metric, 'main', showMain], [pair, 'second', showSecond]].map(([series, kind, on]) => keyItem(series, kind, on, !both, () => {
            if (on)
                hidden().add(series);
            else
                hidden().delete(series);
            remember();
            draw();
            key.querySelector(`.trend-key-${kind}`)?.focus();
        })));
    }
    if (hasData)
        draw();
    else {
        chart.replaceChildren();
        key.hidden = true;
    }
    return hasData;
}
