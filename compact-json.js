export const compactJsonSchema = 'urn:repometry:schema:compact-json:1';
export function packJson(value) {
    const strings = [];
    const shapes = [];
    const stringIds = new Map();
    const shapeIds = new Map();
    function pack(item, depth) {
        if (depth > 512)
            throw new Error('Compact JSON exceeds the nesting limit');
        if (item === null || item === undefined)
            return null;
        if (typeof item === 'boolean')
            return item;
        if (typeof item === 'number')
            return Number.isFinite(item) ? item : null;
        if (typeof item === 'string') {
            let index = stringIds.get(item);
            if (index === undefined) {
                index = strings.length;
                strings.push(item);
                stringIds.set(item, index);
            }
            return [0, index];
        }
        if (Array.isArray(item))
            return [1, ...item.map(value => pack(value, depth + 1))];
        if (typeof item !== 'object')
            throw new Error('Unsupported value in compact JSON');
        const entries = Object.entries(item).filter(([, value]) => value !== undefined);
        const keys = entries.map(([key]) => key);
        const signature = JSON.stringify(keys);
        let index = shapeIds.get(signature);
        if (index === undefined) {
            index = shapes.length;
            shapes.push(keys);
            shapeIds.set(signature, index);
        }
        return [index + 2, ...entries.map(([, value]) => pack(value, depth + 1))];
    }
    const data = pack(value, 0);
    return { schema: compactJsonSchema, strings, shapes, data };
}
export function unpackJson(value) {
    if (typeof value !== 'object' || value === null || !('schema' in value) || value.schema !== compactJsonSchema)
        return value;
    const candidate = value;
    if (!Array.isArray(candidate.strings) || !candidate.strings.every(item => typeof item === 'string')
        || !Array.isArray(candidate.shapes) || !candidate.shapes.every(shape => Array.isArray(shape)
        && shape.every(key => typeof key === 'string') && new Set(shape).size === shape.length)) {
        throw new Error('Invalid compact JSON dictionary');
    }
    const { strings, shapes } = candidate;
    function unpack(item, depth) {
        if (depth > 512)
            throw new Error('Compact JSON exceeds the nesting limit');
        if (item === null || typeof item === 'boolean' || typeof item === 'number')
            return item;
        if (!Array.isArray(item) || !Number.isSafeInteger(item[0]))
            throw new Error('Invalid compact JSON node');
        const [tag, ...values] = item;
        if (tag === 0) {
            const id = values[0];
            if (values.length !== 1 || typeof id !== 'number' || !Number.isSafeInteger(id) || id < 0 || id >= strings.length) {
                throw new Error('Invalid compact JSON string reference');
            }
            return strings[id];
        }
        if (tag === 1)
            return values.map(child => unpack(child, depth + 1));
        const keys = tag >= 2 ? shapes[tag - 2] : undefined;
        if (!keys || keys.length !== values.length)
            throw new Error('Invalid compact JSON object reference');
        return Object.fromEntries(keys.map((key, index) => [key, unpack(values[index], depth + 1)]));
    }
    return unpack(candidate.data, 0);
}
