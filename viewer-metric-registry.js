/** Repometry's metrics keep short query words; every other key comes from the descriptor's id. */
const builtinKeys = {
    'repometry.cyclomatic-complexity/v1': 'complexity',
    'repometry.cognitive-complexity/v1': 'cognitive',
    'repometry.code-lines/v1': 'lines',
    'repometry.functions/v1': 'functions',
    'repometry.lcov-line-coverage/v1': 'coverage',
    'repometry.duplicated-lines/v1': 'duplication',
    'repometry.git-change-frequency/v1': 'changes',
    'repometry.lcov-branch-coverage/v1': 'branch-coverage',
    'repometry.lcov-covered-branches/v1': 'covered-branches',
    'repometry.lcov-coverable-branches/v1': 'coverable-branches',
    'repometry.lcov-function-coverage/v1': 'function-coverage',
    'repometry.lcov-covered-functions/v1': 'covered-functions',
    'repometry.lcov-coverable-functions/v1': 'coverable-functions',
};
/** The Repometry metrics the viewer draws from typed fields, with their own labels, columns and charts. */
const handDrawn = new Set([
    'repometry.physical-lines/v1', 'repometry.functions/v1', 'repometry.cyclomatic-complexity/v1', 'repometry.cognitive-complexity/v1',
    'repometry.classes/v1', 'repometry.code-lines/v1', 'repometry.comment-lines/v1',
    'repometry.duplicated-lines/v1', 'repometry.lcov-line-coverage/v1', 'repometry.lcov-covered-lines/v1',
    'repometry.lcov-coverable-lines/v1', 'repometry.git-change-frequency/v1',
]);
/** Words that already mean something: query keys, Trends views and the series the viewer draws by hand. */
const reserved = new Set(['lang', 'language', 'component', 'path', 'in', 'has', 'analysed', 'ignored', 'lockfile', 'lockfiles', 'remediation', 'missedlines', 'churn',
    'debt', 'debt-changes', 'diff-coverage', 'effort', 'smells', 'files', 'classes', 'comments', 'code-lines', 'security',
    'coverable-lines', 'covered-lines', 'uncovered-lines', 'security-high', 'security-medium', 'security-low',
    ...Object.values(builtinKeys)]);
/** Curated groups first, in this order; groups named by plugins follow in catalog order. */
/** The metrics the viewer draws from their typed fields; a Repometry metric it has no code for goes by its descriptor. */
export function isBuiltin(descriptor) {
    return handDrawn.has(descriptor.id);
}
function parts(id) {
    const bare = id.replace(/\/v\d+$/, '');
    const dot = bare.indexOf('.');
    return dot < 0 ? { plugin: bare, name: bare } : { plugin: bare.slice(0, dot), name: bare.slice(dot + 1) };
}
/** One spec per descriptor, in catalog order; descriptors listed twice by id count once. */
export function metricSpecs(catalog) {
    const unique = [...new Map(catalog.map(descriptor => [descriptor.id, descriptor])).values()];
    // Two versions of one metric share a key: a commit measures one of them. Two plugins naming the same metric do not.
    const owners = new Map();
    for (const { id } of unique) {
        const { plugin, name } = parts(id);
        owners.set(name, (owners.get(name) ?? new Set()).add(plugin));
    }
    // The counts a ratio is made of (covered and measured branches) read through the ratio: no chart, column or rank of their own.
    const ratioParts = new Set(unique.flatMap(({ ratio }) => ratio ? [ratio.of, ratio.over] : []));
    return unique.map(descriptor => {
        const { plugin, name } = parts(descriptor.id);
        const builtin = builtinKeys[descriptor.id];
        const key = builtin ?? (owners.get(name).size > 1 || reserved.has(name.toLowerCase()) ? `${plugin}.${name}` : name);
        const rolls = descriptor.aggregate !== 'none' || descriptor.ratio !== undefined;
        const ranked = descriptor.scope === 'file' && descriptor.better !== 'none' && !ratioParts.has(descriptor.id);
        return {
            descriptor, key, plugin,
            label: descriptor.label,
            short: descriptor.short ?? descriptor.label,
            group: descriptor.group ?? plugin,
            // Repometry's own columns are curated; its metrics without a hand-made column are there to pick, not shown first.
            column: ranked && plugin !== 'repometry',
            hotspot: ranked && rolls,
            trend: !ratioParts.has(descriptor.id) && (descriptor.scope === 'repository' || rolls),
            ratioPart: ratioParts.has(descriptor.id),
        };
    });
}
/** Positions of a catalog's metrics by id: summary indexes for commits, the detail's own indexes for files. */
export function catalogIndex(catalog) {
    return new Map((catalog ?? []).map((descriptor, index) => [descriptor.id, index]));
}
/** Specs by key, for the query field and the Trends views. Where versions share a key, the one the index holds wins. */
export function byKey(specs, index) {
    const keys = new Map();
    for (const spec of specs) {
        if (!keys.has(spec.key) || index?.has(spec.descriptor.id))
            keys.set(spec.key, spec);
    }
    return keys;
}
/** The id whose presence says a metric was measured: a ratio is measured where its denominator is. */
function measuredId(spec) {
    return spec.descriptor.ratio?.over ?? spec.descriptor.id;
}
/**
 * The repository's value at a commit. A commit lists every metric it measured, null where nothing rolls up; a metric
 * it does not list did not run, and has no value, so a chart breaks its line instead of dropping to zero.
 */
export function commitValue(index, commit, spec) {
    const values = commit.values;
    const at = index.get(measuredId(spec));
    if (!values || at === undefined || !(at in values))
        return null;
    const ratio = spec.descriptor.ratio;
    if (!ratio)
        return values[at] ?? null;
    const of = index.get(ratio.of);
    const total = values[at] ?? 0;
    return total > 0 && of !== undefined ? (values[of] ?? 0) / total : null;
}
/**
 * A file's or folder's value against the detail's own catalog. None for a file no plugin analysed, nor for a metric
 * this inspection did not run. Within a measured metric a missing value is zero: zeros are left out of the data.
 */
export function entryValue(index, entry, spec) {
    if (entry.analysed === false)
        return null;
    const at = index.get(measuredId(spec));
    if (at === undefined)
        return null;
    const ratio = spec.descriptor.ratio;
    if (!ratio)
        return entry.values?.[at] ?? 0;
    const of = index.get(ratio.of);
    const total = entry.values?.[at] ?? 0;
    return total > 0 && of !== undefined ? (entry.values?.[of] ?? 0) / total : null;
}
/** The values of a group of files the viewer gathers itself (a filtered folder, a component), by each metric's roll-up. */
export function rollUp(catalog, entries) {
    if (!catalog?.length)
        return undefined;
    const values = {};
    catalog.forEach((descriptor, index) => {
        if (descriptor.aggregate === 'none')
            return;
        let total = 0;
        for (const entry of entries) {
            if (entry.analysed === false)
                continue;
            const value = entry.values?.[index] ?? 0;
            total = descriptor.aggregate === 'max' ? Math.max(total, value) : total + value;
        }
        if (total !== 0)
            values[index] = total;
    });
    return values;
}
