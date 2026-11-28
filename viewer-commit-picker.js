import { authorLabel, commitIndex, dayTime, hit, marked, searchCommits, when } from './viewer-commits.js';
import { element } from './viewer-dom.js';
/** What a commit is called in the list: its subject, or its date where the history has no messages. */
function commitTitle(entry) {
    const time = when(entry);
    return entry.subject ?? (time === null ? '' : dayTime.format(new Date(time)));
}
function typedWords(text) {
    return text.trim().toLowerCase().split(/\s+/).filter(Boolean);
}
/** The line under the list: how many commits match, or how many the branch has. */
export function pickerStatus(total, listed, typed, branch) {
    if (total === 0)
        return `No commits match “${typed}”`;
    if (total > listed)
        return `${listed} of ${total.toLocaleString()} ${typed ? 'matches' : 'commits'} · type to narrow`;
    return `${total.toLocaleString()} ${total === 1 ? 'commit' : 'commits'}${branch && !typed ? ` on ${branch}` : ''}`;
}
/**
 * The second line names who and when, then the tags: two fit, the rest fold into a count, and the tags the search
 * found come first. The email appears only when the search found it.
 */
function metaLine(entry, words) {
    const time = when(entry);
    const meta = [authorLabel(entry.author, words), entry.subject && time !== null ? dayTime.format(new Date(time)) : null].filter(Boolean).join(' · ');
    const tags = [...entry.tags ?? []].sort((a, b) => Number(hit(b, words)) - Number(hit(a, words)));
    if (!meta && !tags.length)
        return [];
    const line = element('span', 'commit-option-meta');
    line.append(marked('span', 'commit-meta', meta, words), ...tags.slice(0, 2).map(tag => marked('span', 'commit-tag', tag, words)));
    if (tags.length > 2) {
        const more = element('span', 'commit-tag commit-tag-more', `+${tags.length - 2}`);
        more.title = tags.slice(2).join('\n');
        line.append(more);
    }
    return [line];
}
function optionItem(entry, position, words, selected) {
    const item = element('li', 'commit-option');
    item.id = `commit-option-${position}`;
    item.setAttribute('role', 'option');
    item.setAttribute('aria-selected', String(selected));
    const head = element('span', 'commit-option-head');
    head.append(marked('span', 'commit-hash', entry.sha.slice(0, 7), words, true), marked('span', 'commit-subject', commitTitle(entry), words));
    item.append(head, ...metaLine(entry, words));
    item.addEventListener('mousedown', event => event.preventDefault());
    return item;
}
function searchInput(searchable) {
    const input = element('input', 'commit-search');
    input.type = 'text';
    input.autocomplete = 'off';
    input.spellcheck = false;
    input.placeholder = searchable ? 'Message, tag, author, date or hash' : 'Hash or date';
    input.setAttribute('role', 'combobox');
    input.setAttribute('aria-autocomplete', 'list');
    input.setAttribute('aria-expanded', 'false');
    return input;
}
function branchName(branch) {
    const name = element('span', 'commit-branch', branch);
    name.title = `Branch ${branch}`;
    name.setAttribute('aria-hidden', 'true');
    return name;
}
/** Arrows step through the open list and wrap around it; a page moves eight and stops at either end. */
const arrows = new Map([['ArrowDown', 1], ['ArrowUp', -1]]);
const pages = new Map([['PageDown', 8], ['PageUp', -8]]);
/**
 * Renders the picker into the slot; picking a commit calls back with its SHA. The branch, when the configuration names
 * one, heads the field: the commits are that branch's history.
 */
