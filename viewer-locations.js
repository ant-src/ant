export function allLocations(finding) {
    return [
        ...(finding.path ? [{ path: finding.path, line: finding.line, endLine: finding.endLine }] : []),
        ...(finding.related ?? []),
    ];
}
/** The distinct files a finding touches. */
export function locationPaths(finding) {
    return new Set(allLocations(finding).map(location => location.path));
}
export function atPath(finding, path) {
    return finding.path === path || (finding.related?.some(location => location.path === path) ?? false);
}
export function formatLocation(location) {
    if (location.line === undefined)
        return location.path;
    const end = location.endLine !== undefined && location.endLine > location.line ? `–${location.endLine}` : '';
    return `${location.path}:${location.line}${end}`;
}
/** How many findings touch each file, a finding with several places in one file counted once. */
export function findingIndex(findings) {
    const index = new Map();
    for (const finding of findings)
        for (const path of locationPaths(finding))
            index.set(path, (index.get(path) ?? 0) + 1);
    return index;
}
/** The line of a file a finding points at, through its related places when its primary one is elsewhere. */
export function lineIn(finding, path) {
    return finding.path === path ? finding.line : finding.related?.find(location => location.path === path)?.line;
}
