import { line, plotRight, text } from './viewer-chart-svg.js';
let measure;
/** Rendered width of a 13 px tick label; an estimate where the page cannot measure text. */
function tickWidth(label) {
    if (measure === undefined) {
        measure = typeof document === 'undefined' ? null : document.createElement('canvas').getContext('2d');
        if (measure)
            measure.font = `13px ${getComputedStyle(document.documentElement).fontFamily}`;
    }
    return measure ? measure.measureText(label).width : label.length * 8.6;
}
function widest(labels) {
    return Math.ceil(Math.max(...labels.map(tickWidth)));
}
/** The ticks' words: percentages with as many decimals as the step needs, zero plainly on a diverging scale. */
function tickTexts(ticks, { ratio, zero, format }) {
    const tickStep = ticks.length > 1 ? ticks[1] - ticks[0] : 1;
    const digits = Math.min(3, Math.max(0, -Math.floor(Math.log10(tickStep * 100) + 1e-9)));
    const tickPercent = new Intl.NumberFormat(undefined, { style: 'percent', minimumFractionDigits: digits, maximumFractionDigits: digits });
    const word = ratio ? (tick) => tickPercent.format(tick) : format;
    return ticks.map(tick => zero && tick === 0 ? '0' : word(tick));
}
/** Draws the value axis and its grid, sets the plot's left edge and width, and returns the value-to-y mapping. */
export function yAxis(svg, layout, scale) {
    const { values, ratio, count, tickUnit, zero } = scale;
    const ticks = niceTicks((zero ? [...values, 0] : values).map(value => value / tickUnit), ratio, count).map(tick => tick * tickUnit);
    const labels = tickTexts(ticks, scale);
    layout.left = widest(labels) + 12;
    // A second scale takes a label column at the right, one label per grid line.
    const secondLabels = scale.rightLabels?.(ticks.length) ?? [];
    const reserve = secondLabels.length ? widest(secondLabels) + 12 - plotRight : 0;
    layout.plotWidth = Math.max(40, layout.width - layout.left - plotRight - reserve);
    const low = ticks[0];
    const high = ticks.at(-1);
    const y = (value) => layout.top + layout.plotHeight - (high === low ? .5 : (value - low) / (high - low)) * layout.plotHeight;
    const across = (at) => ({ x1: layout.left, x2: layout.left + layout.plotWidth, y1: at, y2: at });
    for (const [index, tick] of ticks.entries()) {
        const base = zero ? tick === 0 : index === 0;
        const label = text(labels[index], 'tc-tick', layout.left - 8, y(tick) + 4.5);
        label.setAttribute('text-anchor', 'end');
        svg.append(line(base ? 'tc-base' : 'tc-grid', across(y(tick))), label);
        const second = secondLabels[index];
        if (second !== undefined)
            svg.append(text(second, 'tc-tick tc-tick-secondary', layout.left + layout.plotWidth + 8, y(tick) + 4.5));
    }
    if (zero && low !== 0)
        svg.append(line('tc-grid', across(y(low))));
    return y;
}
function paddedRange(values, ratio) {
    let low = Math.min(...values);
    let high = Math.max(...values);
    // A minimum visible range keeps noise flat: 2 percentage points for ratios, 10% of the value for counts.
    const minimum = ratio ? .02 : Math.max(Math.abs(high) * .1, 2);
    if (high - low < minimum) {
        const middle = (high + low) / 2;
        low = middle - minimum / 2;
        high = middle + minimum / 2;
    }
    const span = high - low;
    low -= span * .12;
    high += span * .12;
    if (Math.min(...values) >= 0)
        low = Math.max(0, low);
    if (ratio && Math.max(...values) <= 1)
        high = Math.min(1, high);
    return [low, high];
}
/** A round step near the given one: 1, 2, 2.5 or 5 times a power of ten. */
export function niceStep(raw) {
    if (!(raw > 0))
        return 1;
    const power = 10 ** Math.floor(Math.log10(raw));
    const fraction = raw / power;
    return ([1, 2, 2.5, 5].find(factor => fraction <= factor) ?? 10) * power;
}
/** Evenly spaced round ticks that cover the data with a margin. */
export function niceTicks(values, ratio, count) {
    if (values.length === 0)
        return ratio ? [0, .5, 1] : [0, 1];
    const [low, high] = paddedRange(values, ratio);
    const raw = (high - low) / count;
    const magnitude = 10 ** Math.floor(Math.log10(raw));
    const step = [1, 2, 2.5, 5, 10].map(factor => factor * magnitude).find(candidate => candidate >= raw) ?? raw;
    const integral = !ratio && values.every(Number.isInteger);
    const unit = integral ? Math.max(1, step) : step;
    const first = Math.floor(low / unit) * unit;
    const last = Math.ceil(high / unit) * unit;
    const ticks = [];
    for (let tick = first; tick <= last + unit / 2; tick += unit)
        ticks.push(Number(tick.toPrecision(12)));
    return ticks;
}
