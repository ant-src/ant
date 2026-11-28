import { segmented } from './viewer-controls.js';
import { copyIcon, copyOnClick } from './viewer-copy.js';
import { count, element, store, stored } from './viewer-dom.js';
import { securityGrades, securityRisks } from './viewer-metrics.js';
const formatKey = 'repometry.badge-format';
/** The security step by the worst finding: 4 critical, 3 high, 2 medium, 1 anything else, 0 none. */
function securityStep(security) {
    const worst = [security.critical ?? 0, security.high, security.medium].findIndex(amount => amount > 0);
    if (worst >= 0)
        return 4 - worst;
    return security.low || security.unknown ? 1 : 0;
}
function step(value, limits, reached) {
    return value == null ? null : limits.findIndex(limit => reached(value, limit));
}
function areas({ report, quality, context }) {
    const lines = report.lineCoverage;
    const coverage = lines && lines.total > 0 ? lines.covered / lines.total : null;
    const security = context?.security;
    return [
        { name: 'maintainability', label: 'Maintainability', steps: [['A', '<5%'], ['B', '<10%'], ['C', '<20%'], ['D', '<50%'], ['F', '50%+']],
            current: step(quality?.maintainability?.debtRatio, [.05, .1, .2, .5, Infinity], (value, limit) => value < limit), published: true },
        { name: 'coverage', label: 'Coverage', steps: [['A', '90%+'], ['B', '80%+'], ['C', '70%+'], ['D', '60%+'], ['F', '<60%']],
            current: step(coverage, [.9, .8, .7, .6, -Infinity], (value, limit) => value >= limit), published: true },
        { name: 'security', label: 'Security', steps: securityGrades.map((grade, index) => [grade, securityRisks[index]]),
            current: security ? securityStep(security) : null, published: security !== undefined },
    ];
}
/** The scale as one strip in the grade colours, the latest inspection's step at full strength. */
function scale(steps, current) {
    const box = element('div', 'profile-scale');
    const strip = element('div', 'profile-scale-strip');
    const words = element('div', 'profile-scale-words');
    for (const [index, [grade, range, word]] of steps.entries()) {
        const segment = element('span', 'profile-scale-grade', word ?? grade);
        segment.dataset.grade = grade;
        if (index === current)
            segment.setAttribute('aria-current', 'true');
        strip.append(segment);
        words.append(element('span', '', range));
    }
    box.append(strip, words);
    return box;
}
/** Copies the snippet; without clipboard access the snippet is selected, ready for the copy shortcut. */
function copyButton(label, text) {
    const button = element('button', 'fx-copy profile-copy');
    button.type = 'button';
    button.title = label;
    button.setAttribute('aria-label', label);
    button.append(copyIcon());
    copyOnClick(button, text, {
        copied: () => undefined,
        reset: () => undefined,
        refused: () => {
            const range = document.createRange();
            range.selectNodeContents(button.previousElementSibling);
            getSelection()?.removeAllRanges();
            getSelection()?.addRange(range);
        },
    });
    return button;
}
function snippetText(format, label, url, home) {
    return format === 'markdown' ? `[![${label}](${url})](${home})` : `<a href="${home}"><img src="${url}" alt="${label}"></a>`;
}
/** One badge: its image, its scale, and the snippet that embeds it in the chosen format. */
function badgeArea(area, format, snippets) {
    const { name, label, steps, current, published } = area;
    const section = element('section', 'fx-aside-section profile-badge-area');
    section.append(element('h3', '', label));
    if (!published) {
        section.append(scale(steps, current), element('p', 'fx-note', 'No scanner report, so no badge.'));
        return section;
    }
    const image = element('img', 'profile-badge');
    image.src = `./${name}.svg`;
    image.alt = `${label} badge`;
    image.height = 22;
    const url = new URL(`./${name}.svg`, location.href).href;
    const home = new URL('./', location.href).href;
    const box = element('div', 'profile-snippet');
    const text = element('code');
    box.append(text, copyButton(`Copy the ${label.toLowerCase()} badge snippet`, () => text.textContent ?? ''));
    const update = () => { text.textContent = snippetText(format(), label, url, home); };
    update();
    snippets.push(update);
    section.append(image, scale(steps, current), box);
    return section;
}
/** How the maintainability grade's ratio is made: remediation time over the time the debt model gives the code. */
function debtNote(model) {
    if (!model)
        return 'Debt ratio: remediation time over the estimated time to write the code.';
    const name = model.name === 'basic-cocomo-organic' ? 'basic COCOMO (organic)' : model.name;
    return `Debt ratio: remediation time over the time to write the code, by ${name}: ${model.scale} × KLOC^${model.exponent} person-months of ${count(Math.round(model.minutesPerPersonMonth / 60))} hours.`;
}
/** The badges as a README shows them: each with the scale it grades by, and the snippet that embeds it. */
export function badges(latest, details) {
    const panel = element('aside', 'fx-aside profile-aside');
    panel.setAttribute('aria-label', 'Badges');
    const heading = element('div', 'fx-aside-heading');
    heading.append(element('h2', '', 'Badges'));
    let format = stored(() => localStorage, formatKey) === 'html' ? 'html' : 'markdown';
    const snippets = [];
    const toggle = segmented('Snippet format', [['markdown', 'Markdown'], ['html', 'HTML']], format, value => {
        format = value;
        store(() => localStorage, formatKey, value);
        for (const update of snippets)
            update();
    });
    toggle.classList.add('profile-format');
    heading.append(toggle);
    panel.append(heading, ...areas(latest).map(area => badgeArea(area, () => format, snippets)));
    // Footnotes close the sidebar, under the last badge: how the maintainability grade's ratio is made, and what badges show.
    panel.append(element('p', 'fx-note profile-badge-note', debtNote(details?.profile?.debtModel)));
    panel.append(element('p', 'fx-note profile-footnote', 'Badges show the latest inspection; their links work once the report is hosted.'));
    return panel;
}
