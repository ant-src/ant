const totals = new Intl.NumberFormat();
const percentage = new Intl.NumberFormat(undefined, { style: 'percent', maximumFractionDigits: 1 });
const date = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' });
const search = document.querySelector('#search');
const sort = document.querySelector('#sort');
const rows = document.querySelector('#projects');
const summary = document.querySelector('#project-count');
const statusLabels = { complete: 'Complete', partial: 'Partial', 'not-applicable': 'Not applicable' };
try {
    const response = await fetch('./projects.json');
    if (!response.ok)
        throw new Error('No index');
    const data = await response.json();
    if (data.schema !== 'urn:repometry:schema:projects:1' || !Array.isArray(data.projects))
        throw new Error('Invalid index');
    const render = () => {
        const query = search.value.trim().toLowerCase();
        const matches = data.projects.filter(project => project.name.toLowerCase().includes(query));
        matches.sort(compare);
        rows.replaceChildren(...matches.map(projectRow));
        summary.textContent = matches.length === 0 ? 'No matching projects.' : `${matches.length} of ${data.projects.length} projects`;
    };
    search.addEventListener('input', render);
    sort.addEventListener('change', render);
    render();
}
catch {
    summary.textContent = 'Could not load project reports.';
}
function compare(a, b) {
    switch (sort.value) {
        case 'date': return when(b).localeCompare(when(a));
        case 'coverage': return (coverage(a) ?? Infinity) - (coverage(b) ?? Infinity);
        case 'debt': return (b.quality?.maintainability?.debtRatio ?? -1) - (a.quality?.maintainability?.debtRatio ?? -1);
        case 'size': return (b.quality?.codeLines ?? -1) - (a.quality?.codeLines ?? -1);
        default: return a.name.localeCompare(b.name);
    }
}
/** The latest commit's time, falling back to the inspection time for exports without commit times. */
function when(project) {
    return project.inspection?.committedAt ?? project.inspection?.generatedAt ?? '';
}
function projectRow(project) {
    const { name, inspection, quality } = project;
    const maintainability = quality?.maintainability ?? null;
    const row = document.createElement('tr');
    const key = document.createElement('th');
    key.scope = 'row';
    const link = document.createElement('a');
    link.href = `./${encodeURIComponent(name)}/`;
    link.textContent = name;
    key.append(link);
    row.append(key);
    const debt = row.insertCell();
    debt.className = 'numeric';
    debt.textContent = formatPercent(maintainability?.debtRatio ?? null);
    if (maintainability?.grade) {
        const grade = document.createElement('abbr');
        grade.className = 'grade grade-inline';
        grade.title = 'Maintainability grade';
        grade.dataset.grade = maintainability.grade;
        grade.textContent = maintainability.grade;
        debt.append(grade);
    }
    for (const value of [formatPercent(coverage(project)), formatPercent(quality?.duplication?.ratio ?? null),
        quality?.codeLines == null ? '—' : totals.format(quality.codeLines)]) {
        const cell = row.insertCell();
        cell.className = 'numeric';
        cell.textContent = value;
    }
    const last = row.insertCell();
    if (inspection) {
        const sha = document.createElement('code');
        sha.textContent = inspection.sha.slice(0, 7);
        last.append(sha, ` · ${date.format(new Date(when(project)))}`);
        if (inspection.status !== 'complete')
            last.append(` · ${statusLabels[inspection.status]}`);
    }
    else
        last.textContent = 'Not inspected';
    return row;
}
function coverage(project) {
    const data = project.inspection?.lineCoverage;
    return data && data.total > 0 ? data.covered / data.total : null;
}
function formatPercent(value) {
    return value === null ? '—' : percentage.format(value);
}
export {};
