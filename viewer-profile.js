import { count, element } from './viewer-dom.js';
import { splitter } from './viewer-layout.js';
import { duration } from './viewer-metrics.js';
import { badges } from './viewer-profile-badges.js';
import { pagedTable } from './viewer-profile-table.js';
/** What each structural rule measures, and in what unit its limit and its cost are counted. */
const ruleNames = {
    'function-complexity': ['Cognitive complexity of a function', 'point'],
    'file-complexity': ['Cognitive complexity of a file', 'point'],
    'function-parameters': ['Parameters of a function', 'parameter'],
    'return-statements': ['Return statements in a function', 'return'],
    'boolean-logic': ['Operators in one condition', 'operator'],
    'nested-control-flow': ['Depth of nested control flow', 'level'],
};
const ruleOrder = Object.keys(ruleNames);
export function renderProfile(container, data) {
    const view = element('div', 'fx-view');
    const split = element('div', 'fx-split profile');
    const main = element('div', 'fx-main profile-main');
    const profile = data.details?.profile;
    if (profile)
        main.append(identity(profile), files(profile), rules(profile), exclusions(profile, data.context));
    else
        main.append(fallback(data));
    const scanners = data.context?.scanners;
    if (scanners?.length) {
        const section = block('Imported scanners');
        section.append(chips(scanners.map(({ name, version }) => version ? `${name} ${version}` : name)));
        main.append(section);
    }
    split.append(main, splitter(split), badges(data.latest, data.details));
    view.append(split);
    container.replaceChildren(view);
}
/** The profile's name and the analyzers it ran, with their versions. */
function identity(profile) {
    const section = block('Analyzers');
    const facts = element('dl', 'profile-facts');
    if (profile.id)
        facts.append(row('Profile', code(profile.id)));
    if (profile.analyzers.length) {
        facts.append(row('Plugins', chips(profile.analyzers.map(analyzer => {
            const tools = Object.entries(analyzer.tools).map(([tool, version]) => `${tool} ${version}`);
            return `${analyzer.id} ${analyzer.version}${tools.length ? ` (${tools.join(', ')})` : ''}`;
        }))));
    }
    if (profile.churn) {
        facts.append(row('Change window', element('span', '', `${count(profile.churn.lookbackDays)} days, at most ${count(profile.churn.maxCommits)} commits`)));
    }
    section.append(facts);
    return section;
}
/**
 * The languages the plugins claim, each with the patterns of the files it reads; an older inspection did not record
 * them. A language whose tool could not run says why, e.g. "go not found on PATH": its files were not analysed.
 */
