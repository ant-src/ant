import { collator, count, element, inside, parentPath, svgElement } from './viewer-dom.js';
import { coverageText, duplicationText, hotspotScore } from './viewer-file-figures.js';
import { band, bandLimits, levelLegend, meaningOf, palette, relativePeaks, scoreLabel, spreadLevels, sunburstLegend } from './viewer-map-scale.js';
import { fullCircle, mapShape, readout, shapeSwitch, startAngle } from './viewer-map-shape.js';
import { securityRisks } from './viewer-metrics.js';
import { centerLabel, drawRings } from './viewer-sunburst.js';
import { unreadFolders } from './viewer-totals.js';
import { drawTiles } from './viewer-treemap.js';
/** The files and folders under the map's folder, by parent, sized by code lines where the folder has them. */
function mapTree(metrics, directory, metric, useCodeLines, levels) {
    const byParent = new Map();
    let measured = levels !== undefined;
    const add = (entry, folder) => {
        const area = useCodeLines ? entry.quality?.codeLines : entry.lines;
        if (!area || !inside(entry.path, directory))
            return;
        const value = hotspotScore(entry, metric);
        measured ||= value !== null;
        // Security levels cover every file; any other score is unknown where the file was not measured.
        const branch = { entry, folder, area, score: levels ? levels.get(entry.path) ?? 0 : value };
        const parent = parentPath(entry.path);
        const siblings = byParent.get(parent);
        if (siblings)
            siblings.push(branch);
        else
            byParent.set(parent, [branch]);
    };
    for (const folder of metrics.directories)
        if (folder.path)
            add(folder, true);
    for (const file of metrics.files)
        add(file, false);
    for (const group of byParent.values())
        group.sort((a, b) => collator.compare(a.entry.path, b.entry.path));
    return { byParent, measured };
}
/** How many levels the tree has below the map's folder. */
function treeDepth(byParent, directory) {
    const base = directory ? directory.split('/').length : 0;
    let depth = 1;
    for (const group of byParent.values())
        for (const branch of group)
            depth = Math.max(depth, branch.entry.path.split('/').length - base);
    return depth;
}
/**
 * How each place is shaded: security by its level, anything else by its band. A file no plugin read, or a folder of
 * nothing else, has no measure: grey. Scanners read every file, so the security map colours them all.
 */
function shading(metrics, tree, metric, limits, levels) {
    if (levels)
        return branch => branch.score;
    const unread = unreadFolders(metrics.files);
    const peakOf = relativePeaks([...tree.byParent.values()].flat().map(branch => ({ path: branch.entry.path, folder: branch.folder, score: branch.score })));
    const grey = (branch) => branch.score === null || branch.entry.analysed === false || unread.has(branch.entry.path);
    return branch => grey(branch) ? 'none' : band(branch.entry, metric, branch.score, peakOf(branch.entry.path, branch.folder), limits);
}
/** The scale runs the panel's width under the title row: the key to the whole map. */
function mapLegend(metric, limits, levels, grey) {
    if (grey)
        return element('p', 'fx-note', 'Nothing measured for this map, so it shows sizes only.');
    return levels ? levelLegend() : sunburstLegend(metric, limits);
}
const canvasSizes = {
    rings: { compact: [400, 400], full: [400, 400] },
    tiles: { compact: [400, 320], full: [640, 420] },
};
/** What a place's readout adds to its score: the share the ranking counts. */
function readingContext(metric, entry) {
    if (metric === 'missedLines')
        return ` · ${coverageText(entry)} covered`;
    return metric === 'duplication' ? ` · ${duplicationText(entry)} duplicated` : '';
}
function mapHeading(options, shape) {
    const heading = element('div', 'fx-map-heading');
    if (options.reshape)
        heading.append(shapeSwitch(shape, options.reshape));
    if (options.controls)
        heading.append(options.controls);
    return heading;
}
/**
 * The map's panel. The view kind heads the map, at the left; what it is coloured by sits at the right. A title would
 * only repeat what the switch and the map already say, so the panel is named for screen readers alone. The compact map
 * has no title of its own: the switch alone heads it.
 */
