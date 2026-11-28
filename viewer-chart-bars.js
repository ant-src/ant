import { markBoundaries, measuredValues, pointLink } from './viewer-chart-points.js';
import { yAxis } from './viewer-chart-scale.js';
import { centredText, line, setAttributes, svgNode } from './viewer-chart-svg.js';
import { placeX } from './viewer-chart-x.js';
/** One width for every bar, set by the two closest inspections: only a bar's height says how much changed. */
function barWidth(xs) {
    const gaps = xs.slice(1).map((at, index) => at - xs[index]);
    return Math.max(2, Math.min(28, gaps.length ? Math.min(...gaps) * .62 : 28));
}
/** The room beside a bar: the distance to the nearer neighbour. */
function room(x, count, index) {
    const before = index > 0 ? x(index) - x(index - 1) : Infinity;
    const after = index < count - 1 ? x(index + 1) - x(index) : Infinity;
    return Math.min(before, after);
}
/** A bar from zero to the value, labelled where the label fits or the commit is the selected one. */
function bar({ points, x, y, format, upIsWorse, width }, index, space) {
    const point = points[index];
    const value = point.value;
    const worse = upIsWorse ? value > 0 : value < 0;
    const rect = svgNode('rect', `tc-bar ${worse ? 'tc-bar-worse' : 'tc-bar-better'}${point.current ? ' tc-bar-current' : ''}`);
    setAttributes(rect, { x: x(index) - width / 2, y: Math.min(y(value), y(0)), width, height: Math.max(1, Math.abs(y(value) - y(0))) });
    if (!point.current && format(value).length * 7.6 >= space)
        return [rect];
    const at = value > 0 ? y(value) - 7 : y(value) + 16;
    return [rect, centredText(format(value), point.current ? 'tc-value tc-value-current' : 'tc-value', x(index), at)];
}
/** A point's marks: its bar, a flat mark for no change, or a tick on the zero line where nothing was measured. */
function barMarks(context, index) {
    const { points, x, y, width } = context;
    const value = points[index].value;
    const space = room(x, points.length, index);
    const cx = x(index);
    if (value === null)
        return [line('tc-missing-mark', { x1: cx, x2: cx, y1: y(0) - 3.5, y2: y(0) + 3.5 })];
    if (value === 0)
        return [line('tc-zero-mark', { x1: cx - width / 2, x2: cx + width / 2, y1: y(0), y2: y(0) })];
    return bar(context, index, space);
}
export function barChart(svg, layout, points, options) {
    svg.classList.add('tc-bars');
    const y = yAxis(svg, layout, { values: measuredValues(points.map(point => point.value)), ratio: false, count: options.compact ? 3 : 4,
        format: options.format, tickUnit: options.tickUnit ?? 1, zero: true });
    placeX(svg, layout, points);
    markBoundaries(svg, layout, points);
    const x = (index) => layout.xs[index];
    layout.ys = points.map(point => point.value === null ? null : y(point.value));
    // In a long history the zero line already says "no change" and "not measured"; only bars that moved are drawn.
    const dense = layout.xs.some((at, index) => index > 0 && at - layout.xs[index - 1] < 7);
    const context = { ...options, points, x, y, width: barWidth(layout.xs) };
    for (const [index, point] of points.entries()) {
        const link = pointLink(point, index);
        const quiet = !point.current && !point.value;
        if (!dense || !quiet)
            link.append(...barMarks(context, index));
        svg.append(link);
    }
}
