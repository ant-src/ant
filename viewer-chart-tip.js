import { flagIcon } from './viewer-chart-points.js';
import { setAttributes, svgNode } from './viewer-chart-svg.js';
const flagNotes = { best: 'Best of the shown inspections', worst: 'Worst of the shown inspections' };
/** The card scrolls, so a cluster lists its commits whole up to a bound a hand can still scroll through. */
const listed = 50;
function span(className, content) {
    const node = document.createElement('span');
    node.className = className;
    node.textContent = content;
    return node;
}
/** The value, and its change where it says something: in a cluster an unchanged value goes without saying. */
function valueLine(point, brief) {
    const value = span('tc-tip-value', point.tip.value);
    const change = point.tip.change;
    if (!change || (brief && change.direction === 'same'))
        return value;
    const node = span('tc-tip-change', change.text);
    node.dataset.direction = change.direction;
    value.append(' ', node);
    return value;
}
/** Each card opens its commit: in a cluster the card is how one of several commits at a point is picked. */
function card(point, brief) {
    const entry = document.createElement(point.href ? 'a' : 'div');
    entry.className = 'tc-tip-entry';
    if (entry instanceof HTMLAnchorElement)
        entry.href = point.href;
    const heading = document.createElement('strong');
    heading.textContent = point.tip.heading;
    const commit = document.createElement('code');
    commit.textContent = point.tip.commit;
    entry.append(heading, commit);
    if (point.tip.subject)
        entry.append(span('tc-tip-subject', point.tip.subject));
    entry.append(valueLine(point, brief));
    if (point.flag) {
        const flag = document.createElement('span');
        flag.className = 'tc-tip-flag';
        flag.append(flagIcon(point.flag), flagNotes[point.flag]);
        entry.append(flag);
    }
    entry.append(...point.tip.notes.map(note => span('tc-tip-note', note)));
    return entry;
}
/** Marks closer than a few pixels overlap, so the card lists every commit under the pointer, newest first. */
function fillTip(tip, points, layout, x) {
    const cluster = points.flatMap((_, other) => Math.abs(layout.xs[other] - x) < 4 ? [other] : []).reverse();
    tip.replaceChildren(...cluster.slice(0, listed).map(member => card(points[member], cluster.length > 1)));
    if (cluster.length > listed)
        tip.append(span('tc-tip-note', `${cluster.length - listed} more commits at this point`));
    tip.scrollTop = 0;
    // Only a card of several commits takes the pointer; a single card stays out of the way of scanning along the chart.
    tip.classList.toggle('tc-tip-held', cluster.length > 1);
    tip.hidden = false;
}
/** The card keeps the side it opened on while it fits, so it does not jump across the pointer heading for it. */
function placeTip(tip, svg, layout, mark, side) {
    const { x, y: markY } = mark;
    const scale = svg.getBoundingClientRect().width / Math.max(1, layout.width) || 1;
    const width = tip.offsetWidth;
    const room = svg.clientWidth || layout.width;
    const right = x * scale + 14;
    const leftSide = x * scale - 14 - width;
    const keepsLeft = side === 'left' && leftSide >= 0;
    const next = !keepsLeft && right + width <= room ? 'right' : 'left';
    tip.style.left = `${Math.max(0, next === 'right' ? right : leftSide)}px`;
    const bottom = layout.top * scale + layout.plotHeight * scale - tip.offsetHeight;
    tip.style.top = `${Math.max(0, Math.min((markY ?? layout.top + layout.plotHeight) * scale - tip.offsetHeight / 2, bottom))}px`;
    return next;
}
/** Marks the point as active, and puts the hover dot on it where the point draws no dot of its own. */
function highlight(svg, pointer, layout, index) {
    svg.querySelectorAll('.tc-point.is-active').forEach(node => node.classList.remove('is-active'));
    const own = svg.querySelector(`.tc-point[data-index="${index}"]`);
    own?.classList.add('is-active');
    const markY = layout.ys[index];
    if (markY == null || own?.querySelector('.tc-dot') || svg.classList.contains('tc-bars')) {
        pointer.setAttribute('visibility', 'hidden');
        return;
    }
    setAttributes(pointer, { cx: layout.xs[index], cy: markY });
    pointer.setAttribute('visibility', 'visible');
}
/** The point nearest the pointer; the later point on ties, so the newest of a cluster is reachable. */
function nearestPoint(svg, layout, clientX) {
    const box = svg.getBoundingClientRect();
    const px = (clientX - box.left) * (layout.width / Math.max(1, box.width));
    let best = 0;
    for (const [index, position] of layout.xs.entries()) {
        if (Math.abs(position - px) <= Math.abs(layout.xs[best] - px))
            best = index;
    }
    return best;
}
/** Whether the pointer is on its way to a held card: moving towards it and not there yet. */
function headingForCard(tip, side, dx, clientX) {
    if (tip.hidden || !tip.classList.contains('tc-tip-held'))
        return false;
    const box = tip.getBoundingClientRect();
    return side === 'right' ? dx > 0 && clientX < box.left : dx < 0 && clientX > box.right;
}
const keySteps = new Map([
    ['ArrowLeft', (from) => Math.max(0, from - 1)],
    ['ArrowRight', (from, last) => Math.min(last, from + 1)],
    ['Home', () => 0],
    ['End', (_, last) => last],
]);
function guideLine(svg) {
    const guide = svgNode('line', 'tc-guide');
    guide.setAttribute('visibility', 'hidden');
    svg.insertBefore(guide, svg.firstChild);
    return guide;
}
/** One dot follows the pointer or the keyboard along the line, so the chart needs no element per commit. */
function hoverDot(svg) {
    const pointer = svgNode('circle', 'tc-dot tc-hover-dot');
    pointer.setAttribute('r', '4.5');
    pointer.setAttribute('visibility', 'hidden');
    svg.append(pointer);
    return pointer;
}
/** One hover card for the whole chart: the point nearest the pointer, or the focused one. */
export function interact(svg, tip, layout, points) {
    if (points.length === 0)
        return;
    const guide = guideLine(svg);
    const pointer = hoverDot(svg);
    const state = {
        active: -1, side: null, pending: undefined, lastX: 0
    };
    const show = (index) => {
        if (index === state.active)
            return;
        state.active = index;
        const x = layout.xs[index];
        setAttributes(guide, { x1: x, x2: x, y1: layout.top - 6, y2: layout.top + layout.plotHeight });
        guide.setAttribute('visibility', 'visible');
        highlight(svg, pointer, layout, index);
        fillTip(tip, points, layout, x);
        state.side = placeTip(tip, svg, layout, { x, y: layout.ys[index] }, state.side);
    };
    const hide = () => {
        clearTimeout(state.pending);
        state.side = null;
        state.active = -1;
        tip.hidden = true;
        guide.setAttribute('visibility', 'hidden');
        pointer.setAttribute('visibility', 'hidden');
        svg.querySelectorAll('.tc-point.is-active').forEach(node => node.classList.remove('is-active'));
    };
    // A pointer heading for the card may cross other commits on the way. While it moves towards the card, the switch
    // waits a moment; reaching the card keeps it, so its commits can be scrolled and picked.
    svg.addEventListener('pointermove', event => {
        const candidate = nearestPoint(svg, layout, event.clientX);
        const dx = event.clientX - state.lastX;
        state.lastX = event.clientX;
        clearTimeout(state.pending);
        if (candidate === state.active)
            return;
        if (headingForCard(tip, state.side, dx, event.clientX))
            state.pending = setTimeout(() => show(candidate), 220);
        else
            show(candidate);
    });
    svg.addEventListener('pointerleave', event => {
        if (!tip.contains(event.relatedTarget))
            hide();
    });
    tip.addEventListener('pointerenter', () => clearTimeout(state.pending));
    tip.addEventListener('pointerleave', event => {
        if (!svg.contains(event.relatedTarget))
            hide();
    });
    svg.addEventListener('click', event => {
        if (event.target.closest('.tc-point'))
            return;
        const href = points[nearestPoint(svg, layout, event.clientX)]?.href;
        if (href)
            location.assign(href);
    });
    svg.addEventListener('focusin', event => {
        const index = Number(event.target.closest('.tc-point')?.dataset.index);
        if (Number.isInteger(index))
            show(index);
    });
    svg.addEventListener('focusout', hide);
    // The chart itself takes focus: arrows step through commits, Home and End jump to the ends, Enter opens the commit.
    svg.tabIndex = 0;
    svg.setAttribute('aria-keyshortcuts', 'ArrowLeft ArrowRight Home End Enter');
    svg.setAttribute('aria-label', `${points.length} ${points.length === 1 ? 'commit' : 'commits'}; arrow keys move between them, Enter opens one`);
    svg.addEventListener('focus', () => {
        const current = points.findIndex(point => point.current);
        show(current >= 0 ? current : points.length - 1);
    });
    svg.addEventListener('keydown', event => {
        const last = points.length - 1;
        const step = keySteps.get(event.key);
        if (step) {
            event.preventDefault();
            const next = step(state.active < 0 ? last : state.active, last);
            state.active = -1;
            show(next);
        }
        else if (event.key === 'Enter' && state.active >= 0) {
            event.preventDefault();
            location.assign(points[state.active].href);
        }
    });
}
