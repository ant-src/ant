import { isIgnored, isLockfile, languageOf } from './viewer-languages.js';
import { labelOf } from './viewer-metrics.js';
function byKind(table, clause) {
    return table[clause.kind](clause);
}
function ratio(part, whole) {
    return part !== undefined && whole !== undefined && whole > 0 ? part / whole : null;
}
const measures = {
    complexity: { label: labelOf('complexity'), read: entry => entry.complexity?.maximum ?? null },
    cognitive: { label: labelOf('cognitive'), read: entry => entry.quality?.cognitiveComplexity ?? null },
    lines: { label: labelOf('code-lines'), read: entry => entry.quality?.codeLines ?? entry.lines },
    functions: { label: 'Functions', read: entry => entry.functions },
    coverage: { label: labelOf('coverage'), percent: true, read: entry => ratio(entry.lineCoverage?.covered, entry.lineCoverage?.total) },
    duplication: { label: 'Duplication', percent: true,
        read: entry => ratio(entry.quality?.duplication?.duplicatedLines, entry.quality?.duplication?.codeLines) },
    remediation: { label: labelOf('effort'), minutes: true, read: entry => entry.quality?.maintainability?.effortMinutes ?? null },
    changes: { label: 'Git changes', read: entry => entry.quality?.churn ?? null },
};
/** Other plugins' metrics, by key, for the inspection on screen. */
let pluginMeasures = new Map();
export function usePluginMeasures(entries) {
    pluginMeasures = new Map(entries);
}
/** Other plugins' metrics the query can name, by key. */
export function pluginMeasureList() {
    return [...pluginMeasures];
}
export function measureOf(name) {
    return name in measures ? measures[name] : pluginMeasures.get(name);
}
/** Splits the query into clauses; a word that names nothing known becomes a path fragment. */
export function parseQuery(text) {
    const clauses = [];
    for (const match of text.matchAll(/(\w+):"([^"]*)"|(\S+)/g)) {
        const clause = match[1] ? keyed(match[1], match[2]) : parseWord(match[3]);
        if (clause)
            clauses.push(clause);
    }
    return clauses;
}
/** A condition's number in the measure's own unit: a ratio for percentages (80 and 80% alike), minutes for hours. */
function conditionValue(spec, number, unit) {
    if (spec.percent)
        return unit === '%' || number > 1 ? number / 100 : number;
    return unit === 'h' ? number * 60 : number;
}
export function parseWord(word) {
    const keyedWord = /^(\w+):(.+)$/.exec(word);
    if (keyedWord)
        return keyed(keyedWord[1], keyedWord[2]);
    const condition = /^([\w.-]+)(>=|<=|>|<|=)(\d+(?:\.\d+)?)(%|h|min)?$/i.exec(word);
    if (!condition)
        return word ? { kind: 'path', value: word } : null;
    const measure = condition[1].toLowerCase();
    const spec = measureOf(measure);
    if (!spec)
        return null;
    return { kind: 'measure', measure, operator: condition[2], value: conditionValue(spec, Number(condition[3]), condition[4]?.toLowerCase()) };
}
const yesNo = (value) => value === 'yes' || value === 'no';
const flag = (kind, read = value => value) => (value) => yesNo(value) ? { kind, value: read(value) } : null;
const named = (kind) => (value) => ({ kind, value });
/** The keys a clause is written with. The interface calls the files no plugin read "ignored"; `analysed:` stays a silent alias. */
const keys = new Map([
    ['lang', named('lang')], ['language', named('lang')],
    ['component', named('component')],
    ['path', named('path')], ['in', named('path')],
    ['has', (value) => value === 'findings' || value === 'diagnostics' ? { kind: 'has', value } : null],
    ['analysed', flag('analysed')], ['analyzed', flag('analysed')],
    ['ignored', flag('analysed', value => value === 'yes' ? 'no' : 'yes')],
    ['lockfile', flag('lockfile')], ['lockfiles', flag('lockfile')],
]);
function keyed(key, value) {
    return keys.get(key.toLowerCase())?.(value) ?? null;
}
export function formatQuery(clauses) {
    return clauses.map(clauseText).join(' ');
}
const quoted = (value) => /\s|"/.test(value) ? `"${value.replaceAll('"', '')}"` : value;
/** A measure's bound as written: percentages as percentages, everything else as the number. */
function boundText(clause) {
    return measureOf(clause.measure)?.percent ? `${Math.round(clause.value * 1000) / 10}%` : String(clause.value);
}
/** A path that reads as anything else is written out as one. */
function pathText(value) {
    return /^\w+:|[<>=]/.test(value) || /\s/.test(value) ? `path:${quoted(value)}` : value;
}
const clauseTexts = {
    lang: clause => `lang:${quoted(clause.value)}`,
    component: clause => `component:${quoted(clause.value)}`,
    path: clause => pathText(clause.value),
    has: clause => `has:${clause.value}`,
    analysed: clause => `ignored:${clause.value === 'no' ? 'yes' : 'no'}`,
    lockfile: clause => `lockfile:${clause.value}`,
    measure: clause => `${clause.measure}${clause.operator}${boundText(clause)}`,
};
export function clauseText(clause) {
    return byKind(clauseTexts, clause);
}
const operatorSigns = { '>': '>', '>=': '≥', '<': '<', '<=': '≤', '=': '=' };
const yesNoWords = { yes: 'Yes', no: 'No' };
function measureLabel(clause) {
    const { label, percent, minutes } = measureOf(clause.measure) ?? { label: clause.measure };
    const bound = percent ? `${Math.round(clause.value * 1000) / 10}%` : minutes ? `${clause.value} min` : String(clause.value);
    return { key: label, value: `${operatorSigns[clause.operator]} ${bound}` };
}
const clauseLabels = {
    lang: clause => ({ key: 'Language', value: clause.value }),
    component: clause => ({ key: 'Component', value: clause.value }),
    path: clause => ({ key: 'Path', value: clause.value }),
    has: clause => ({ key: 'Has', value: clause.value }),
    analysed: clause => ({ key: 'Ignored', value: clause.value === 'no' ? 'Yes' : 'No' }),
    lockfile: clause => ({ key: 'Lockfile', value: yesNoWords[clause.value] }),
    measure: measureLabel,
};
/** The chip's words: what it selects, in the interface's language rather than the query's. */
export function clauseLabel(clause) {
    return byKind(clauseLabels, clause);
}
export function sameClause(a, b) {
    return clauseText(a) === clauseText(b);
}
const comparisons = {
    '>': (value, bound) => value > bound,
    '>=': (value, bound) => value >= bound,
    '<': (value, bound) => value < bound,
    '<=': (value, bound) => value <= bound,
    '=': (value, bound) => value === bound,
};
const conditions = {
    has: (clause, entry, facts) => clause.value === 'findings' ? (facts.findings.get(entry.path) ?? 0) > 0 : entry.diagnosticCount > 0,
    analysed: (clause, entry) => isIgnored(entry) === (clause.value === 'no'),
    lockfile: (clause, entry) => isLockfile(entry.path) === (clause.value === 'yes'),
    measure: (clause, entry) => {
        const { measure, operator, value: bound } = clause;
        const value = measureOf(measure)?.read(entry) ?? null;
        return value !== null && comparisons[operator](value, bound);
    },
};
/** Whether the file passes the chips of one kind: any of them, or no chip of that kind at all. */
function anyOf(clauses, kind, test) {
    const values = clauses.flatMap(clause => clause.kind === kind ? [clause.value.toLowerCase()] : []);
    return values.length === 0 || values.some(test);
}
/**
 * Whether a file passes the query. Chips of one kind widen the selection (TypeScript or JavaScript), chips of
 * different kinds narrow it (TypeScript in the core component), and every condition must hold.
 */
export function matches(entry, clauses, facts, text = '') {
    const path = entry.path.toLowerCase();
    const language = languageOf(entry).toLowerCase();
    const components = (facts.components.get(entry.path) ?? []).map(name => name.toLowerCase());
    const widened = [
        path.includes(text.toLowerCase()),
        anyOf(clauses, 'lang', value => language === value),
        anyOf(clauses, 'component', value => components.includes(value)),
        anyOf(clauses, 'path', value => path.includes(value)),
    ];
    return widened.every(Boolean) && clauses.every(clause => conditions[clause.kind]?.(clause, entry, facts) ?? true);
}
