import { element } from './viewer-dom.js';
import { clauseLabel, clauseText, parseWord, sameClause } from './viewer-filter.js';
import { languageColor } from './viewer-languages.js';
function dot(color) {
    const node = element('span', 'qf-dot');
    node.style.background = color;
    return node;
}
/** A chip: what it selects, in words, and a button that takes it away. */
function chip(clause, remove) {
    const node = element('span', `qf-chip qf-chip-${clause.kind}`);
    const { key, value } = clauseLabel(clause);
    if (clause.kind === 'lang')
        node.append(dot(languageColor(clause.value)));
    node.append(element('span', 'qf-chip-key', key), element('span', 'qf-chip-value', value));
    const button = element('button', 'qf-remove');
    button.type = 'button';
    button.setAttribute('aria-label', `Remove ${key} ${value}`);
    button.addEventListener('click', remove);
    node.append(button);
    return node;
}
function option(suggestion, id) {
    const item = element('li', 'qf-option');
    item.id = id;
    item.setAttribute('role', 'option');
    if (suggestion.color)
        item.append(dot(suggestion.color));
    item.append(element('span', 'qf-option-label', suggestion.label));
    if (suggestion.detail)
        item.append(element('span', 'qf-option-detail', suggestion.detail));
    item.addEventListener('mousedown', event => event.preventDefault());
    return item;
}
function groupHeading(group) {
    const heading = element('li', 'qf-group', group);
    heading.setAttribute('role', 'presentation');
    return heading;
}
function combobox(listId) {
    const input = element('input', 'qf-input');
    input.type = 'text';
    input.autocomplete = 'off';
    input.spellcheck = false;
    input.setAttribute('role', 'combobox');
    input.setAttribute('aria-label', 'Filter files');
    input.setAttribute('aria-autocomplete', 'list');
    input.setAttribute('aria-expanded', 'false');
    input.setAttribute('aria-controls', listId);
    return input;
}
/**
 * The field: chips for the chosen clauses, then a text input; the list of suggestions hangs from it in one frame.
 * Backspace in an empty input takes the last chip back into text, so a chip can be edited, not just removed.
 */
export function queryField(clauses, options) {
    let chosen = [...clauses];
    let active = 0;
    let shown = [];
    const root = element('div', 'qf');
    const box = element('div', 'qf-box');
    const chips = element('span', 'qf-chips');
    const list = element('ul', 'qf-options');
    list.id = `qf-options-${Math.random().toString(36).slice(2, 8)}`;
    list.setAttribute('role', 'listbox');
    const input = combobox(list.id);
    const popover = element('div', 'qf-popover');
    popover.hidden = true;
    const help = element('p', 'qf-help', 'Combine chips: same kind widens, different kinds narrow. Type complexity>10 or coverage<80% for a condition.');
    popover.append(list, help);
    box.append(chips, input);
    root.append(box, popover);
    function renderChips() {
        chips.replaceChildren(...chosen.map((clause, index) => chip(clause, () => {
            chosen.splice(index, 1);
            commit();
            input.focus();
        })));
        input.placeholder = chosen.length ? 'Add a filter' : 'Filter by language, component, condition or path';
    }
    function commit() {
        renderChips();
        options.change(chosen, input.value.trim());
        if (!popover.hidden)
            update();
    }
    function update() {
        shown = options.suggestions(input.value);
        active = Math.min(active, Math.max(0, shown.length - 1));
        list.replaceChildren(...shown.flatMap((suggestion, index) => {
            const item = option(suggestion, `${list.id}-${index}`);
            item.addEventListener('click', () => choose(suggestion));
            item.addEventListener('mousemove', () => {
                if (active === index)
                    return;
                active = index;
                highlight();
            });
            return suggestion.group === shown[index - 1]?.group ? [item] : [groupHeading(suggestion.group), item];
        }));
        highlight();
    }
    function highlight() {
        for (const item of list.querySelectorAll('.qf-option'))
            item.classList.toggle('active', item.id === `${list.id}-${active}`);
        const current = list.querySelector(`#${list.id}-${active}`);
        if (!current) {
            input.removeAttribute('aria-activedescendant');
            return;
        }
        input.setAttribute('aria-activedescendant', current.id);
        current.scrollIntoView({ block: 'nearest' });
    }
    function open() {
        if (!popover.hidden)
            return;
        popover.hidden = false;
        root.classList.add('open');
        input.setAttribute('aria-expanded', 'true');
        active = 0;
        update();
        document.addEventListener('pointerdown', outside, true);
    }
    function close() {
        if (popover.hidden)
            return;
        popover.hidden = true;
        root.classList.remove('open');
        input.setAttribute('aria-expanded', 'false');
        document.removeEventListener('pointerdown', outside, true);
    }
    function outside(event) {
        if (!root.contains(event.target))
            close();
    }
    function choose(suggestion) {
        if (!chosen.some(clause => sameClause(clause, suggestion.clause)))
            chosen.push(suggestion.clause);
        input.value = '';
        active = 0;
        commit();
    }
    /** Enter takes the highlighted suggestion, or makes a chip of the typed text. */
    function enter() {
        const typed = input.value.trim();
        const suggestion = shown[active];
        if (suggestion && !popover.hidden)
            choose(suggestion);
        else if (typed)
            choose({ group: '', label: typed, clause: parseWord(typed) ?? { kind: 'path', value: typed } });
    }
    function step(by) {
        open();
        if (shown.length)
            active = (active + by + shown.length) % shown.length;
        highlight();
    }
    /** Each key the field handles, and whether it takes the key from the browser. */
    const keys = new Map([
        ['ArrowDown', () => { step(1); return true; }],
        ['ArrowUp', () => { step(-1); return true; }],
        ['Enter', () => { enter(); return true; }],
        ['Escape', () => {
                if (popover.hidden)
                    return false;
                close();
                return true;
            }],
        ['Backspace', () => {
                if (input.value !== '' || !chosen.length)
                    return false;
                input.value = clauseText(chosen.pop());
                commit();
                return true;
            }],
    ]);
    box.addEventListener('pointerdown', event => {
        if (event.target !== box && event.target !== chips)
            return;
        event.preventDefault();
        input.focus();
    });
    input.addEventListener('focus', open);
    input.addEventListener('input', () => {
        open();
        active = 0;
        update();
        options.change(chosen, input.value.trim());
    });
    input.addEventListener('keydown', event => {
        if (keys.get(event.key)?.())
            event.preventDefault();
    });
    renderChips();
    return {
        element: root,
        text: () => input.value.trim(),
        set(next) {
            chosen = [...next];
            commit();
        },
    };
}
