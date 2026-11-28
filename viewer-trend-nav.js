import { element } from './viewer-dom.js';
import { coverageGrade, formatMetric, gradeScope } from './viewer-metrics.js';
import { groups, trendValue, viewLabel, viewMetric } from './viewer-trend-data.js';
function gradeLabel(text, grade) {
    const label = element('span', 'trend-reading trend-reading-grade', text);
    label.dataset.grade = grade;
    return label;
}
/**
 * What a quality group comes to at the selected commit, as the badges show it: the maintainability grade, and line
 * coverage in its grade's colour, from green at 90% to red below 60%. Other groups carry no total: it would repeat a row.
 */
const groupReadings = {
    Maintainability: (data, selected) => {
        const grade = data.quality[selected.sha]?.maintainability?.grade;
        return grade ? gradeLabel(grade, grade) : null;
    },
    Coverage: (data, selected) => {
        const ratio = trendValue(data, selected, 'coverage');
        return ratio === null ? null : gradeLabel(formatMetric(ratio, 'coverage'), coverageGrade(ratio));
    },
};
/** A series in the list, with its value at the selected commit, in the tabs' counter. */
function seriesLink(data, selected, view, current, period) {
    const item = element('li');
    const link = element('a', '', viewLabel(view));
    link.href = data.link(selected.sha, view, period);
    // The group's own label carries the signal (the grade, the coverage); a row reads like a summary table.
    const metric = viewMetric(view);
    const reading = metric === null ? null : trendValue(data, selected, metric);
    if (metric !== null && reading !== null)
        link.append(element('span', 'trend-value', formatMetric(reading, metric)));
    if (current)
        link.setAttribute('aria-current', 'page');
    item.append(link);
    return item;
}
function groupBlock({ data, selected, current, period }, group, present) {
    const block = element('div', 'trend-group');
    const title = element('h3', '', group);
    const reading = groupReadings[group]?.(data, selected);
    if (reading)
        title.append(reading);
    const list = element('ul');
    list.append(...present.map(view => seriesLink(data, selected, view, view === current, period)));
    const scope = group === 'Maintainability' ? gradeScope(data.quality[selected.sha]) : undefined;
    block.append(title, ...scope ? [element('p', 'trend-group-note', scope)] : [], list);
    return block;
}
/** The series list, and a select with the same series for a narrow screen. */
export function viewNavigation(data, views, current, selected, period) {
    const wrapper = element('aside', 'trend-picker fx-aside');
    const navigation = element('nav', 'trend-nav');
    navigation.setAttribute('aria-label', 'Trend series');
    const select = document.createElement('select');
    for (const [group, members] of groups()) {
        const present = members.filter(view => views.includes(view));
        if (!present.length)
            continue;
        navigation.append(groupBlock({ data, selected, current, period }, group, present));
        const options = document.createElement('optgroup');
        options.label = group;
        options.append(...present.map(view => new Option(viewLabel(view), view, false, view === current)));
        select.append(options);
    }
    select.addEventListener('change', () => { location.hash = data.link(selected.sha, select.value, period); });
    const label = element('label', 'trend-select', 'Series');
    label.append(select);
    wrapper.append(navigation, label);
    return wrapper;
}
export function periodControl(data, sha, view, period) {
    // The choices say what they are ("Last 100 inspections"), so the label is for screen readers only.
    const label = element('label', 'trend-period');
    label.append(element('span', 'visually-hidden', 'Period'));
    const select = document.createElement('select');
    const total = data.reports.length;
    for (const option of ['30', '100']) {
        if (Number(option) < total)
            select.append(new Option(`Last ${option} inspections`, option, false, option === period));
    }
    select.append(new Option(`All ${total} inspections`, 'all', false, period === 'all' || Number(period) >= total));
    select.addEventListener('change', () => { location.hash = data.link(sha, view, select.value); });
    label.append(select);
    return label;
}