function claims(languages) {
    if (!languages)
        return none('Not recorded for this inspection');
    if (!languages.length)
        return none('No language plugin');
    const list = element('ul', 'profile-claims');
    list.append(...languages.map(language => {
        const item = element('li');
        const patterns = element('div', 'profile-claim-patterns');
        patterns.append(chips([...language.extensions.map(extension => `**/*${extension}`), ...language.globs]));
        if (language.unavailableReason)
            patterns.append(none(`Not analysed: ${language.unavailableReason}`));
        item.append(element('span', 'profile-claim', language.label), patterns);
        return item;
    }));
    return list;
}
/** What the inspection reads and what it leaves out, as the patterns the profile states. */
function files(profile) {
    const section = block('Files');
    const facts = element('dl', 'profile-facts');
    facts.append(row('Included', claims(profile.languages)));
    facts.append(row('Excluded paths', profile.excludedPaths.length ? chips(profile.excludedPaths) : none()));
    facts.append(row('Excluded directories', profile.excludedDirectories.length
        ? chips(profile.excludedDirectories.map(directory => `**/${directory}/`)) : none()));
    section.append(facts);
    return section;
}
/** Every check the profile applies, from every plugin: what it measures, where its limit lies, what each unit past it costs. */
function rules(profile) {
    const section = block('Rules');
    const rows = profile.rules.map(rule => {
        const [measure, unit] = ruleNames[rule.rule] ?? [null, 'unit'];
        return { plugin: rule.plugin, rule: rule.rule, measure, limit: rule.limit, limitText: count(rule.limit),
            minutes: rule.minutesPerExcess, cost: `${duration(rule.minutesPerExcess)} per ${unit} over` };
    });
    const duplication = profile.duplication;
    if (duplication) {
        rows.push({ plugin: 'duplication', rule: 'identical-code', measure: 'Repeated code', limit: duplication.minCodeLines,
            limitText: `${count(duplication.minCodeLines)} lines, ${count(duplication.minNodes)} nodes`,
            minutes: duplication.minutesPerRepeatedLine, cost: `${duration(duplication.minutesPerRepeatedLine)} per repeated line` });
    }
    rows.sort((a, b) => a.plugin.localeCompare(b.plugin) || order(a.rule) - order(b.rule) || a.rule.localeCompare(b.rule));
    section.append(pagedTable({
        label: 'Rules', noun: ['rule', 'rules'], rows, empty: 'The profile records no rules.', controls: true,
        search: row => `${row.rule} ${row.plugin} ${row.measure ?? ''}`,
        facet: { text: 'Plugin', all: 'All plugins', plural: 'plugins', key: row => row.plugin },
        columns: [
            { label: 'Rule', cell: row => ruleCell(row.rule, row.measure), sort: row => row.rule },
            { label: 'Plugin', cell: row => element('span', 'profile-plugin', row.plugin), sort: row => row.plugin },
            { label: 'Limit', numeric: true, cell: row => row.limitText, sort: row => row.limit },
            { label: 'Cost', numeric: true, cell: row => row.cost, sort: row => row.minutes },
        ],
    }));
    return section;
}
function ruleCell(name, measure) {
    const cell = element('span');
    cell.append(element('code', 'profile-rule', name));
    if (measure)
        cell.append(element('span', 'profile-rule-meta', measure));
    return cell;
}
/** The findings the profile tells the inspection to skip, by tool, rule and path. */
function exclusions(profile, context) {
    const section = block('Rule exclusions');
    const skipped = context?.ignores?.count;
    if (profile.ignore.length && skipped !== undefined) {
        section.append(element('p', 'detail', `${count(skipped)} ${skipped === 1 ? 'finding' : 'findings'} skipped at this commit.`));
    }
    section.append(pagedTable({
        label: 'Rule exclusions', noun: ['exclusion', 'exclusions'], rows: profile.ignore,
        empty: 'None. “Ignore rule” on an issue gives the entry to add to the profile’s ignore list.',
        search: row => [row.tool, row.rule, row.path, row.reason].filter(Boolean).join(' '),
        facet: { text: 'Tool', all: 'All tools', plural: 'tools', key: row => row.tool ?? 'Any tool' },
        columns: [
            { label: 'Rule', cell: row => code([row.tool, row.rule].filter(Boolean).join(' / ') || 'Any rule'), sort: row => `${row.tool ?? ''} ${row.rule ?? ''}` },
            { label: 'Path', cell: row => row.path ? element('span', 'profile-ignore-path', row.path) : none('Everywhere'), sort: row => row.path ?? '' },
            { label: 'Reason', cell: row => row.reason ?? '', sort: row => row.reason ?? '' },
        ],
    }));
    return section;
}
/** Until the details arrive only the exclusions are known; the summary does not carry the profile itself. */
function fallback(data) {
    const section = block('Files');
    const exclusions = data.context?.exclusions;
    if (exclusions) {
        const facts = element('dl', 'profile-facts');
        facts.append(row('Excluded paths', exclusions.paths.length ? chips(exclusions.paths) : none()));
        facts.append(row('Excluded directories', exclusions.directories.length ? chips(exclusions.directories.map(directory => `**/${directory}/`)) : none()));
        section.append(facts);
    }
    if (data.details)
        section.append(element('p', 'detail', 'This inspection did not record its profile.'));
    return section;
}
function order(rule) {
    const index = ruleOrder.indexOf(rule);
    return index < 0 ? ruleOrder.length : index;
}
function block(title) {
    const section = element('section', 'profile-block');
    section.append(element('h3', '', title));
    return section;
}
function row(label, value) {
    const item = element('div');
    const detail = element('dd');
    detail.append(value);
    item.append(element('dt', '', label), detail);
    return item;
}
function chips(values) {
    const list = element('ul', 'profile-chips');
    list.append(...values.map(value => element('li', '', value)));
    return list;
}
function none(text = 'None') {
    return element('span', 'detail', text);
}
function code(text) {
    return element('code', '', text);
}
