// Light highlighting for JavaScript and TypeScript: a scanner, not a parser, that colours comments, strings, regular
// expressions, numbers and keywords. A file is read by its language as the inspection names it; any other language comes
// back as plain text, never in another language's colours.
import { element } from './viewer-dom.js';
import { languageIdOf } from './viewer-languages.js';
const keywords = new Set(('abstract as async await break case catch class const continue debugger declare default delete do else enum export '
    + 'extends false finally for from function get if implements import in instanceof interface let new null of package private protected '
    + 'public readonly return satisfies set static super switch this throw true try type typeof undefined var void while with yield').split(' '));
/** The languages this scanner was written for. */
const scanned = new Set(['javascript', 'typescript']);
/** Words after which a slash opens a regular expression rather than divides. */
const beforeRegex = new Set('return typeof instanceof in of new delete void throw case do else yield await'.split(' '));
/** Past an escape: the index after the closing quote, or the line's end when it stays open. */
function closeQuote(line, from, quote) {
    let at = from;
    while (at < line.length && line[at] !== quote)
        at += line[at] === '\\' ? 2 : 1;
    return at < line.length ? [at + 1, true] : [line.length, false];
}
/** Takes the run up to `stop` as one token. */
function take(cursor, kind, stop) {
    cursor.push(kind, cursor.line.slice(cursor.at, stop));
    cursor.at = stop;
}
function insideComment(cursor) {
    const end = cursor.line.indexOf('*/', cursor.at);
    take(cursor, 'c', end < 0 ? cursor.line.length : end + 2);
    if (end >= 0)
        cursor.state.open = '';
}
function insideTemplate(cursor) {
    const [stop, closed] = closeQuote(cursor.line, cursor.at, '`');
    take(cursor, 's', stop);
    if (!closed)
        return;
    cursor.state.open = '';
    cursor.state.regexAllowed = false;
}
/**
 * The end of a regular expression literal that starts at `at`, its flags included, honouring escapes and character
 * classes; -1 where the line has no closing slash, so the slash divides after all.
 */
function regexEnd(line, at) {
    let end = at + 1;
    let inClass = false;
    while (end < line.length) {
        const current = line[end];
        if (current === '\\') {
            end += 2;
            continue;
        }
        if (current === '/' && !inClass)
            break;
        if (current === '[')
            inClass = true;
        else if (current === ']')
            inClass = false;
        end += 1;
    }
    if (end >= line.length)
        return -1;
    end += 1;
    while (end < line.length && /[a-z]/i.test(line[end]))
        end += 1;
    return end;
}
const char = (cursor, offset = 0) => cursor.line[cursor.at + offset];
/** Where a token can start, tried in order; each takes what it recognises and says whether it did. */
const starts = [
    function lineComment(cursor) {
        if (char(cursor) !== '/' || char(cursor, 1) !== '/')
            return false;
        take(cursor, 'c', cursor.line.length);
        return true;
    },
    function blockComment(cursor) {
        if (char(cursor) !== '/' || char(cursor, 1) !== '*')
            return false;
        take(cursor, 'c', cursor.at + 2);
        cursor.state.open = 'comment';
        return true;
    },
    function template(cursor) {
        if (char(cursor) !== '`')
            return false;
        take(cursor, 's', cursor.at + 1);
        cursor.state.open = 'template';
        return true;
    },
    function quoted(cursor) {
        const quote = char(cursor);
        if (quote !== '"' && quote !== "'")
            return false;
        take(cursor, 's', closeQuote(cursor.line, cursor.at + 1, quote)[0]);
        cursor.state.regexAllowed = false;
        return true;
    },
    function regex(cursor) {
        if (char(cursor) !== '/' || !cursor.state.regexAllowed)
            return false;
        const end = regexEnd(cursor.line, cursor.at);
        if (end < 0)
            return false;
        take(cursor, 's', end);
        cursor.state.regexAllowed = false;
        return true;
    },
    function word(cursor) {
        const found = /^[A-Za-z_$][\w$]*/.exec(cursor.line.slice(cursor.at))?.[0];
        if (!found)
            return false;
        take(cursor, keywords.has(found) ? 'k' : '', cursor.at + found.length);
        cursor.state.regexAllowed = beforeRegex.has(found);
        return true;
    },
    function number(cursor) {
        const found = /^(?:0x[\da-f_]+n?|\d[\d_]*(?:\.\d[\d_]*)?(?:e[+-]?\d+)?n?)/i.exec(cursor.line.slice(cursor.at))?.[0];
        if (!found)
            return false;
        take(cursor, 'n', cursor.at + found.length);
        cursor.state.regexAllowed = false;
        return true;
    },
    function punctuation(cursor) {
        const found = char(cursor);
        take(cursor, '', cursor.at + 1);
        // After a closing bracket a slash divides; after any other punctuation it opens a regular expression.
        if (!/\s/.test(found))
            cursor.state.regexAllowed = found !== ')' && found !== ']';
        return true;
    },
];
/** Collects a line's runs, joining plain text into one run between coloured ones. */
function runs() {
    const result = [];
    let plain = '';
    return {
        push(kind, value) {
            if (!value)
                return;
            if (!kind) {
                plain += value;
                return;
            }
            if (plain)
                result.push(['', plain]);
            plain = '';
            result.push([kind, value]);
        },
        done() {
            if (plain)
                result.push(['', plain]);
            return result;
        },
    };
}
/**
 * Splits lines of JavaScript or TypeScript into coloured runs: comments, strings, regular expressions, numbers and
 * keywords. It honours escapes and character classes, and it tells a regular expression from a division by what comes
 * before the slash. State carries across lines, so a block comment or a template literal stays coloured to its end.
 * Other languages come back as plain text.
 */
export function tokenizer(language) {
    if (!scanned.has(language))
        return line => [['', line]];
    const state = { open: '', regexAllowed: true };
    return line => {
        const collected = runs();
        const cursor = { line, at: 0, state, push: collected.push };
        while (cursor.at < line.length) {
            if (state.open === 'comment')
                insideComment(cursor);
            else if (state.open === 'template')
                insideTemplate(cursor);
            else
                starts.find(start => start(cursor));
        }
        return collected.done();
    };
}
/** A line's runs as nodes: coloured spans and plain text. */
export function runNodes(lineRuns) {
    return lineRuns.map(([kind, value]) => kind ? element('span', `tok-${kind}`, value) : document.createTextNode(value));
}
/** Highlights a file's lines, or a snippet of them, by the file's language. */
export function highlighter(path) {
    const tokenize = tokenizer(languageIdOf(path));
    return line => runNodes(tokenize(line));
}
