import { element, percent } from './viewer-dom.js';
import { sameClause } from './viewer-filter.js';
import { ignoredLabel, languageColor, languageShares, lockfilesLabel } from './viewer-languages.js';
import { labelOf } from './viewer-metrics.js';
/** Lockfiles and the files no plugin read stand at the end in muted tones: categories, not languages. */
const categories = new Map([
    [lockfilesLabel, { clause: { kind: 'lockfile', value: 'yes' }, color: 'var(--lockfile)', className: 'fx-language-lockfile' }],
    [ignoredLabel, { clause: { kind: 'analysed', value: 'no' }, color: 'var(--grade-none)', className: 'fx-language-unread' }],
]);
/** The clause a language's chip stands for. */
function clauseOf(language) {
    return categories.get(language)?.clause ?? { kind: 'lang', value: language };
}
/** Whether the query holds a language's clause: a language by name in any case, a category exactly. */
function holds(clauses, language) {
    const clause = clauseOf(language);
    if (clause.kind !== 'lang')
        return clauses.some(other => sameClause(other, clause));
    return clauses.some(other => other.kind === 'lang' && other.value.toLowerCase() === language.toLowerCase());
}
/** The query with a language's chip pressed or released. */
export function toggleLanguage(clauses, language) {
    const clause = clauseOf(language);
    if (!holds(clauses, language))
        return [...clauses, clause];
    if (clause.kind !== 'lang')
        return clauses.filter(other => other.kind !== clause.kind);
    return clauses.filter(other => !(other.kind === 'lang' && other.value.toLowerCase() === language.toLowerCase()));
}
/** A share too small to round to 0.1% still reads as present, not as 0%. */
export function shareText(ratio) {
    return ratio > 0 && ratio < 0.0005 ? `<${percent(0.001)}` : percent(ratio);
}
function chip(language, share, color, pressed) {
    const category = categories.get(language);
    const toggle = element('button', category ? `fx-language ${category.className}` : 'fx-language');
    toggle.type = 'button';
    toggle.setAttribute('aria-pressed', String(pressed));
    const dot = element('span', 'fx-language-dot');
    dot.style.background = color;
    toggle.append(dot, element('span', 'fx-language-name', language), element('span', 'fx-language-share', share));
    return toggle;
}
/**
 * The languages of the selection as one bar, with a legend whose names toggle their chips in the query. Nothing shows
 * for a single language or no weight at all.
 */
export function renderLanguages(container, files, clauses, set) {
    const { measure, shares } = languageShares(files);
    const total = shares.reduce((sum, [, weight]) => sum + weight, 0);
    container.hidden = shares.length < 2 || total === 0;
    if (container.hidden)
        return;
    // Any language in the query dims the rest of the bar, even one this selection does not hold.
    const anyPicked = clauses.some(clause => clause.kind === 'lang') || [...categories.keys()].some(name => holds(clauses, name));
    const bar = element('div', 'fx-language-bar');
    bar.setAttribute('aria-hidden', 'true');
    const legend = element('div', 'fx-language-legend');
    legend.setAttribute('role', 'group');
    const measured = measure === 'bytes' ? 'size' : labelOf(measure).toLowerCase();
    legend.setAttribute('aria-label', `Languages by ${measured}`);
    for (const [language, weight] of shares) {
        const pressed = holds(clauses, language);
        const color = categories.get(language)?.color ?? languageColor(language);
        const segment = element('span', !anyPicked || pressed ? 'fx-language-segment' : 'fx-language-segment fx-off');
        segment.style.flexGrow = String(weight);
        segment.style.background = color;
        bar.append(segment);
        const toggle = chip(language, shareText(weight / total), color, pressed);
        toggle.addEventListener('click', () => set(toggleLanguage(clauses, language)));
        legend.append(toggle);
    }
    // The bar says what it weighs, so its shares can be checked against a column of the table.
    legend.append(element('span', 'fx-language-measure', `Share of ${measured}`));
    container.replaceChildren(bar, legend);
}
