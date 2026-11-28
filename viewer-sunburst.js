import { count, svgElement } from './viewer-dom.js';
import { chartCenter, innerRadius, outerRadius, sectorPath } from './viewer-map-shape.js';
/** Wedges thinner than this angle, or past this many, are left out: they could be neither seen nor pointed at. */
const thinnest = 0.018;
const mostWedges = 144;
/** The tree as rings: each level one ring further out, each wedge as wide as its share of its parent's lines. */
export function drawRings(map, children, first, last, level) {
    if (level >= map.depth)
        return;
    const ring = (outerRadius - innerRadius) / map.depth;
    const total = children.reduce((sum, branch) => sum + branch.area, 0);
    const scale = (last - first) / total;
    let start = first;
    for (const branch of children) {
        const end = start + branch.area * scale;
        if (end - start >= thinnest && map.drawn < mostWedges) {
            const radius = innerRadius + level * ring;
            map.tile(branch, sectorPath(radius, radius + ring, start, end));
            drawRings(map, map.byParent.get(branch.entry.path) ?? [], start, end, level + 1);
        }
        start = end;
    }
}
/** The ring's hole holds the total the map is sized by. */
export function centerLabel(total, units) {
    const center = svgElement('g');
    center.setAttribute('aria-hidden', 'true');
    const value = svgElement('text');
    value.setAttribute('x', String(chartCenter));
    value.setAttribute('y', String(chartCenter + 1));
    value.classList.add('fx-sunburst-value');
    value.textContent = count(total);
    const label = svgElement('text');
    label.setAttribute('x', String(chartCenter));
    label.setAttribute('y', String(chartCenter + 21));
    label.classList.add('fx-sunburst-unit');
    label.textContent = units;
    center.append(value, label);
    return center;
}
