import { basename, svgElement } from './viewer-dom.js';
/** The worst aspect ratio of a row of areas laid along a side: the row grows while this improves. */
function worst(areas, side) {
    const sum = areas.reduce((a, b) => a + b, 0);
    return Math.max(...areas.map(area => Math.max((side * side * area) / (sum * sum), (sum * sum) / (side * side * area))));
}
/** Lays a row along the shorter side of the box; returns the boxes and the room left beside them. */
function layRow(row, scale, box) {
    const sum = row.reduce((acc, item) => acc + item.area * scale, 0);
    const vertical = box.w >= box.h;
    const thickness = sum / (vertical ? box.h : box.w);
    let offset = 0;
    const placed = row.map(item => {
        const length = (item.area * scale) / thickness;
        const at = offset;
        offset += length;
        return { branch: item, box: vertical ? { x: box.x, y: box.y + at, w: thickness, h: length } : { x: box.x + at, y: box.y, w: length, h: thickness } };
    });
    const rest = vertical ? { x: box.x + thickness, y: box.y, w: box.w - thickness, h: box.h }
        : { x: box.x, y: box.y + thickness, w: box.w, h: box.h - thickness };
    return { placed, rest };
}
/**
 * Squarified treemap layout: rows of tiles kept as close to square as the areas allow, largest first, so sizes stay
 * comparable at a glance.
 */
export function squarify(children, room) {
    const items = [...children].sort((a, b) => b.area - a.area);
    const total = items.reduce((sum, item) => sum + item.area, 0);
    if (total <= 0 || room.w <= 0 || room.h <= 0)
        return [];
    const scale = (room.w * room.h) / total;
    const areas = (line) => line.map(item => item.area * scale);
    const placed = [];
    let box = room;
    let row = [];
    const lay = () => {
        const laid = layRow(row, scale, box);
        placed.push(...laid.placed);
        box = laid.rest;
    };
    for (const item of items) {
        const side = Math.min(box.w, box.h);
        if (row.length === 0 || worst(areas([...row, item]), side) <= worst(areas(row), side))
            row.push(item);
        else {
            lay();
            row = [item];
        }
    }
    if (row.length)
        lay();
    return placed;
}
/** A label cut to the room it has, at about six pixels a character. */
export function fit(text, room) {
    const chars = Math.floor(room / 6.2);
    if (text.length <= chars)
        return text;
    return chars > 2 ? `${text.slice(0, chars - 1)}…` : '';
}
/** A chain of folders that each hold one folder is one frame: src/main/ts/ rather than three nested ones. */
export function collapseChain(byParent, branch) {
    let inner = byParent.get(branch.entry.path) ?? [];
    let name = basename(branch.entry.path);
    while (inner.length === 1 && inner[0].folder) {
        name += `/${basename(inner[0].entry.path)}`;
        inner = byParent.get(inner[0].entry.path) ?? [];
    }
    return { name, inner };
}
/** The strip a folder's name takes at the top of its frame: a full one where it fits, a hairline where it does not. */
function stripHeight(open, box) {
    if (!open)
        return 0;
    return box.w > 64 && box.h > 44 ? 15 : 3;
}
function labelFits(folder, open, strip, box) {
    if (open && strip <= 3)
        return false;
    return box.w > (folder ? 64 : 44) && box.h > (open ? 0 : 17);
}
function tileLabel(text, box, shade) {
    const label = svgElement('text');
    label.classList.add('fx-tile-label');
    if (shade !== 'none' && shade >= 3)
        label.classList.add('fx-tile-label-light');
    label.setAttribute('x', (box.x + 4).toFixed(1));
    label.setAttribute('y', (box.y + 11.5).toFixed(1));
    label.textContent = fit(text, box.w - 8);
    return label;
}
function rectangle(box) {
    return `M${box.x.toFixed(1)},${box.y.toFixed(1)}h${box.w.toFixed(1)}v${box.h.toFixed(1)}h${(-box.w).toFixed(1)}z`;
}
/** The tree as tiles: folders frame their contents under a name strip, files fill the space by size. */
export function drawTiles(map, children, room, level) {
    for (const { branch, box } of squarify(children, room)) {
        if (box.w < 2 || box.h < 2 || map.drawn >= 600)
            continue;
        const tile = map.tile(branch, rectangle(box));
        if (branch.folder)
            tile.classList.add('fx-tile-folder');
        const { name, inner } = collapseChain(map.byParent, branch);
        const open = branch.folder && inner.length > 0 && level < map.depth && box.w > 28 && box.h > 28;
        const strip = stripHeight(open, box);
        if (labelFits(branch.folder, open, strip, box))
            tile.append(tileLabel(name + (branch.folder ? '/' : ''), box, map.shade(branch)));
        if (open)
            drawTiles(map, inner, { x: box.x + 3, y: box.y + strip, w: box.w - 6, h: box.h - strip - 3 }, level + 1);
    }
}
