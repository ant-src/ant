import { isPeriod, isTrendView } from './viewer-trend-data.js';
export const pages = ['overview', 'metrics', 'hotspots', 'trends', 'issues', 'diagnostics', 'settings', 'pulls'];
/** The route's parts by their names in the URL, in the order the URL lists them. */
const parameters = [['directory', 'dir'], ['file', 'file'], ['view', 'view'], ['period', 'period'], ['query', 'q'], ['group', 'group'],
    ['line', 'line'], ['category', 'category'], ['level', 'level'], ['rank', 'rank'], ['layer', 'layer']];
export function href(route) {
    const params = new URLSearchParams({ sha: route.sha });
    for (const [key, name] of parameters) {
        const value = route[key];
        if (value)
            params.set(name, String(value));
    }
    // Category lists read as written: category=structure,duplication.
    return `#${route.page}?${params.toString().replaceAll('%2C', ',')}`;
}
export function isFilePage(page) {
    return page === 'metrics' || page === 'hotspots' || page === 'issues' || page === 'diagnostics';
}
/** The pages that read the inspection's details; the summary's counts serve the rest. */
export function needsDetails(page) {
    return isFilePage(page) || page === 'overview' || page === 'settings';
}
/**
 * Reads the route from a location hash. A bare commit opens its Overview; an unknown commit reads as the newest one.
 * Components became a grouping of Metrics, so old links land there.
 */
export function readRoute(hash, commits, newest) {
    const [rawPage = '', query = ''] = hash.slice(1).split('?', 2);
    if (commits.has(rawPage))
        return { page: 'overview', sha: rawPage };
    const params = new URLSearchParams(query);
    const named = params.get('sha') ?? '';
    const sha = commits.has(named) ? named : newest;
    if (rawPage === 'components')
        return { page: 'metrics', sha, group: 'components' };
    const optional = (name) => params.get(name) ?? undefined;
    const view = optional('view');
    const period = optional('period');
    const line = Number(params.get('line'));
    return {
        page: pages.find(candidate => candidate === rawPage) ?? 'overview',
        sha,
        directory: optional('dir'),
        file: optional('file'),
        view: isTrendView(view) ? view : undefined,
        period: isPeriod(period) ? period : undefined,
        query: optional('q'),
        group: optional('group'),
        line: line > 0 ? line : undefined,
        category: optional('category'),
        level: optional('level'),
        rank: optional('rank'),
        layer: optional('layer'),
    };
}
