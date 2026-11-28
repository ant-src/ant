import { count } from './viewer-dom.js';
import { sized } from './viewer-languages.js';
import { localCatalog } from './viewer-inspection.js';
import { rollUp } from './viewer-metric-registry.js';
/** The sum of what the files measured; null where none of them did. */
export function sum(files, read) {
    let total = null;
    for (const entry of files) {
        const value = read(entry);
        if (value != null)
            total = (total ?? 0) + value;
    }
    return total;
}
function zeroSum(files, read) {
    return sum(files, read) ?? 0;
}
/** The quality counts that add up, in the order a folder's totals list them. */
const qualitySums = [
    ['codeLines', entry => entry.quality?.codeLines],
    ['commentLines', entry => entry.quality?.commentLines],
    ['classes', entry => entry.quality?.classes],
    ['cognitiveComplexity', entry => entry.quality?.cognitiveComplexity],
    ['churn', entry => entry.quality?.churn],
];
function qualityTotals(files) {
    const quality = {};
    for (const [key, read] of qualitySums) {
        const value = sum(files, read);
        if (value !== null)
            quality[key] = value;
    }
    if (files.some(entry => entry.quality?.duplication)) {
        quality.duplication = { duplicatedLines: zeroSum(files, entry => entry.quality?.duplication?.duplicatedLines),
            codeLines: zeroSum(files, entry => entry.quality?.duplication?.codeLines) };
    }
    if (files.some(entry => entry.quality?.maintainability)) {
        quality.maintainability = { smells: zeroSum(files, entry => entry.quality?.maintainability?.smells),
            effortMinutes: sum(files, entry => entry.quality?.maintainability?.effortMinutes) };
    }
    return quality;
}
function peakComplexity(files) {
    const peaks = files.flatMap(entry => entry.complexity?.maximum == null ? [] : [entry.complexity.maximum]);
    return peaks.length ? peaks.reduce((max, value) => Math.max(max, value), 0) : null;
}
/** Totals for any set of files, measured the way folder totals are: sums, the peak, and ratios of sums. */
export function aggregate(path, files) {
    const covered = files.filter(entry => entry.lineCoverage);
    const values = rollUp(localCatalog(), files);
    const quality = qualityTotals(files);
    return {
        path, files: files.length, lines: sum(files, entry => entry.lines), functions: sum(files, entry => entry.functions),
        complexity: files.some(entry => entry.complexity) ? { sum: zeroSum(files, entry => entry.complexity?.sum), maximum: peakComplexity(files),
            measuredFunctions: zeroSum(files, entry => entry.complexity?.measuredFunctions) } : null,
        lineCoverage: covered.length ? { covered: zeroSum(files, entry => entry.lineCoverage?.covered),
            total: zeroSum(files, entry => entry.lineCoverage?.total) } : null,
        diagnosticCount: zeroSum(files, entry => entry.diagnosticCount),
        sizeMeasuredFiles: files.filter(entry => entry.lines !== null).length, coverageMeasuredFiles: covered.length,
        ...(Object.keys(quality).length ? { quality } : {}),
        ...(values ? { values } : {}),
        // A partial sum would understate the size: only a scope where every file has one gets a total.
        ...(sized(files) ? { bytes: zeroSum(files, entry => entry.bytes) } : {}),
    };
}
/**
 * Where only some of the files whose code was measured carry a measure, their total is partial and says so; a total
 * that all of them carry, or none, says nothing more. A file only a scanner read (a workflow, a manifest) has no code
 * measures and is not expected to carry one.
 */
export function partialNote(files, present) {
    const read = files.filter(entry => entry.quality !== undefined);
    const measured = read.filter(present).length;
    return measured > 0 && measured < read.length ? `Measured in ${count(measured)} of ${count(read.length)} files` : undefined;
}
/** The folders all of whose files are only counted, found in one pass over the files. */
export function unreadFolders(files) {
    const read = new Map();
    for (const file of files) {
        const analysed = file.analysed !== false;
        for (const folder of ancestors(file.path))
            read.set(folder, read.get(folder) === true || analysed);
    }
    return new Set([...read].filter(([, analysed]) => !analysed).map(([folder]) => folder));
}
/** The folders above a path, nearest first. */
export function ancestors(path) {
    const folders = [];
    for (let slash = path.lastIndexOf('/'); slash > 0; slash = path.lastIndexOf('/', slash - 1))
        folders.push(path.slice(0, slash));
    return folders;
}
/** Each path's components, by name. */
export function componentIndex(components) {
    const index = new Map();
    for (const { component, paths } of components)
        for (const path of paths)
            index.set(path, [...index.get(path) ?? [], component.name]);
    return index;
}
