import { node, reviewFiles, reviewIssues } from './viewer-pull-sections.js';
import { formatPercent, lineCount } from './viewer-pull-trend.js';
const points = new Intl.NumberFormat(undefined, { signDisplay: 'exceptZero', maximumFractionDigits: 1 });
const gateWords = {
    passed: 'Gate passed', failed: 'Gate failed', 'not-applicable': 'Gate skipped', indeterminate: 'Gate pending data',
};
const gateStates = { passed: 'complete', indeterminate: 'partial' };
const coverageReasons = {
    'reports-unavailable': 'Save inspections for the base and head commits to compare total coverage.',
    'observations-unavailable': 'The saved inspections lack complete line coverage data.',
    'source-scope-unknown': 'The inspected source scope is unavailable for one of the revisions.',
    'source-scope-changed': 'The base and head used different source scopes.',
    'method-unknown': 'The coverage method is unavailable for one of the revisions.',
    'method-changed': 'The base and head used different coverage methods.',
    'head-coverage-changed': 'The review used a different LCOV input than the saved head inspection.',
};
export function renderPulls(container, records) {
    const heading = node('h2', 'Pull requests');
    // The page title stands outside the card, as on every other page; the review picker shares its row.
    const toolbar = node('div', '', 'section-heading pull-heading');
    toolbar.append(heading);
    const selection = node('section', '', 'panel');
    container.replaceChildren(toolbar, selection);
    if (records.length === 0) {
        selection.append(node('p', 'No saved PR reviews. Run Repometry on a pull request with its base revision and PR number.', 'detail empty'));
        return;
    }
    const view = node('div');
    selection.append(view);
    const display = (review) => view.replaceChildren(reviewSummary(review), reviewFiles(review), reviewIssues(review));
    display(records[0]);
    if (records.length === 1) {
        const number = records[0].number;
        if (number !== null)
            heading.append(node('span', ` #${number}`, 'pull-number'));
        return;
    }
    const label = node('label', 'Saved review', 'pull-picker');
    const select = node('select');
    for (const [index, review] of records.entries()) {
        select.append(new Option(`#${review.number} · ${review.head.slice(0, 7)}`, String(index)));
    }
    label.append(select);
    toolbar.append(label);
    select.addEventListener('change', () => {
        const review = records[select.selectedIndex];
        if (review)
            display(review);
    });
}
/** Total coverage and how it moved from the base. */
function totalCoverage(coverage) {
    const total = coverage.head?.value == null ? 'Unavailable' : formatPercent(coverage.head.value);
    return [total, coverage.delta === null ? 'No comparable base coverage' : `${points.format(coverage.delta)} pp from base`];
}
/** Changed-line coverage, with the measured lines behind it or why there are none. */
function changedCoverage(diff) {
    if (diff.percent === null)
        return diff.status === 'not-applicable' ? 'No measured changed lines' : 'Coverage unavailable';
    return `${lineCount(diff.covered)} of ${lineCount(diff.total)} measured lines${diff.status === 'partial' ? ' · Partial' : ''}`;
}
/** New issues; where the gate counted a different number of quality issues, the line under the figure says so. */
function newIssues({ added, gate, qualityCount }) {
    const value = added === null ? 'Unavailable' : lineCount(added.length);
    if (gate && qualityCount !== null && qualityCount !== added?.length)
        return [value, `Gate counted ${lineCount(qualityCount)} quality issues`];
    return [value, added === null ? '' : 'Compared with base'];
}
function reviewSummary({ base, head, coverage, diff, findings }) {
    const header = node('div', '', 'pull-revisions');
    header.append(node('p', 'Base ', 'detail'), node('code', base.slice(0, 7)), node('span', '→', 'detail'), node('p', 'Head ', 'detail'), node('code', head.slice(0, 7)));
    const summary = node('div', '', 'pull-summary');
    summary.append(summaryMetric('Total coverage', ...totalCoverage(coverage), coverage.gate?.status), summaryMetric('Changed-line coverage', formatPercent(diff.percent), changedCoverage(diff), diff.gate?.status), summaryMetric('New issues', ...newIssues(findings), findings.gate?.status));
    const result = node('div', '', 'pull-summary-block');
    result.append(header, summary);
    const measured = coverage.base || coverage.head || coverage.gate;
    if (coverage.reason && measured)
        result.append(node('p', coverageReasons[coverage.reason], 'detail'));
    return result;
}
function summaryMetric(label, value, detail, gate) {
    const figure = node('div', '', 'pull-fact');
    figure.append(node('span', label, 'detail'), node('strong', value, 'pull-value'));
    if (detail)
        figure.append(node('p', detail, 'detail'));
    if (gate) {
        const badge = node('span', gateWords[gate] ?? gate, 'status pull-gate');
        badge.dataset.status = gateStates[gate] ?? gate;
        figure.append(badge);
    }
    return figure;
}
