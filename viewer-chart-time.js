import { clampLabel } from './viewer-chart-svg.js';
const monthLabel = new Intl.DateTimeFormat(undefined, { month: 'short' });
const monthYearLabel = new Intl.DateTimeFormat(undefined, { month: 'short', year: 'numeric' });
const dayLabel = new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short' });
const dayYearLabel = new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short', year: 'numeric' });
const hourLabel = new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' });
function changes(previous, date) {
    const year = previous === null || previous.getFullYear() !== date.getFullYear();
    const month = year || previous.getMonth() !== date.getMonth();
    return { year, month, day: month || previous.getDate() !== date.getDate() };
}
/** A tick within a day names its time, and the date where the date changes; midnight is the date alone. */
function clockLabel(date, changed) {
    if (!changed.day)
        return hourLabel.format(date);
    const days = (changed.year ? dayYearLabel : dayLabel).format(date);
    return date.getHours() === 0 && date.getMinutes() === 0 ? days : `${days}, ${hourLabel.format(date)}`;
}
function dateLabel(date, changed) {
    if (changed.year)
        return dayYearLabel.format(date);
    return changed.month ? dayLabel.format(date) : String(date.getDate());
}
const unitLabels = {
    year: date => String(date.getFullYear()),
    month: (date, changed) => (changed.year ? monthYearLabel : monthLabel).format(date),
    day: dateLabel,
    hour: clockLabel,
    minute: clockLabel,
};
/**
 * Labels for ticks in time order. Each names its own unit, and the coarser units join it only where they change:
 * Jul 2023, Nov, Dec, Jan 2024, Feb; 30 Jun, 1 Jul, 2; 23:00, 1 Jul, 01:00. The first tick always carries them.
 */
export function tickLabels(times, unit) {
    let previous = null;
    return times.map(time => {
        const date = new Date(time);
        const changed = changes(previous, date);
        previous = date;
        return unitLabels[unit](date, changed);
    });
}
/**
 * The labels under the ticks. Which ticks keep a label is settled first: one that would sit far off its tick at the
 * edge, or touch the label before it, keeps only its mark. Only then are the labels written, so the first label shown
 * carries the coarser units, and every later one carries them only where they change.
 */
export function axisLabels(ticks, unit, layout) {
    let shown = [...ticks];
    for (;;) {
        const labels = tickLabels(shown.map(tick => tick.time), unit);
        let previous = -Infinity;
        const placed = [];
        const drop = shown.findIndex((tick, index) => {
            const label = labels[index];
            const at = clampLabel(tick.x, label, layout);
            const half = label.length * 3.7;
            if (Math.abs(at - tick.x) > 12 || at - half < previous + 10)
                return true;
            previous = at + half;
            placed.push({ at, label });
            return false;
        });
        if (drop < 0)
            return placed;
        shown = shown.filter((_, index) => index !== drop);
    }
}
const minute = 60e3;
const hour = 60 * minute;
const day = 24 * hour;
const steps = [
    { unit: 'minute', count: 5, size: 5 * minute }, { unit: 'minute', count: 10, size: 10 * minute },
    { unit: 'minute', count: 15, size: 15 * minute }, { unit: 'minute', count: 30, size: 30 * minute },
    { unit: 'hour', count: 1, size: hour }, { unit: 'hour', count: 3, size: 3 * hour }, { unit: 'hour', count: 6, size: 6 * hour },
    { unit: 'hour', count: 12, size: 12 * hour }, { unit: 'day', count: 1, size: day }, { unit: 'day', count: 2, size: 2 * day },
    { unit: 'day', count: 7, size: 7 * day }, { unit: 'day', count: 14, size: 14 * day }, { unit: 'month', count: 1, size: 30.4 * day },
    { unit: 'month', count: 3, size: 91 * day }, { unit: 'month', count: 6, size: 182 * day }, { unit: 'year', count: 1, size: 365 * day },
    { unit: 'year', count: 2, size: 730 * day }, { unit: 'year', count: 5, size: 1826 * day },
];
/** Calendar-aligned ticks between two times, as many as fit the width without overlapping labels. */
export function timeTicks(start, end, width) {
    for (const step of steps) {
        const periods = (end - start) / step.size;
        if (periods > Math.max(2, width / 64))
            continue;
        const ticks = calendarTicks(start, end, step.unit, step.count);
        // Sized by the typical label: the few that carry a year or a date beside them are placed with care by the axis.
        const lengths = ticks.map(tick => tick.label.length).sort((x, y) => x - y);
        const widest = (lengths[Math.floor(lengths.length / 2)] ?? 0) * 7.4 + 14;
        if (ticks.length < 2 || width / Math.max(1, periods) >= widest)
            return ticks;
    }
    return calendarTicks(start, end, 'year', 10);
}
/** How coarse each unit is: a tick of a unit starts at zero in every finer one. */
const coarseness = { minute: 0, hour: 1, day: 2, month: 3, year: 4 };
function startOf(time, unit) {
    const cursor = new Date(time);
    cursor.setSeconds(0, 0);
    const level = coarseness[unit];
    if (level >= 1)
        cursor.setMinutes(0);
    if (level >= 2)
        cursor.setHours(0);
    if (level >= 3)
        cursor.setDate(1);
    if (level >= 4)
        cursor.setMonth(0);
    return cursor;
}
const advances = {
    minute: cursor => cursor.setMinutes(cursor.getMinutes() + 1),
    hour: cursor => cursor.setHours(cursor.getHours() + 1),
    day: cursor => cursor.setDate(cursor.getDate() + 1),
    month: cursor => cursor.setMonth(cursor.getMonth() + 1),
    year: cursor => cursor.setFullYear(cursor.getFullYear() + 1),
};
function weekOf(date) {
    return Math.floor(date.getTime() / (7 * day));
}
/** Weeks start on Mondays, every other one for two-week steps; days step from the first of the month, never to the 31st. */
function alignedDay(cursor, count) {
    if (count !== 7 && count !== 14)
        return (cursor.getDate() - 1) % count === 0 && cursor.getDate() < 31;
    return cursor.getDay() === 1 && (count === 7 || weekOf(cursor) % 2 === 0);
}
const alignments = {
    minute: (cursor, count) => cursor.getMinutes() % count === 0,
    hour: (cursor, count) => cursor.getHours() % count === 0,
    day: alignedDay,
    month: (cursor, count) => cursor.getMonth() % count === 0,
    year: (cursor, count) => cursor.getFullYear() % count === 0,
};
function calendarTicks(start, end, unit, count) {
    const cursor = startOf(start, unit);
    const advance = advances[unit];
    const times = [];
    let guard = 0;
    while (cursor.getTime() < start && guard++ < 6000)
        advance(cursor);
    while (cursor.getTime() <= end && guard++ < 6000) {
        if (alignments[unit](cursor, count))
            times.push(cursor.getTime());
        advance(cursor);
    }
    const labels = tickLabels(times, unit);
    return times.map((time, index) => ({ time, label: labels[index], unit }));
}
