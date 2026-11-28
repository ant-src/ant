import { text } from './viewer-chart-svg.js';
/** The x of every flag whose pole and pennant a label between `left` and `anchor` at `top` would cover. */
function flagsCrossed({ points, x, y }, left, anchor, top) {
    return points.flatMap((other, at) => {
        if (!other.flag || other.value === null)
            return [];
        const foot = y(other.value) - 4;
        const across = x(at) >= left && x(at) <= anchor + 11;
        return across && top > foot - 22 && top - 14 < foot ? [x(at)] : [];
    });
}
/** The label runs left from its anchor; it steps left of every flag it would cross, so it never covers one. */
function clearOfFlags(context, start, top, width) {
    let anchor = start;
    let poles = [];
    for (let pass = 0; pass < 4; pass += 1) {
        const crossed = flagsCrossed(context, anchor - width - 4, anchor, top);
        if (!crossed.length || Math.min(...crossed) - 7 >= anchor)
            break;
        poles = crossed;
        anchor = Math.min(...crossed) - 7;
    }
    return { anchor, poles };
}
/**
 * The value of the labelled point. A flag on this point or one just beside it takes the space above, so the value
 * moves left of the pole; a label pushed far from its point would sit on the line, so it goes under the point instead,
 * where a flag never is.
 */
export function valueLabel(context, index) {
    const { points, x, y, format, axis } = context;
    const point = points[index];
    const value = point.value;
    const last = index === points.length - 1;
    const top = y(value) - 11;
    const cleared = clearOfFlags(context, x(index), top, format(value).length * 7.2);
    const below = x(index) - cleared.anchor > 40 && y(value) + 22 <= axis;
    const anchor = below ? x(index) + (last ? 4 : 0) : cleared.anchor;
    const label = text(format(value), point.current ? 'tc-value tc-value-current' : 'tc-value', anchor, below ? y(value) + 21 : top);
    const flush = (cleared.poles.length > 0 && !below) || (last && points.length > 1);
    label.setAttribute('text-anchor', flush ? 'end' : 'middle');
    return label;
}
