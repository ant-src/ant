import { lineCount } from './viewer-pull-trend.js';
const findingReasons = {
    'reports-unavailable': 'Save inspections for the base and head commits to compare findings.',
    'findings-unavailable': 'One or both revisions lack findings data.',
    'partial-inspection': 'One or both revisions have incomplete analysis.',
    'analysis-changed': 'The base and head used different analysis settings.',
};
export function reviewFiles(review) {
    const section = node('section', '', 'pull-section');
    const heading = node('h3', 'Changed files');
    section.append(heading);
    const { diff } = review;
    if (diff.excludedPaths?.length) {
        section.append(node('p', `Path exclusions · ${diff.excludedPaths.join(', ')}`, 'detail pull-exclusions'));
    }
    if (diff.files.length === 0 && diff.gaps.length === 0) {
        section.append(node('p', diff.status === 'unavailable' ? 'Provide LCOV for this head revision to inspect its changed lines.'
            : 'No added or modified source lines in the inspected scope.', 'detail'));
        return section;
    }
    const files = new Map(diff.uncovered.map(file => [file.path, file.lines]));
    const gaps = new Map(diff.gaps.map(gap => [gap.path, gap.reason]));
    const rows = [...diff.files].sort((left, right) => missedCount(right) - missedCount(left) || right.addedLines - left.addedLines)
        .map(record => {
        const { path, addedLines, covered, total } = record;
        const label = node('th');
        const parts = path.split('/');
        for (const [index, part] of parts.entries()) {
            label.append(part);
            if (index < parts.length - 1)
                label.append('/', node('wbr'));
        }
        label.scope = 'row';
        const missing = files.get(path);
        const gap = gaps.get(path);
        gaps.delete(path);
        if (missing) {
            const details = node('details', '', 'pull-missed-lines');
            const ranges = lineRanges(missing);
            details.append(node('summary', `${lineCount(missing.length)} missed ${missing.length === 1 ? 'line' : 'lines'}`), node('span', ranges.join(', ')));
            label.append(details);
        }
        if (gap)
            label.append(node('p', gap, 'detail'));
        const row = node('tr');
        row.append(label, cell(lineCount(addedLines), true), cell(covered === null || total === null ? '—' : `${lineCount(covered)} / ${lineCount(total)}`, true));
        return row;
    });
    if (rows.length > 0)
        section.append(table('Changed file coverage', ['File', 'Added lines', 'Covered / measured'], rows, 'pull-coverage-table'));
    if (gaps.size > 0) {
        const notes = node('ul', '', 'pull-gaps');
        for (const [path, reason] of gaps)
            notes.append(node('li', `${path} · ${reason}`));
        section.append(node('h4', 'Coverage gaps'), notes);
    }
    return section;
}
export function reviewIssues({ findings }) {
    const section = node('section', '', 'pull-section');
    section.append(node('h3', 'New issues'));
    if (findings.added === null || findings.added.length === 0) {
        section.append(node('p', findings.reason ? findingReasons[findings.reason] : 'No new issues in the inspected scope.', 'detail'));
        return section;
    }
    const rows = findings.added.map(finding => {
        const row = node('tr');
        const message = node('th', finding.message);
        message.scope = 'row';
        message.append(node('p', [finding.tool, finding.rule].join(' · '), 'detail'));
        const location = finding.path ? `${finding.path}${finding.line ? `:${finding.line}` : ''}` : '—';
        row.append(message, cell(location), cell(finding.level));
        return row;
    });
    section.append(table('Newly introduced findings', ['Finding', 'Source', 'Severity'], rows));
    return section;
}
function lineRanges(lines) {
    const ranges = [];
    for (let start = 0; start < lines.length;) {
        let end = start;
        while (lines[end + 1] === lines[end] + 1)
            end++;
        ranges.push(start === end ? String(lines[start]) : `${lines[start]}–${lines[end]}`);
        start = end + 1;
    }
    return ranges;
}
function missedCount({ covered, total }) {
    return covered === null || total === null ? -1 : total - covered;
}
function table(label, headers, rows, className = '') {
    const table = node('table', '', `pull-table ${className}`.trim());
    const thead = node('thead');
    const headings = node('tr');
    for (const text of headers) {
        const th = node('th', text);
        th.scope = 'col';
        headings.append(th);
    }
    thead.append(headings);
    const body = node('tbody');
    body.append(...rows);
    table.append(thead, body);
    table.setAttribute('aria-label', label);
    return table;
}
function cell(text, numeric = false) {
    return node('td', text, numeric ? 'numeric' : '');
}
export function node(name, text = '', className = '') {
    const element = document.createElement(name);
    if (text)
        element.textContent = text;
    if (className)
        element.className = className;
    return element;
}
