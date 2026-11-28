/** The chart's SVG building blocks and the plot's layout, shared by its axes, marks and hover card. */
export const plotTop = 34;
export const plotBottom = 30;
export const plotRight = 18;
/** Keeps the first and last marks off the plot edges. */
export const inset = 12;
const svgNamespace = 'http://www.w3.org/2000/svg';
export function svgNode(tag, className) {
    const node = document.createElementNS(svgNamespace, tag);
    if (className)
        node.setAttribute('class', className);
    return node;
}
/** Sets numeric attributes, rounded to a tenth of a pixel. */
export function setAttributes(node, attributes) {
    for (const [name, value] of Object.entries(attributes))
        node.setAttribute(name, String(Number(value.toFixed(1))));
}
export function text(content, className, x, y) {
    const node = svgNode('text', className);
    node.textContent = content;
    setAttributes(node, { x, y });
    return node;
}
/** A label centred on its anchor. */
export function centredText(content, className, x, y) {
    const node = text(content, className, x, y);
    node.setAttribute('text-anchor', 'middle');
    return node;
}
/** Where a centred label stands so that it stays inside the plot. */
export function clampLabel(x, label, { left, plotWidth }) {
    const half = label.length * 3.6 + 2;
    return Math.min(Math.max(x, left + half), left + plotWidth - half);
}
/** A line across the plot, at a height or at a point. */
export function line(className, attributes) {
    const node = svgNode('line', className);
    setAttributes(node, attributes);
    return node;
}