export function createCommitPicker(slot, entries, pick, branch) {
    const index = commitIndex(entries);
    const bySha = new Map(entries.map(entry => [entry.sha, entry]));
    let current = entries[0].sha;
    let active = 0;
    let matches = [];
    // One control: the field shows the selected commit and, once focused, takes the search; the list opens attached
    // under it, inside the same frame.
    const field = element('div', 'commit-field');
    const input = searchInput(entries.some(entry => entry.subject || entry.author || entry.tags?.length));
    const shown = element('span', 'commit-current');
    shown.setAttribute('aria-hidden', 'true');
    // The selected commit is drawn over the empty input, so both share one box after the branch.
    const entry = element('span', 'commit-entry');
    entry.append(input, shown);
    if (branch)
        field.append(branchName(branch));
    field.append(entry);
    const popover = element('div', 'commit-popover');
    popover.hidden = true;
    const list = element('ul', 'commit-options');
    list.id = 'commit-options';
    list.setAttribute('role', 'listbox');
    list.setAttribute('aria-label', 'Commits');
    input.setAttribute('aria-controls', list.id);
    const status = element('p', 'commit-status');
    status.setAttribute('aria-live', 'polite');
    popover.append(list, status);
    slot.replaceChildren(field, popover);
    function label() {
        const chosen = bySha.get(current);
        shown.replaceChildren(element('span', 'commit-hash', chosen.sha.slice(0, 7)), element('span', 'commit-current-text', commitTitle(chosen)));
        field.title = [chosen.sha, chosen.subject].filter(Boolean).join('\n');
        const subject = chosen.subject ? `: ${chosen.subject}` : '';
        input.setAttribute('aria-label', `${branch ? `Branch ${branch}, commit` : 'Commit'} ${chosen.sha.slice(0, 7)}${subject}. Type to search commits`);
    }
    function update() {
        const typed = input.value.trim();
        const result = searchCommits(index, input.value);
        matches = result.matches;
        active = typed ? 0 : Math.max(0, matches.findIndex(match => match.entry.sha === current));
        const words = typedWords(input.value);
        list.replaceChildren(...matches.map((match, position) => {
            const item = optionItem(match.entry, position, words, match.entry.sha === current);
            item.addEventListener('click', () => choose(match.entry.sha));
            item.addEventListener('mousemove', () => {
                if (active === position)
                    return;
                active = position;
                highlight();
            });
            return item;
        }));
        status.textContent = pickerStatus(result.total, matches.length, typed, branch);
        highlight();
    }
    function highlight() {
        for (const [position, item] of [...list.children].entries())
            item.classList.toggle('active', position === active);
        const item = list.children[active];
        if (!item) {
            input.removeAttribute('aria-activedescendant');
            return;
        }
        input.setAttribute('aria-activedescendant', item.id);
        item.scrollIntoView({ block: 'nearest' });
    }
    const isOpen = () => !popover.hidden;
    function open() {
        if (isOpen())
            return;
        popover.hidden = false;
        slot.classList.add('open');
        input.setAttribute('aria-expanded', 'true');
        update();
        document.addEventListener('pointerdown', outside, true);
    }
    function close() {
        if (!isOpen())
            return;
        popover.hidden = true;
        slot.classList.remove('open');
        input.setAttribute('aria-expanded', 'false');
        input.removeAttribute('aria-activedescendant');
        input.value = '';
        document.removeEventListener('pointerdown', outside, true);
    }
    function outside(event) {
        if (!slot.contains(event.target))
            close();
    }
    function choose(sha) {
        close();
        input.blur();
        if (sha !== current)
            pick(sha);
    }
    function move(to) {
        active = to;
        highlight();
        return true;
    }
    /** A page key moves through the list whether or not it is open, and keeps the page itself from scrolling. */
    function pageKey(key) {
        const by = pages.get(key);
        return by !== undefined && move(Math.min(matches.length - 1, Math.max(0, active + by)));
    }
    /** What a key does while the list is open; whether it took the key from the browser. */
    function openKey(key) {
        const by = arrows.get(key);
        if (by !== undefined)
            return move(matches.length === 0 ? 0 : (active + by + matches.length) % matches.length);
        if (key === 'Enter') {
            const match = matches[active];
            if (match)
                choose(match.entry.sha);
            return true;
        }
        if (key !== 'Escape')
            return false;
        close();
        return true;
    }
    field.addEventListener('pointerdown', event => {
        if (event.target !== input)
            event.preventDefault();
        input.focus();
        open();
    });
    input.addEventListener('focus', open);
    input.addEventListener('blur', () => { if (!slot.matches(':hover'))
        close(); });
    input.addEventListener('input', () => isOpen() ? update() : open());
    input.addEventListener('keydown', event => {
        // A closed list opens on the keys that would move through it or pick from it.
        const opens = !isOpen() && (event.key === 'ArrowDown' || event.key === 'ArrowUp' || event.key === 'Enter');
        if (opens)
            open();
        if (opens || pageKey(event.key) || (isOpen() && openKey(event.key)))
            event.preventDefault();
    });
    label();
    return {
        set(sha) {
            if (!bySha.has(sha))
                return;
            current = sha;
            label();
        },
        get hidden() { return slot.hidden === true; },
        set hidden(value) {
            if (value)
                close();
            slot.hidden = value;
        },
    };
}
