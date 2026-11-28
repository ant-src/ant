import { valueLabel } from './viewer-chart-label.js';
import { flagMark, markBoundaries, measuredValues, pointLink, segments, stepPath } from './viewer-chart-points.js';
import { niceStep, yAxis } from './viewer-chart-scale.js';
import { line, setAttributes, svgNode } from './viewer-chart-svg.js';
import { placeX } from './viewer-chart-x.js';
/**
 * The second scale starts at zero and shares the main grid lines: each line carries one round value of each. Its top
 * is known once the main axis has settled how many lines there are.
 */
function secondScale(secondary) {
    const values = measuredValues(secondary.values);
    const scale = { high: 0, labels: undefined };
    if (!values.length)
        return scale;
    scale.labels = ticks => {
        const unit = secondary.tickUnit ?? 1;
        const step = niceStep(Math.max(...values, 0) / unit / Math.max(1, ticks - 1)) * unit;
        scale.high = step * (ticks - 1);
        return Array.from({ length: ticks }, (_, index) => index === 0 ? '0' : secondary.format(step * index));
    };
    return scale;
}
function drawSecondary(svg, layout, secondary, high, x) {
    const y = (value) => layout.top + layout.plotHeight - value / high * layout.plotHeight;
    for (const segment of segments(secondary.values)) {
        const path = svgNode('path', 'tc-line tc-line-secondary');
        path.setAttribute('d', stepPath(segment, index => secondary.values[index], x, y));
        svg.append(path);
    }
}
/**
 * A dot per commit is noise once commits sit close together: the line carries the curve, and a dot marks only what
 * says something (the selected and the last commit, flags, and a measurement with no neighbour to draw a line to).
 * Hover and the keyboard move one dot along the line instead. Only a short, sparse series keeps a dot per commit.
 */
function marked(points, index, sparse) {
    const point = points[index];
    if (sparse || point.current || index === points.length - 1)
        return true;
    if (point.flag || point.boundary)
        return true;
    return point.value !== null && points[index - 1]?.value == null && points[index + 1]?.value == null;
}
function pointMarks(point, cx, cy) {
    const dot = svgNode('circle', point.current ? 'tc-dot tc-dot-current' : 'tc-dot');
    if (point.partial)
        dot.classList.add('tc-dot-partial');
    setAttributes(dot, { cx, cy, r: point.current ? 5 : 3.5 });
    return point.flag ? [flagMark(point.flag, cx, cy), dot] : [dot];
}
/** Each marked point's dot, flag or gap mark, and the one value the chart labels. */
function drawMarks(svg, layout, points, { format, labelCurrent, y }) {
    const x = (index) => layout.xs[index];
    const axis = layout.top + layout.plotHeight;
    const sparse = points.length <= 12 && layout.xs.every((at, index) => index === 0 || at - layout.xs[index - 1] >= 24);
    // One value is labelled: the selected commit's, or the last one when the selection is off the chart.
    const selectedIndex = points.findIndex(point => point.current && point.value !== null);
    const labelled = labelCurrent && selectedIndex >= 0 ? selectedIndex : points.length - 1;
    const context = { points, x, y, format, axis };
    for (const [index, point] of points.entries()) {
        if (!marked(points, index, sparse))
            continue;
        const link = pointLink(point, index);
        if (point.value === null)
            link.append(line('tc-missing-mark', { x1: x(index), x2: x(index), y1: axis - 7, y2: axis }));
        else
            link.append(...pointMarks(point, x(index), y(point.value)));
        if (point.value !== null && index === labelled)
            link.append(valueLabel(context, index));
        svg.append(link);
    }
}
export function lineChart(svg, layout, points, { format, ratio, compact, tickUnit = 1, labelCurrent = true, secondary }) {
    const second = secondary && secondScale(secondary);
    const y = yAxis(svg, layout, { values: measuredValues(points.map(point => point.value)), ratio, count: compact ? 3 : 4, format, tickUnit,
        zero: false, rightLabels: second?.labels });
    placeX(svg, layout, points);
    const x = (index) => layout.xs[index];
    markBoundaries(svg, layout, points);
    if (secondary && second.high > 0)
        drawSecondary(svg, layout, secondary, second.high, x);
    for (const segment of segments(points.map(point => point.value))) {
        const path = svgNode('path', 'tc-line');
        path.setAttribute('d', stepPath(segment, index => points[index].value, x, y));
        svg.append(path);
    }
    layout.ys = points.map(point => point.value === null ? null : y(point.value));
    drawMarks(svg, layout, points, { format, labelCurrent, y });
}
