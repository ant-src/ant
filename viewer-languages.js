/** The inspection's languages, which its files point to by index, the colour of each by its name, and each file's id. */
let languages = [];
let colors = new Map();
let ids = new Map();
function nameOf(language) {
    return language.label ?? language.id;
}
/** Takes the languages of the inspection on screen: their names, and the colours the plugins and the catalog give them. */
export function useLanguages(metrics) {
    languages = metrics.languages ?? [];
    colors = new Map(languages.flatMap(language => language.color ? [[nameOf(language), language.color]] : []));
    ids = new Map(metrics.files.flatMap(file => file.language === undefined || !languages[file.language] ? []
        : [[file.path, languages[file.language].id]]));
}
/** The id of a file's language, as the inspection gives it: what decides how its source is read. */
export function languageIdOf(path) {
    return ids.get(path);
}
/**
 * Dependency lockfiles, by exact file name. They are generated, often the largest files in a repository, and no
 * plugin reads them: the language bar and the Metrics list give them a category of their own beside the ignored files.
 */
export const lockfileNames = new Set([
    'package-lock.json', 'npm-shrinkwrap.json', 'yarn.lock', 'pnpm-lock.yaml', 'bun.lock', 'bun.lockb', 'deno.lock',
    'Cargo.lock', 'Gemfile.lock', 'composer.lock', 'poetry.lock', 'Pipfile.lock', 'uv.lock', 'go.sum', 'flake.lock',
    'mix.lock', 'pubspec.lock', 'packages.lock.json', 'gradle.lockfile',
]);
/** Whether a path names a lockfile: its file name, at any depth, exactly as the tools write it. */
export function isLockfile(path) {
    return lockfileNames.has(path.slice(path.lastIndexOf('/') + 1));
}
/** Whether no plugin read a file and it is not a lockfile: what the interface calls ignored. */
export function isIgnored(entry) {
    return entry.analysed === false && !isLockfile(entry.path);
}
/** The language bar's names for the two categories that are not languages. */
export const lockfilesLabel = 'Lockfiles';
export const ignoredLabel = 'Ignored';
/** Whether every file carries its size: inspections made before sizes were recorded carry none. */
export function sized(files) {
    return files.length > 0 && files.every(entry => entry.bytes !== undefined);
}
const weights = {
    bytes: entry => entry.bytes,
    lines: entry => entry.lines,
    'code-lines': entry => entry.quality?.codeLines,
};
/**
 * Size in bytes when every file has it: one measure for code, lockfiles and ignored files alike, untouched by
 * formatting. Without sizes, code lines where every file was read and has them, otherwise physical lines, the one line
 * measure all files have: one bar never adds the two.
 */
function shareMeasure(files) {
    if (sized(files))
        return 'bytes';
    return files.every(entry => entry.analysed !== false && entry.quality?.codeLines !== undefined) ? 'code-lines' : 'lines';
}
/** Where a file's weight goes on the bar: its language, or one of the two categories. */
function shareOf(entry) {
    if (isLockfile(entry.path))
        return lockfilesLabel;
    return entry.analysed === false ? ignoredLabel : languageOf(entry);
}
/** What the language bar weighs, largest language first, then lockfiles, then the ignored files. */
export function languageShares(files) {
    const measure = shareMeasure(files);
    const totals = new Map();
    for (const entry of files)
        totals.set(shareOf(entry), (totals.get(shareOf(entry)) ?? 0) + (weights[measure](entry) ?? 0));
    const categories = [lockfilesLabel, ignoredLabel];
    const shares = [...totals].filter(([name, weight]) => weight > 0 && !categories.includes(name)).sort((a, b) => b[1] - a[1]);
    for (const name of categories) {
        const weight = totals.get(name) ?? 0;
        if (weight > 0)
            shares.push([name, weight]);
    }
    return { measure, shares };
}
/** A file's language by the inspection's name for it; a file it names none for is Other. */
export function languageOf(entry) {
    const language = entry.language === undefined ? undefined : languages[entry.language];
    return language ? nameOf(language) : 'Other';
}
/** A language's colour, as the inspection gives it; a neutral grey for one it gives none. */
export function languageColor(language) {
    return colors.get(language) ?? 'var(--neutral-2)';
}
