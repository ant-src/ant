import { collator, count, element } from './viewer-dom.js';
export function searchField(text, placeholder) {
    const label = element('label', 'fx-field');
    label.append(element('span', 'fx-field-text', text));
    const input = element('input');
    input.type = 'search';
    input.placeholder = placeholder;
    input.spellcheck = false;
    label.append(input);
    return { label, input };
}
/** Two or three exclusive views as one segmented switch; the pressed segment is the current view. */
export function segmented(text, options, selected, change) {
    const group = element('div', 'fx-segmented');
    group.setAttribute('role', 'group');
    group.setAttribute('aria-label', text);
    const buttons = options.map(([value, name]) => {
        const button = element('button', '', name);
        button.type = 'button';
        button.setAttribute('aria-pressed', String(value === selected));
        button.addEventListener('click', () => {
            for (const other of buttons)
                other.setAttribute('aria-pressed', String(other === button));
            change(value);
        });
        return button;
    });
    group.append(...buttons);
    return group;
}
/** Presses the segment with this name, as a click on it would look, without running its change. */
export function pressSegment(group, name) {
    group.querySelectorAll('button').forEach(button => button.setAttribute('aria-pressed', String(button.textContent === name)));
}
/** A details menu that closes on a click elsewhere or on Escape, as a menu does. */
export function dismissible(menu) {
    const close = (event) => {
        const key = event instanceof KeyboardEvent;
        if (key ? event.key !== 'Escape' : menu.contains(event.target))
            return;
        if (menu.open && key)
            menu.querySelector('summary')?.focus();
        menu.open = false;
    };
    menu.addEventListener('toggle', () => {
        const listen = menu.open ? document.addEventListener.bind(document) : document.removeEventListener.bind(document);
        listen('pointerdown', close, true);
        listen('keydown', close);
    });
}
/** What a multiple choice's button says: all, up to two names, or how many are chosen. */
export function chosenLabel(chosen, all, plural) {
    if (chosen.length === 0)
        return all;
    // Two choices read by name, so a filtered view says what it holds; more read as a count.
    return chosen.length <= 2 ? chosen.join(', ') : `${chosen.length} ${plural}`;
}
/**
 * A dropdown of checkboxes: none ticked means all. The button names what is chosen, so the filter reads at a glance;
 * each option carries its count. It hides itself when there is nothing to choose between.
 */
export function multiChoice({ text, all, plural }, options, change, initial = []) {
    const offered = new Set(options.map(([value]) => value));
    const values = new Set([...initial].filter(value => offered.has(value)));
    const menu = element('details', 'fx-columns fx-multi');
    menu.hidden = options.length < 2;
    const summary = element('summary', '', all);
    summary.setAttribute('aria-label', `${text}: ${all}`);
    const list = element('div', 'fx-column-toggles');
    list.setAttribute('role', 'group');
    list.setAttribute('aria-label', text);
    function label() {
        const chosen = options.filter(([value]) => values.has(value)).map(([, name]) => name);
        summary.textContent = chosenLabel(chosen, all, plural);
        summary.title = chosen.join(', ');
        summary.setAttribute('aria-label', `${text}: ${chosen.length ? chosen.join(', ') : all}`);
        menu.classList.toggle('fx-multi-active', chosen.length > 0);
    }
    for (const [value, name, total] of options) {
        const row = element('label', 'fx-column-toggle');
        const box = element('input');
        box.type = 'checkbox';
        box.value = value;
        box.checked = values.has(value);
        box.addEventListener('change', () => {
            if (box.checked)
                values.add(value);
            else
                values.delete(value);
            label();
            change();
        });
        row.append(box, element('span', 'fx-multi-name', name), element('span', 'fx-multi-count', total.toLocaleString()));
        list.append(row);
    }
    menu.append(summary, list);
    dismissible(menu);
    label();
    return { element: menu, values };
}
export function choice(text, options, selected) {
    const label = element('label', 'fx-field');
    label.append(element('span', 'fx-field-text', text));
    const select = element('select');
    for (const [value, name] of options) {
        const option = element('option', '', name);
        option.value = value;
        select.append(option);
    }
    select.value = selected;
    label.append(select);
    return { label, select };
}
/** A select's options: the catch-all first, then each distinct value by its label. */
export function choices(first, values, labels) {
    const unique = [...new Set(values)].map(value => [value, labels?.[value] ?? value]);
    return [['', first], ...unique.sort((a, b) => collator.compare(a[1], b[1]))];
}
export function navigationButtons(label, previous, next) {
    const section = element('nav', 'fx-page-buttons');
    section.setAttribute('aria-label', label);
    const back = element('button', 'button', 'Previous');
    const forward = element('button', 'button', 'Next');
    back.type = forward.type = 'button';
    back.addEventListener('click', previous);
    forward.addEventListener('click', next);
    const pageLabel = element('span', 'fx-note');
    pageLabel.setAttribute('aria-live', 'polite');
    section.append(back, pageLabel, forward);
    return { section, update: (page, pages) => {
            back.disabled = page === 0;
            forward.disabled = page + 1 >= pages;
            pageLabel.textContent = `${count(page + 1)} of ${count(pages)}`;
        } };
}
/** A paged list's state: its buttons, and the page they move, rendered again on each move. */
export function pager(label, render) {
    const state = { page: 0 };
    const buttons = navigationButtons(label, () => { state.page--; render(); }, () => { state.page++; render(); });
    return Object.assign(state, buttons);
}
