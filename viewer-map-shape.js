import { segmented } from './viewer-controls.js';
import { element, store, stored, svgElement } from './viewer-dom.js';
export const chartCenter = 200;
export const innerRadius = 47;
export const outerRadius = 196;
export const fullCircle = 2 * Math.PI;
export const startAngle = -Math.PI / 2;
const shapeKey = 'repometry.map';
export function mapShape() {
    return stored(() => localStorage, shapeKey) === 'tiles' ? 'tiles' : 'rings';
}
/** The switch between the two shapes, remembering the choice before the map redraws. */
export function shapeSwitch(shape, change) {
    return segmented('Map shape', [['rings', 'Sunburst'], ['tiles', 'Treemap']], shape, value => {
        store(() => localStorage, shapeKey, value);
        change(value);
    });
}
function stubCanvas(shape) {
    const canvas = svgElement('svg');
    canvas.classList.add('fx-sunburst-chart');
    canvas.setAttribute('aria-hidden', 'true');
    if (shape === 'tiles') {
        canvas.classList.add('fx-treemap');
        canvas.setAttribute('viewBox', '0 0 400 320');
        const tile = svgElement('rect');
        tile.setAttribute('width', '400');
        tile.setAttribute('height', '320');
        tile.setAttribute('rx', '3');
        canvas.append(tile);
        return canvas;
    }
    canvas.setAttribute('viewBox', '0 0 400 400');
    const ring = svgElement('circle');
    ring.setAttribute('cx', '200');
    ring.setAttribute('cy', '200');
    ring.setAttribute('r', String((outerRadius + innerRadius) / 2));
    ring.setAttribute('stroke-width', String(outerRadius - innerRadius));
    canvas.append(ring);
    return canvas;
}
/**
 * The compact map's stand-in while the inspection loads: the same frame, switch and canvas size in a flat tone, so the
 * column keeps its shape and the map fades in where it will stand.
 */
export function mapStub() {
    const shape = mapShape();
    const panel = element('section', 'fx-sunburst fx-sunburst-compact fx-map-stub');
    panel.dataset.shape = shape;
    panel.setAttribute('aria-busy', 'true');
    const heading = element('div', 'fx-map-heading');
    heading.append(shapeSwitch(shape, () => panel.replaceWith(mapStub())));
    panel.append(heading, stubCanvas(shape), element('p', 'fx-map-summary fx-map-hint fx-map-hint-quiet'));
    return panel;
}
/** A ring sector between two angles; the whole circle is drawn as two halves, since one arc cannot close on itself. */
export function sectorPath(inner, outer, first, last) {
    const large = last - first > Math.PI ? 1 : 0;
    const radius = (r, angle) => `${(chartCenter + r * Math.cos(angle)).toFixed(3)},${(chartCenter + r * Math.sin(angle)).toFixed(3)}`;
    if (last - first >= fullCircle - 0.00001) {
        const middle = first + Math.PI;
        return `M${radius(outer, first)} A${outer},${outer} 0 1 1 ${radius(outer, middle)} A${outer},${outer} 0 1 1 ${radius(outer, first)}
      M${radius(inner, first)} A${inner},${inner} 0 1 0 ${radius(inner, middle)} A${inner},${inner} 0 1 0 ${radius(inner, first)} Z`;
    }
    return `M${radius(outer, first)} A${outer},${outer} 0 ${large} 1 ${radius(outer, last)}
    L${radius(inner, last)} A${inner},${inner} 0 ${large} 0 ${radius(inner, first)} Z`;
}
/**
 * The readout under the map, filled on hover or focus: values appear only there, so the map does not repeat the
 * ranking beside it. The pointed place keeps its colour together with its parents and its contents; the rest fades.
 */
export function readout(canvas, summary, words) {
    // One outline drawn above every wedge, so neighbours never paint over it.
    const outline = svgElement('path');
    outline.classList.add('fx-wedge-outline');
    outline.setAttribute('aria-hidden', 'true');
    const reset = () => {
        canvas.removeAttribute('data-inspecting');
        for (const wedge of canvas.querySelectorAll('.fx-wedge[data-related]'))
            wedge.removeAttribute('data-related');
        outline.remove();
        summary.classList.add('fx-map-hint');
        summary.textContent = '';
    };
    canvas.addEventListener('pointerleave', reset);
    canvas.addEventListener('focusout', event => {
        if (!canvas.contains(event.relatedTarget))
            reset();
    });
    return (branch, tile) => {
        const own = branch.entry.path;
        canvas.setAttribute('data-inspecting', '');
        for (const wedge of canvas.querySelectorAll('.fx-wedge')) {
            const other = wedge.dataset.path;
            wedge.toggleAttribute('data-related', other === own || own.startsWith(`${other}/`) || other.startsWith(`${own}/`));
        }
        outline.setAttribute('d', tile.querySelector('path').getAttribute('d'));
        canvas.append(outline);
        summary.classList.remove('fx-map-hint');
        summary.replaceChildren(element('code', 'fx-path', own), element('strong', '', words(branch)));
    };
}
