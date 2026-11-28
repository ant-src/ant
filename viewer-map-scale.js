import { element } from './viewer-dom.js';
import { bandLabels as pluginBandLabels, bandOf, bandScale } from './viewer-metric-format.js';
import { labelOf, pluginSpec } from './viewer-metrics.js';
import { ancestors } from './viewer-totals.js';
const bandMeaning = {
    complexity: 'Colour shows the most complex function.',
    cognitive: 'Colour shows cognitive complexity relative to the highest file, or folder at the same depth.',
    missedLines: 'Colour shows the share of uncovered lines.',
    duplication: 'Colour shows the share of duplicated code lines.',
    churn: 'Colour shows Git changes relative to the busiest file, or folder at the same depth.',
};
/** The band limits used when the profile sets none. */
const defaultLimits = { complexity: [5, 10, 20, 50], missedLines: [.05, .15, .3, .5], duplication: [.03, .1, .25] };
const scaled = (metric) => metric in defaultLimits;
const rankTitles = {
    complexity: labelOf('complexity'),
    cognitive: labelOf('cognitive'),
    missedLines: 'Uncovered lines',
    duplication: 'Duplicated lines',
    churn: 'Git changes',
};
/** What a metric is called where files are ranked by it. */
export function rankTitle(metric) {
    return pluginSpec(metric)?.label ?? rankTitles[metric];
}
export function scoreLabel(metric) {
    return rankTitle(metric).toLowerCase();
}
function ascending(values) {
    return values.every((value, index) => Number.isFinite(value) && (index === 0 || value > values[index - 1]));
}
/** The profile's limits for a metric with a meaningful scale, when they are well formed; otherwise the defaults. */
export function bandLimits(details, metric) {
    if (!scaled(metric))
        return undefined;
    const given = details.bands?.[metric];
    return given?.length === defaultLimits[metric].length && ascending(given) ? given : defaultLimits[metric];
}
/** The legend's words for the bands, written from the limits: 1–5, 6–10 … 51+, or <5%, 5–15% … 50%+. */
function bandLabels(metric, limits) {
    const percent = (value) => `${+(value * 100).toFixed(1)}%`;
    if (metric === 'complexity') {
        return [...limits.map((limit, index) => `${index === 0 ? 1 : limits[index - 1] + 1}–${limit}`), `${limits.at(-1) + 1}+`];
    }
    const ranges = [...limits.map((limit, index) => index === 0 ? `<${percent(limit)}`
            : `${percent(limits[index - 1]).replace('%', '')}–${percent(limit)}`), `${percent(limits.at(-1))}+`];
    return metric === 'duplication' ? ['0%', ...ranges] : ranges;
}
/** What the map's colour shows, said to assistive technology. */
export function meaningOf(metric) {
    const plugin = pluginSpec(metric);
    if (!plugin)
        return bandMeaning[metric];
    return `Colour shows ${plugin.label.toLowerCase()}${plugin.descriptor.bands ? '' : ' relative to the highest file, or folder at the same depth'}.`;
}
function step(value, bounds) {
    const index = bounds.findIndex(limit => value <= limit);
    return index < 0 ? bounds.length : index;
}
/** The share of a count the score is, or none where the file counts nothing. */
function share(score, whole) {
    return whole && whole > 0 ? score / whole : null;
}
const scales = {
    complexity: (_, score, limits) => step(score, limits),
    missedLines: (entry, score, limits) => {
        const ratio = share(score, entry.lineCoverage?.total);
        return ratio === null ? 0 : step(ratio, limits);
    },
    duplication: (entry, score, limits) => {
        const ratio = share(score, entry.quality?.codeLines);
        return score === 0 || ratio === null ? 0 : 1 + step(ratio, limits);
    },
};
/** Maps a score to one of five colour bands: by the profile's limits where the metric has a meaningful scale. */
export function band(entry, metric, score, highest, limits = scaled(metric) ? defaultLimits[metric] : undefined) {
    const plugin = pluginSpec(metric);
    if (plugin)
        return bandOf(plugin, score, highest);
    const scale = scales[metric];
    return scale ? scale(entry, score, limits) : Math.min(4, Math.floor(score / highest * 5));
}
/** Each file's worst level spread to the folders above it. */
export function spreadLevels(files) {
    const spread = new Map(files);
    for (const [path, level] of files)
        for (const folder of ancestors(path))
            spread.set(folder, Math.max(level, spread.get(folder) ?? 0));
    return spread;
}
/**
 * The top of a relative colour scale, per level: files against files, folders against the folders at their depth. One
 * scale for both would let a folder's sum outshine every file in it, and the map would read hot folders over pale files.
 */
export function relativePeaks(entries) {
    const level = (path, folder) => folder ? `folder:${path.split('/').length}` : 'file';
    const peaks = new Map();
    for (const { path, folder, score } of entries)
        peaks.set(level(path, folder), Math.max(peaks.get(level(path, folder)) ?? 1, score ?? 0));
    return (path, folder) => peaks.get(level(path, folder)) ?? 1;
}
function swatches() {
    const bar = element('span', 'fx-scale-bar');
    bar.setAttribute('aria-hidden', 'true');
    for (let index = 0; index < 5; index++) {
        const swatch = element('span');
        swatch.dataset.band = String(index);
        bar.append(swatch);
    }
    return bar;
}
/** Each word stands under its band. */
function labelledScale(legend, labels) {
    legend.classList.add('fx-scale-labelled');
    const words = element('span', 'fx-scale-words');
    words.append(...labels.map(label => element('span', '', label)));
    legend.append(swatches(), words);
    return legend;
}
export function levelLegend() {
    const legend = element('p', 'fx-scale fx-note');
    legend.dataset.palette = 'grade';
    return labelledScale(legend, ['None', 'Low', 'Medium', 'High', 'Critical']);
}
function legendLabels(metric, limits) {
    const plugin = pluginSpec(metric);
    if (plugin)
        return pluginBandLabels(plugin) ?? undefined;
    return scaled(metric) && limits ? bandLabels(metric, limits) : undefined;
}
export function sunburstLegend(metric, limits) {
    const legend = element('p', 'fx-scale fx-note');
    const labels = legendLabels(metric, limits);
    if (labels)
        return labelledScale(legend, labels);
    legend.append(element('span', '', 'Lower'), swatches(), element('span', '', 'Higher'));
    return legend;
}
/** The map's colours: grade steps for security, grey for nothing measured, heat where higher is worse, else neutral. */
export function palette(metric, levels, grey) {
    if (levels)
        return 'grade';
    if (grey)
        return 'none';
    const plugin = pluginSpec(metric);
    return (plugin ? bandScale(plugin) === 'heat' : metric !== 'churn') ? 'heat' : 'neutral';
}
