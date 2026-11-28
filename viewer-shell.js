import { byId, count } from './viewer-dom.js';
import { labelOf } from './viewer-metrics.js';
import { href, pages } from './viewer-route.js';
import { charted, stableWindow } from './viewer-trend-data.js';
import { seriesPoints, trendLegend } from './viewer-trend-points.js';
import { renderSeries } from './viewer-trend-series.js';
/** The area the Overview opens on: maintainability where it was measured, else coverage, else complexity. */
export function firstArea({ reports, quality }) {
    if (reports.some(report => quality[report.sha]?.maintainability?.debtRatio != null))
        return 'debt';
    return reports.some(report => report.lineCoverage !== null) ? 'coverage' : 'complexity';
}
/** The Overview's quality tabs: one selected and focusable, the arrows moving between the visible ones. */
export function verdictTabs(select) {
    const verdicts = document.querySelectorAll('.verdict-tab');
    const mark = (chosen) => {
        for (const verdict of verdicts) {
            const selected = verdict.dataset.metric === chosen;
            verdict.setAttribute('aria-selected', String(selected));
            verdict.tabIndex = selected ? 0 : -1;
        }
    };
    for (const verdict of verdicts) {
        verdict.addEventListener('click', () => {
            const metric = verdict.dataset.metric;
            mark(metric);
            select(metric);
        });
        verdict.addEventListener('keydown', event => {
            if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft')
                return;
            const visible = [...verdicts].filter(item => !item.closest('.verdict').hidden);
            const next = visible[(visible.indexOf(verdict) + (event.key === 'ArrowRight' ? 1 : -1) + visible.length) % visible.length];
            next.focus();
            next.click();
        });
    }
    return mark;
}
export function showTab(page, visible) {
    const tab = document.querySelector(`#navigation a[data-page="${page}"]`);
    if (tab)
        tab.hidden = !visible;
}
/** Tabs follow the code: pages with nothing to show leave the bar, and issues and diagnostics carry their count. */
export function updateTabs(page, findings, diagnostics) {
    byId('issues-count').textContent = count(findings);
    byId('diagnostics-count').textContent = count(diagnostics);
    showTab('diagnostics', diagnostics > 0 || page === 'diagnostics');
}
export function updateDetailsTabs(page, details) {
    updateTabs(page, details.findings?.length ?? 0, details.diagnostics.length);
}
/** Shows the route's page and points the navigation at its commit; focus follows a page change made from the page. */
export function showPage(current) {
    const previousPage = pages.find(page => !byId(page).hidden);
    const focusWasInPage = previousPage !== undefined && byId(previousPage).contains(document.activeElement);
    byId('inspection-meta').hidden = current.page === 'pulls';
    for (const page of pages)
        byId(page).hidden = page !== current.page;
    if (previousPage && previousPage !== current.page) {
        window.scrollTo({ top: 0 });
        if (focusWasInPage) {
            const section = byId(current.page);
            section.tabIndex = -1;
            section.focus({ preventScroll: true });
        }
    }
    for (const link of document.querySelectorAll('#navigation a')) {
        const page = link.dataset.page;
        link.href = href({ page, sha: current.sha });
        if (page === current.page)
            link.setAttribute('aria-current', 'page');
        else
            link.removeAttribute('aria-current');
    }
    document.querySelector('.brand').href = href({ page: 'overview', sha: current.sha });
}
export function renderTrend(data, selected, metric, compact) {
    // Enough inspections to show a curve, not a pair of dots; the Trends page holds the whole history.
    const size = compact ? 24 : 60;
    const { start, end } = stableWindow(data.reports, data.reports.indexOf(selected), size, 'overview');
    const shown = charted(data.reports.slice(start, end).reverse());
    const link = (sha) => href({ page: 'overview', sha });
    const chart = byId('chart');
    const hasData = renderSeries({ chart, key: byId('trend-key'), data, shown, metric, selected: selected.sha, link }, { compact, height: compact ? undefined : 296, labelCurrent: false });
    chart.hidden = !hasData;
    byId('trend-empty').hidden = hasData;
    // With no chart, its empty line is the panel's one statement: no key, no reading notes, no second "not measured".
    byId('trend').dataset.empty = String(!hasData);
    const points = seriesPoints(data, shown, metric, selected.sha, link);
    // The figures under the chart state the value and its change, so the chart is named for screen readers only.
    byId('trend-title').textContent = `${labelOf(metric)} by commit`;
    const legend = trendLegend(points);
    byId('trend-detail').textContent = legend;
    byId('trend-detail').hidden = legend === '' || !hasData;
}
export async function readOptional(path) {
    try {
        const response = await fetch(path);
        if (!response.ok)
            throw new Error(`Could not read ${path}`);
        return await response.json();
    }
    catch {
        return null;
    }
}
/** The project's name, where the export gives one, in the tab and the header. */
export function showTitle() {
    const title = document.querySelector('meta[name="repometry-project"]')?.content.trim();
    if (!title)
        return;
    document.title = `${title} · Repometry`;
    byId('project-name').textContent = title;
}
