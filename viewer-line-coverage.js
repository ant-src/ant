export function isSourceLayer(value) {
    return value === 'issues' || value === 'coverage';
}
/** Switches a rendered source between its findings and its coverage, without drawing it again. */
export function setSourceLayer(container, layer) {
    const code = container.querySelector(':scope > .src');
    if (code)
        code.dataset.layer = layer;
}
/** Calls back for each line of each range, from line 1 at the earliest. */
function eachLine(ranges, visit) {
    for (const range of ranges ?? [])
        for (let line = Math.max(1, range[0]); line <= range[1]; line++)
            visit(line, range);
}
/**
 * Each line's coverage, by line number; a line no range names was not measured. Where ranges overlap, the worse
 * state wins: a line a test missed is never shown as covered.
 */
export function lineStates(coverage) {
    const states = new Map();
    for (const state of ['covered', 'partial', 'uncovered'])
        eachLine(coverage[state], line => states.set(line, state));
    return states;
}
/** How often tests ran each line, where the report carries the counts; an older report has none. */
export function lineHits(coverage) {
    const hits = new Map();
    eachLine(coverage.hits, (line, run) => hits.set(line, run[2]));
    return hits;
}
// Two figures are enough for a glance; the tooltip has the whole count.
const compactCount = new Intl.NumberFormat('en', { notation: 'compact', maximumSignificantDigits: 2 });
const fullCount = new Intl.NumberFormat();
/** A count short enough for the gutter: 3×, 980×, 1.2K×, 45K×, 270K×, 3.4M×. */
export function hitLabel(count) {
    return `${count < 1000 ? count : compactCount.format(count)}×`;
}
const lineStateNames = {
    covered: 'Covered', partial: 'Partly covered: some branches missed', uncovered: 'Not covered',
};
/** What a measured line says to a pointer and a screen reader: its state, and how often it ran where that is known. */
export function lineCoverageText(state, count) {
    if (count === undefined || state === 'uncovered')
        return lineStateNames[state];
    return `${lineStateNames[state]}. Ran ${fullCount.format(count)} ${count === 1 ? 'time' : 'times'}`;
}
function times(count) {
    return `${fullCount.format(count)} ${count === 1 ? 'time' : 'times'}`;
}
/**
 * A declaration line reads as its functions' calls, not as the line's own count: that is how often the module was
 * loaded. Called is covered; never called is red, with no number; where only some of the functions starting on the
 * line ran, it is partial.
 */
export function callReading(calls) {
    const total = calls.reduce((sum, count) => sum + count, 0);
    const idle = calls.filter(count => count === 0).length;
    if (total === 0)
        return { state: 'uncovered', text: calls.length === 1 ? 'Not called' : `Not called: none of the ${calls.length} functions here ran` };
    if (idle === 0)
        return { state: 'covered', count: total, text: `Called ${times(total)}` };
    return { state: 'partial', count: total, text: `Partly called: ${idle} of the ${calls.length} functions here never ran. Called ${times(total)}` };
}
/** The calls of the functions that start on each line. */
function callsByLine(coverage) {
    const calls = new Map();
    for (const [line, count] of coverage.functionCalls ?? [])
        calls.set(line, [...calls.get(line) ?? [], count]);
    return calls;
}
/**
 * What each measured line shows. A line that starts functions shows their calls. Inside a function body a line shows
 * how often it ran. Module-level code runs once per load, so its count says nothing about tests: it keeps its tint and
 * shows no number. A report without function bodies cannot tell the two apart, and every line shows its own count.
 */
export function lineReadings(coverage) {
    const states = lineStates(coverage);
    const hits = lineHits(coverage);
    const calls = callsByLine(coverage);
    const inBody = new Set();
    eachLine(coverage.functionBodies, line => inBody.add(line));
    const counted = (line) => coverage.functionBodies === undefined || inBody.has(line);
    const readings = new Map();
    for (const [line, state] of states) {
        const count = counted(line) ? hits.get(line) : undefined;
        readings.set(line, { state, ...(count === undefined ? {} : { count }), text: lineCoverageText(state, count) });
    }
    for (const [line, counts] of calls)
        readings.set(line, callReading(counts));
    return readings;
}
/**
 * The tint the lines without a coverage record take so the strip has no breaks. Coverage records only code, so such a
 * line is a blank, a comment or a declaration no test can run: between two records it takes the state above, the block
 * it closes; before the first record, the state below. After the last record the strip ends.
 */
export function bridgeStates(states) {
    const lines = [...states.keys()].sort((a, b) => a - b);
    const bridged = new Map();
    const fill = (from, to, state) => { for (let at = from; at < to; at++)
        bridged.set(at, state); };
    if (lines.length)
        fill(1, lines[0], states.get(lines[0]));
    for (const [index, line] of lines.slice(0, -1).entries())
        fill(line + 1, lines[index + 1], states.get(line));
    return bridged;
}
