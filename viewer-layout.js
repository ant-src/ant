import { element, link, store, stored } from './viewer-dom.js';
const asideKey = 'repometry.aside';
/**
 * The handle between a list and its side panel: a line that shows on hover and drags the panel wider or narrower.
 * Arrow keys move it too, a double click restores the default, and the width is remembered for this viewer.
 */
export function splitter(split) {
    const handle = element('div', 'fx-splitter');
    handle.tabIndex = 0;
    handle.setAttribute('role', 'separator');
    handle.setAttribute('aria-orientation', 'vertical');
    handle.setAttribute('aria-label', 'Resize the side panel');
    handle.title = 'Drag to resize; double-click to reset';
    const apply = (width) => {
        if (width === null) {
            split.style.removeProperty('--aside');
            return;
        }
        const clamped = Math.round(Math.min(Math.max(224, split.clientWidth * .55), Math.max(224, width)));
        split.style.setProperty('--aside', `${clamped}px`);
        handle.setAttribute('aria-valuenow', String(clamped));
    };
    const save = (width) => store(() => localStorage, asideKey, width === null ? null : String(Math.round(width)));
    const current = () => split.lastElementChild instanceof HTMLElement && split.lastElementChild !== handle
        ? split.lastElementChild.getBoundingClientRect().width : 320;
    const remembered = Number(stored(() => localStorage, asideKey));
    if (remembered > 0)
        requestAnimationFrame(() => apply(remembered));
    handle.addEventListener('pointerdown', event => {
        event.preventDefault();
        handle.setPointerCapture(event.pointerId);
        split.classList.add('fx-resizing');
        const right = split.getBoundingClientRect().right;
        const move = (next) => apply(right - next.clientX - handle.offsetWidth / 2);
        const stop = () => {
            split.classList.remove('fx-resizing');
            save(current());
            handle.removeEventListener('pointermove', move);
            handle.removeEventListener('pointerup', stop);
            handle.removeEventListener('pointercancel', stop);
        };
        handle.addEventListener('pointermove', move);
        handle.addEventListener('pointerup', stop);
        handle.addEventListener('pointercancel', stop);
    });
    handle.addEventListener('keydown', event => {
        const step = arrowSteps.get(event.key);
        if (step === undefined)
            return;
        event.preventDefault();
        apply(current() + step);
        save(current());
    });
    handle.addEventListener('dblclick', () => { apply(null); save(null); });
    liftToTitle(split, ':scope > .fx-aside h2');
    return handle;
}
/** The side panel sits at the right: the left arrow widens it. */
const arrowSteps = new Map([['ArrowLeft', 24], ['ArrowRight', -24]]);
/** Where a text's baseline is: its box ends below the baseline by the descent, about a fifth of its height. */
function baseline(node) {
    const range = document.createRange();
    range.selectNodeContents(node);
    const box = range.getBoundingClientRect();
    return box.bottom - box.height * .2;
}
/** The child of `parent` that holds `node`. */
function columnOf(node, parent) {
    let column = node;
    while (column.parentElement && column.parentElement !== parent)
        column = column.parentElement;
    return column;
}
/**
 * Raises a side column so its heading stands on the page title's baseline: the list and its summary read as one
 * row of headings. What sits between the title and the panel (a folder path, a toolbar) varies, so the distance is
 * measured, and measured again whenever the page above the panel changes height. One column on a narrow screen: no lift.
 */
export function liftToTitle(split, headingSelector) {
    const measure = () => {
        const title = split.closest('section')?.querySelector(':scope > h2');
        const heading = split.querySelector(headingSelector);
        const columns = getComputedStyle(split).gridTemplateColumns.split(' ').length;
        if (!split.isConnected || !title || !heading || columns < 2) {
            split.style.removeProperty('--lift');
            return;
        }
        // The column may be sticky, and once the page scrolls its box no longer sits where the layout put it. So the
        // heading is measured inside its column, and the column's place is taken from the grid, which never sticks:
        // unlifted, a first-row column starts at the top of the grid's content box.
        const style = getComputedStyle(split);
        const top = split.getBoundingClientRect().top + parseFloat(style.borderTopWidth) + parseFloat(style.paddingTop);
        const offset = baseline(heading) - columnOf(heading, split).getBoundingClientRect().top;
        const current = parseFloat(split.style.getPropertyValue('--lift')) || 0;
        const next = Math.max(0, Math.round(top + offset - baseline(title)));
        if (next !== current)
            split.style.setProperty('--lift', `${next}px`);
    };
    const observer = new ResizeObserver(() => {
        if (split.isConnected)
            measure();
        else
            observer.disconnect();
    });
    requestAnimationFrame(() => {
        measure();
        // The page section, not the view: a view may be laid out through display: contents and never resize itself.
        const page = split.closest('section') ?? split.parentElement;
        if (page)
            observer.observe(page);
    });
}
/** The folder trail above a list: each folder a link, the current one a label, nothing at the root. */
export function breadcrumbs(view, navigation, directory) {
    const nav = element('nav', 'fx-breadcrumbs');
    nav.setAttribute('aria-label', 'Folder path');
    const list = element('ol');
    const segments = directory === '' ? [] : directory.split('/');
    const crumb = (label, target) => {
        const item = element('li');
        if (target === null) {
            item.setAttribute('aria-current', 'location');
            item.textContent = label;
        }
        else
            item.append(link(navigation.href(view, target ? { directory: target } : undefined), label));
        return item;
    };
    list.append(crumb('repo', directory === '' ? null : ''), ...segments.map((part, index) => crumb(part, index === segments.length - 1 ? null : segments.slice(0, index + 1).join('/'))));
    nav.append(list);
    // At the root the trail is a single label that says nothing the page title does not.
    nav.hidden = directory === '';
    return nav;
}
