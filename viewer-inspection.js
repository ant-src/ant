import { usePluginMeasures } from './viewer-filter.js';
import { useLanguages } from './viewer-languages.js';
import { catalogIndex, entryValue } from './viewer-metric-registry.js';
import { extendCatalog, pluginMetrics } from './viewer-metrics.js';
/** How many rows a page of any list holds. */
export const pageSize = 50;
/** The inspection's own catalog: its file and folder values are indexed by it, not by the history's. */
let local = { catalog: undefined, index: new Map() };
export function useDetails(details) {
    const catalog = details.metrics.catalog;
    extendCatalog(catalog);
    useLanguages(details.metrics);
    local = { catalog, index: catalogIndex(catalog) };
    usePluginMeasures(pluginMetrics().filter(spec => spec.descriptor.scope === 'file' && measured(spec.descriptor)).map(spec => [spec.key, {
            label: spec.label, percent: spec.descriptor.unit === 'ratio', minutes: spec.descriptor.unit === 'minutes',
            read: (entry) => localValue(entry, spec),
        }]));
}
export function localCatalog() {
    return local.catalog;
}
/** A file's or a folder's value of a registry metric, read through this inspection's catalog. */
export function localValue(entry, spec) {
    return entryValue(local.index, entry, spec);
}
/** Whether this inspection ran the metric: a ratio ran where its denominator did. */
export function measured(descriptor) {
    return local.index.has(descriptor.ratio?.over ?? descriptor.id);
}
/** Registry metrics of files that this inspection ran. */
export function measuredFileMetrics() {
    return pluginMetrics().filter(spec => spec.descriptor.scope === 'file' && measured(spec.descriptor));
}
/** Coverage kinds beside line coverage (branches, functions), where the coverage report has them. */
export function coverageKinds() {
    return pluginMetrics().filter(spec => spec.plugin === 'repometry' && spec.group === 'Coverage' && spec.descriptor.unit === 'ratio'
        && spec.descriptor.scope === 'file');
}
