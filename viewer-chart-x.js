import { centredText, clampLabel, inset, line } from './viewer-chart-svg.js';
import { axisLabels, timeTicks } from './viewer-chart-time.js';
/** Equal periods take equal width; the ticks step by one calendar period chosen to fit the width. */
function placeTimed(svg, layout, times, from, span) {
    const axis = layout.top + layout.plotHeight;
    const low = Math.min(...times);
    const high = Math.max(...times);
    const at = (time) => high === low ? from + span / 2 : from + (time - low) / (high - low) * span;
    layout.xs = times.map(at);
    if (high === low)
        return;
    const ticks = timeTicks(low, high, span);
    for (const tick of ticks)
        svg.append(line('tc-xtick', { x1: at(tick.time), x2: at(tick.time), y1: axis, y2: axis + 4 }));
    for (const label of axisLabels(ticks.map(tick => ({ time: tick.time, x: at(tick.time) })), ticks[0]?.unit ?? 'day', layout)) {
        svg.append(centredText(label.label, 'tc-label', label.at, layout.height - 9));
    }
}
/** Whether an ordered point keeps its label: every so many points, and always the last, which the one before yields to. */
function labelled(index, every, last) {
    if (index === last)
        return true;
    return index % every === 0 && last - index >= every;
}
/** Without times the points stand evenly in order under their own labels, as many labels as fit. */
function placeOrdered(svg, layout, points, from, span) {
    const step = points.length > 1 ? span / (points.length - 1) : 0;
    layout.xs = points.map((_, index) => points.length > 1 ? from + index * step : from + span / 2);
    const labelWidth = Math.max(...points.map(point => point.label.length)) * 7.6 + 10;
    const every = step > 0 ? Math.max(1, Math.ceil(labelWidth / step)) : 1;
    for (const [index, point] of points.entries()) {
        if (!labelled(index, every, points.length - 1))
            continue;
        svg.append(centredText(point.label, 'tc-label', clampLabel(layout.xs[index], point.label, layout), layout.height - 9));
    }
}
/**
 * Positions points on the x axis and draws it. When every point has a commit time the axis is time: equal periods take
 * equal width, and the ticks step by one calendar period (a day, a week, a month, a quarter, a year) chosen to fit the
 * width. Without times the points stand evenly in order under their own labels.
 */
export function placeX(svg, layout, points) {
    const from = layout.left + inset;
    const span = Math.max(1, layout.plotWidth - inset * 2);
    const times = points.map(point => point.time);
    if (points.length > 0 && times.every((time) => time !== null))
        placeTimed(svg, layout, times, from, span);
    else
        placeOrdered(svg, layout, points, from, span);
}
