import { element } from './viewer-dom.js';
const whole = new Intl.NumberFormat();
const percentage = new Intl.NumberFormat(undefined, { style: 'percent', maximumFractionDigits: 1 });
const date = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });
/** A review's coverage as a percentage, from the 0–100 the review records. */
export function formatPercent(value) {
    return value === null ? 'Unavailable' : percentage.format(value / 100);
}
export function lineCount(value) {
    return whole.format(value);
}
/** Why a review has no changed-line coverage to show. */
const unmeasured = { partial: 'Partial · Unmeasured', 'not-applicable': 'No measurable changes' };
function trendSubject(review) {
    const subject = element('div', 'pull-trend-subject');
    const time = element('time', 'detail', date.format(new Date(review.generatedAt)));
    time.dateTime = review.generatedAt;
    subject.append(element('strong', '', review.number === null ? review.head.slice(0, 7) : `#${review.number}`), time);
    return subject;
}
function trendResult(diff) {
    const value = element('div', 'pull-trend-result');
    const { percent, status, covered, total } = diff;
    if (percent === null) {
        value.append(element('span', 'detail', unmeasured[status] ?? 'Coverage unavailable'));
        return value;
    }
    const result = `${whole.format(covered)} / ${whole.format(total)} lines`;
    value.append(element('strong', '', formatPercent(percent)), element('span', 'detail', status === 'partial' ? `Partial · ${result}` : result));
    return value;
}
function trendTrack(diff) {
    const track = element('div', 'pull-trend-track');
    track.setAttribute('aria-hidden', 'true');
    if (diff.percent === null) {
        track.classList.add('pull-trend-empty');
        return track;
    }
    const bar = element('div', 'pull-trend-bar');
    bar.dataset.status = diff.status;
    bar.style.inlineSize = `${diff.percent}%`;
    track.append(bar);
    return track;
}
export function renderDiffCoverageTrend(container, records) {
    container.replaceChildren(element('h2', '', 'Changed-line coverage by pull request'));
    const reviews = [...records].sort((left, right) => right.generatedAt.localeCompare(left.generatedAt)).slice(0, 12);
    if (reviews.length === 0) {
        container.append(element('p', 'detail', 'No saved PR reviews.'));
        return;
    }
    const list = element('ol', 'pull-trend-list');
    const axis = element('div', 'pull-trend-axis');
    axis.setAttribute('aria-hidden', 'true');
    for (const value of [0, 25, 50, 75, 100])
        axis.append(element('span', '', `${value}%`));
    container.append(axis, list);
    for (const review of reviews) {
        const row = element('li', 'pull-trend-row');
        row.append(trendSubject(review), trendTrack(review.diff), trendResult(review.diff));
        list.append(row);
    }
}
