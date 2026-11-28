import { line, setAttributes, svgNode } from './viewer-chart-svg.js';
/** The values the points measured, in order, without the gaps. */
export function measuredValues(values) {
    return values.flatMap(value => value === null ? [] : [value]);
}
/** Runs of consecutive measured points, broken only where a measurement is missing; a profile change is marked, not cut. */
export function segments(values) {
    const result = [];
    let run = [];
    for (const [index, value] of values.entries()) {
        if (value !== null) {
            run.push(index);
            continue;
        }
        if (run.length)
            result.push(run);
        run = [];
    }
    if (run.length)
        result.push(run);
    return result.filter(segment => segment.length > 1);
}
/**
 * Code keeps its state until the next commit, so the line holds each value and steps at the commit that changed it;
 * a slope would invent values for the days between commits. A run of equal values is one horizontal stroke, so ten
 * thousand quiet commits cost a single segment.
 */
export function stepPath(segment, valueAt, x, y) {
    return segment.map((index, order) => {
        const value = valueAt(index);
        if (!order)
            return `M${x(index).toFixed(1)},${y(value).toFixed(1)}`;
        const last = order === segment.length - 1;
        return value === valueAt(segment[order - 1]) && !last ? '' : `H${x(index).toFixed(1)}V${y(value).toFixed(1)}`;
    }).join('');
}
/** A dashed line where the profile changed between two inspections. */
export function markBoundaries(svg, layout, points) {
    for (const [index, point] of points.entries()) {
        if (index === 0 || !point.boundary)
            continue;
        const x = layout.xs[index];
        svg.append(line(`tc-profile-${point.boundary}`, { x1: x, x2: x, y1: layout.top, y2: layout.top + layout.plotHeight }));
    }
}
/** A tiny flag planted in the point: the pennant of the logo on a short pole, both in the flag's colour. */
export function flagMark(kind, x, y) {
    const group = svgNode('g', `tc-flag tc-flag-${kind}`);
    const foot = y - 4;
    const head = foot - 12;
    const pole = svgNode('rect');
    setAttributes(pole, { x: x - .75, y: head, width: 1.5, height: foot - head });
    const pennant = svgNode('path');
    pennant.setAttribute('d', `M${(x + .75).toFixed(1)},${head.toFixed(1)}L${(x + 10).toFixed(1)},${(head - .4).toFixed(1)}L${(x + .75).toFixed(1)},${(head + 5.5).toFixed(1)}Z`);
    group.append(pole, pennant);
    return group;
}
export function flagIcon(kind) {
    const icon = svgNode('svg', `tc-flag tc-flag-${kind}`);
    icon.setAttribute('viewBox', '0 -1 11 14');
    icon.setAttribute('aria-hidden', 'true');
    icon.append(flagMark(kind, 1, 16));
    return icon;
}
/** A point's link to its commit, which holds its marks. */
export function pointLink(point, index) {
    const link = svgNode('a', 'tc-point');
    link.setAttribute('href', point.href);
    link.setAttribute('aria-label', point.title);
    link.dataset.index = String(index);
    if (point.current)
        link.setAttribute('aria-current', 'true');
    return link;
}
