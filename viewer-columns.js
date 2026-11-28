import { dismissible } from './viewer-controls.js';
import { count, element } from './viewer-dom.js';
import { coverageRatio, coverageText, duplicationRatio, duplicationText, effortText } from './viewer-file-figures.js';
import { sized } from './viewer-languages.js';
import { coverageKinds, localValue, measured } from './viewer-inspection.js';
import { firstSortAscending, format } from './viewer-metric-format.js';
import { labelOf, pluginMetrics, size } from './viewer-metrics.js';
/** A column of counts, shown with a dash where the file has none. */
function counted(label, read, extra = {}) {
    return { label, value: entry => read(entry) ?? null, display: entry => count(read(entry)), present: entry => read(entry) != null, ...extra };
}
/** Offered where the folder's totals have this quality measure. */
function hasQuality(key) {
    return directory => directory?.quality?.[key] !== undefined;
}
const hidden = () => false;
const builtinColumns = [
    { label: 'Path', value: entry => entry.path, display: entry => entry.path },
    counted(labelOf('code-lines'), entry => entry.quality?.codeLines, { available: hasQuality('codeLines') }),
    // Ignored files have physical lines only: where there are some and no sizes, the language bar weighs in physical
    // lines, and this column shows the numbers behind it. With sizes, the Size column does.
    counted(labelOf('lines'), entry => entry.lines, { defaultVisible: (directory, scope) => (scope.ignored && !scope.sized) || !hasQuality('codeLines')(directory) }),
    { label: 'Size', value: entry => entry.bytes ?? null, display: entry => size(entry.bytes),
        available: directory => directory?.bytes !== undefined, defaultVisible: (_, scope) => scope.sized },
    counted('Functions', entry => entry.functions, { defaultVisible: hidden }),
    // A file measured without functions has no peak, and is measured all the same.
    counted(labelOf('complexity'), entry => entry.complexity?.maximum, { present: entry => entry.complexity != null }),
    counted(labelOf('cognitive'), entry => entry.quality?.cognitiveComplexity, { available: hasQuality('cognitiveComplexity') }),
    { label: labelOf('coverage'), value: coverageRatio, display: coverageText, available: directory => !directory || coverageRatio(directory) !== null },
    { label: 'Duplication', value: duplicationRatio, display: duplicationText, available: hasQuality('duplication'),
        present: entry => entry.quality?.duplication !== undefined },
    { label: labelOf('effort'), value: entry => entry.quality?.maintainability?.effortMinutes ?? null,
        display: entry => effortText(entry.quality?.maintainability?.effortMinutes), available: hasQuality('maintainability'),
        present: entry => entry.quality?.maintainability !== undefined },
    counted('Git changes', entry => entry.quality?.churn, { available: hasQuality('churn'), defaultVisible: hidden }),
    counted('Diagnostics', entry => entry.diagnosticCount, { defaultVisible: hidden }),
];
/** Registry metrics as Metrics columns, in the order the history lists them. */
function pluginColumns(specs) {
    return specs.map(spec => ({
        label: spec.short,
        value: entry => localValue(entry, spec),
        display: entry => format(spec, localValue(entry, spec)),
        available: () => measured(spec.descriptor),
        defaultVisible: () => spec.column,
        ascending: firstSortAscending(spec),
        // A ratio is no sum: a file without branches has no branch coverage, and the folder's ratio is complete.
        ...spec.descriptor.unit === 'ratio' ? {} : { present: (entry) => localValue(entry, spec) !== null },
    }));
}
/**
 * The table's columns. The other coverage kinds sit beside line coverage; every other registry metric follows
 * Repometry's own columns, except the counts a ratio is made of, which read through the ratio as line coverage's do.
 */
export function tableColumns() {
    const kinds = coverageKinds();
    const afterCoverage = builtinColumns.findIndex(column => column.label === labelOf('coverage')) + 1;
    const others = pluginMetrics().filter(spec => spec.descriptor.scope === 'file' && !spec.ratioPart && !kinds.includes(spec));
    return [builtinColumns[0], ...builtinColumns.slice(1, afterCoverage), ...pluginColumns(kinds), ...builtinColumns.slice(afterCoverage),
        ...pluginColumns(others)];
}
/**
 * The Columns menu: a checkbox per column the scope can show, each remembered once changed. The path always shows;
 * a sort by a column that is hidden falls back to it.
 */
export function columnMenu(columns, settings, scope, change) {
    const menu = element('details', 'fx-columns');
    const toggles = element('div', 'fx-column-toggles');
    menu.append(element('summary', '', 'Columns'), toggles);
    dismissible(menu);
    const isVisible = (index) => index === 0 || Boolean(toggles.querySelector(`[value="${index}"]`)?.checked);
    const keepSort = () => {
        if (isVisible(settings.sort))
            return;
        settings.sort = 0;
        settings.ascending = true;
    };
    const opens = { ignored: scope.files.some(entry => entry.analysed === false), sized: sized(scope.files) };
    columns.forEach((column, index) => {
        if (index === 0 || column.available?.(scope.totals) === false)
            return;
        const label = element('label', 'fx-column-toggle', column.label);
        const toggle = element('input');
        toggle.type = 'checkbox';
        toggle.value = String(index);
        toggle.checked = settings.columns.get(index) ?? column.defaultVisible?.(scope.totals, opens) ?? true;
        toggle.addEventListener('change', () => {
            settings.columns.set(index, toggle.checked);
            keepSort();
            change();
        });
        label.prepend(toggle);
        toggles.append(label);
    });
    keepSort();
    return { menu, isVisible };
}