function mapPanel(options, shape, compact) {
    const panel = element('section', compact ? 'fx-sunburst fx-sunburst-compact' : 'panel fx-sunburst');
    panel.dataset.shape = shape;
    if (!compact)
        panel.setAttribute('aria-label', 'Code map');
    const heading = mapHeading(options, shape);
    if (heading.childElementCount)
        panel.append(heading);
    return panel;
}
function mapCanvas(shape, compact, label) {
    const canvas = svgElement('svg');
    canvas.classList.add('fx-sunburst-chart');
    const [width, height] = canvasSizes[shape][compact ? 'compact' : 'full'];
    canvas.setAttribute('viewBox', `0 0 ${width} ${height}`);
    if (shape === 'tiles')
        canvas.classList.add('fx-treemap');
    canvas.setAttribute('role', 'group');
    canvas.setAttribute('aria-label', label);
    return { canvas, width, height };
}
/** What the layouts draw with: each place a link to its folder's map or its file, named with its size and reading. */
function mapDraw(tree, depth, wedges) {
    let drawn = 0;
    return {
        byParent: tree.byParent, depth, shade: wedges.shade,
        get drawn() { return drawn; },
        tile(branch, outline) {
            drawn++;
            const path = branch.entry.path;
            const tile = svgElement('a');
            tile.classList.add('fx-wedge');
            tile.setAttribute('href', branch.folder ? wedges.href('hotspots', { directory: path })
                : wedges.href('metrics', { directory: parentPath(path), file: path }));
            tile.setAttribute('aria-label', `${path} · ${count(branch.area)} ${wedges.units} · ${wedges.reading(branch)}`);
            tile.dataset.path = path;
            const area = svgElement('path');
            area.setAttribute('d', outline);
            area.setAttribute('data-band', String(wedges.shade(branch)));
            tile.append(area);
            tile.addEventListener('focus', () => wedges.inspect(branch, tile));
            tile.addEventListener('pointerenter', () => wedges.inspect(branch, tile));
            wedges.canvas.append(tile);
            return tile;
        },
    };
}
/** The code map of a folder: its files sized by lines and coloured by the metric, as rings or tiles. */
export function renderSunburst(details, options) {
    const { metrics } = details;
    const { metric, directory = '', compact = false } = options;
    const limits = bandLimits(details, metric);
    const shape = options.shape ?? mapShape();
    const panel = mapPanel(options, shape, compact);
    const useCodeLines = metrics.directories.find(entry => entry.path === directory)?.quality?.codeLines !== undefined;
    const levels = options.levels && spreadLevels(options.levels);
    const tree = mapTree(metrics, directory, metric, useCodeLines, levels);
    const root = tree.byParent.get(directory);
    if (!root?.length) {
        panel.append(element('p', 'fx-note', 'No measured lines to show for this folder.'));
        return panel;
    }
    const units = useCodeLines ? 'code lines' : 'physical lines';
    const security = levels !== undefined;
    // Grey marks no data: a metric this inspection did not measure at all (no coverage report) is not a map of zeros.
    const grey = options.muted === true || !tree.measured;
    panel.dataset.palette = palette(metric, security, grey);
    const reading = (branch) => security ? securityRisks[branch.score]
        : branch.score === null ? 'Not measured' : `${count(branch.score)} ${scoreLabel(metric)}`;
    // What the map encodes is said to assistive technology; on screen the legend's scale carries it.
    const meaning = security ? 'Colour shows the worst security finding.' : meaningOf(metric);
    const description = `${shape === 'tiles' ? 'Tile' : 'Wedge'} size shows ${units}. ${meaning}`;
    if (!compact)
        panel.append(mapLegend(metric, limits, security, grey));
    const { canvas, width, height } = mapCanvas(shape, compact, `Code map of ${directory || 'the repository root'}. ${description}`);
    const summary = element('p', 'fx-map-summary fx-map-hint fx-map-hint-quiet');
    const map = mapDraw(tree, Math.min(treeDepth(tree.byParent, directory), compact ? 4 : 5), {
        canvas, units, reading, href: options.href, shade: shading(metrics, tree, metric, limits, security),
        inspect: readout(canvas, summary, branch => reading(branch) + readingContext(metric, branch.entry)),
    });
    if (shape === 'tiles')
        drawTiles(map, root, { x: 0, y: 0, w: width, h: height }, 0);
    else
        drawRings(map, root, startAngle, startAngle + fullCircle, 0);
    // The compact map sits beside the scope line, which already states the total; tiles have no centre to hold it.
    if (!compact && shape === 'rings')
        canvas.append(centerLabel(root.reduce((sum, branch) => sum + branch.area, 0), units));
    panel.append(canvas, summary);
    return panel;
}
