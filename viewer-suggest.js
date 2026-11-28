import { clauseLabel, parseWord, pluginMeasureList, sameClause } from './viewer-filter.js';
import { isIgnored, isLockfile, languageColor, languageOf, lockfilesLabel } from './viewer-languages.js';
/** Ready-made conditions offered before anything is typed: the questions people ask of a codebase most. */
const presets = [
    { text: 'complexity>10', label: 'Complex functions' },
    { text: 'coverage<80%', label: 'Under 80% covered' },
    { text: 'duplication>0', label: 'Has duplication' },
    { text: 'remediation>0', label: 'Has technical debt' },
    { text: 'changes>0', label: 'Changed recently' },
    { text: 'has:findings', label: 'Has findings' },
];
/** The kinds of clause that read as a condition when typed out whole. */
const conditionKinds = new Set(['measure', 'has', 'analysed', 'lockfile']);
function quantity(total) {
    return `${total.toLocaleString()} ${total === 1 ? 'file' : 'files'}`;
}
/** A facet is offered while nothing is typed, or when the text is in its name or begins its key. */
function offered(typing, name, key) {
    return !typing.query || name.toLowerCase().includes(typing.query) || key.startsWith(typing.query);
}
function tallied(names) {
    const counts = new Map();
    for (const name of names)
        counts.set(name, (counts.get(name) ?? 0) + 1);
    return counts;
}
function languageSuggestions(files, typing) {
    const counts = [...tallied(files.map(file => languageOf(file)))].sort((a, b) => b[1] - a[1]);
    return counts.flatMap(([language, total]) => {
        const clause = { kind: 'lang', value: language };
        if (typing.taken(clause) || !offered(typing, language, 'lang'))
            return [];
        return [{ group: 'Languages', label: language, detail: quantity(total), clause, color: languageColor(language) }];
    });
}
function componentSuggestions(facts, typing) {
    return [...tallied([...facts.components.values()].flat())].flatMap(([name, total]) => {
        const clause = { kind: 'component', value: name };
        if (typing.taken(clause) || !offered(typing, name, 'component'))
            return [];
        return [{ group: 'Components', label: name, detail: quantity(total), clause }];
    });
}
/** Files only counted, never read by a plugin, and lockfiles, are one condition away once the text names them. */
function categorySuggestions(files, typing) {
    const { query } = typing;
    const named = (words) => query.length >= 3
        && words.some(phrase => phrase.startsWith(query) || (phrase.includes(' ') && phrase.includes(query)));
    const categories = [
        { label: 'Ignored files', clause: { kind: 'analysed', value: 'no' }, total: files.filter(isIgnored).length,
            words: ['ignored files', 'ignored:yes', 'not analysed', 'analysed:no'] },
        { label: lockfilesLabel, clause: { kind: 'lockfile', value: 'yes' }, total: files.filter(file => isLockfile(file.path)).length,
            words: ['lockfiles', 'lockfile:yes'] },
    ];
    return categories.flatMap(({ label, clause, total, words }) => total && !typing.taken(clause) && named(words) ? [{ group: 'Conditions', label, detail: quantity(total), clause }] : []);
}
/** Other plugins' metrics are offered once the typed text names them; the presets stay Repometry's own. */
function conditionSuggestions(typing) {
    const { query, taken } = typing;
    const plugins = query ? pluginMeasureList().flatMap(([key, { label }]) => {
        const clause = { kind: 'measure', measure: key, operator: '>', value: 0 };
        if (taken(clause) || !(key.startsWith(query) || label.toLowerCase().includes(query)))
            return [];
        return [{ group: 'Conditions', label: `${label} above zero`, detail: `${label} > 0`, clause }];
    }) : [];
    const ready = presets.flatMap(preset => {
        const clause = parseWord(preset.text);
        const { key, value } = clauseLabel(clause);
        const fits = !query || preset.label.toLowerCase().includes(query) || preset.text.startsWith(query) || key.toLowerCase().includes(query);
        return !taken(clause) && fits ? [{ group: 'Conditions', label: preset.label, detail: `${key} ${value}`, clause }] : [];
    });
    return [...plugins, ...ready];
}
/** A condition typed out whole is offered as itself; otherwise the plugins' metrics and the presets are. */
function typedSuggestions(typed, typing) {
    if (!typed || !conditionKinds.has(typed.kind) || typing.taken(typed))
        return conditionSuggestions(typing);
    const { key, value } = clauseLabel(typed);
    return [{ group: 'Conditions', label: `${key} ${value}`, clause: typed }];
}
/** What the field offers for the typed text: the facets of this codebase first, then conditions, then the path. */
export function suggest(text, files, facts, chosen) {
    const trimmed = text.trim();
    const typing = { query: trimmed.toLowerCase(), taken: clause => chosen.some(other => sameClause(other, clause)) };
    const typed = typing.query ? parseWord(trimmed) : null;
    // A path is offered for any text that reads as no other kind of chip.
    const path = typing.query && (!typed || typed.kind === 'path')
        ? [{ group: 'Path', label: `Path contains “${trimmed}”`, clause: { kind: 'path', value: trimmed } }] : [];
    return [...languageSuggestions(files, typing), ...componentSuggestions(facts, typing), ...categorySuggestions(files, typing),
        ...typedSuggestions(typed, typing), ...path];
}
