export function flattenActions<T extends Record<string, unknown>>(
  actions: T[],
): T {
  const result = {} as T;

  for (const instance of actions) {
    const prototype = Object.getPrototypeOf(instance);

    const methodNames: string[] = [];
    let current = prototype;
    while (current && current !== Object.prototype) {
      methodNames.push(
        ...Object.getOwnPropertyNames(current).filter(
          (name) => name !== 'constructor',
        ),
      );
      current = Object.getPrototypeOf(current);
    }

    for (const name of methodNames) {
      const descriptor = Object.getOwnPropertyDescriptor(prototype, name);
      if (descriptor && typeof descriptor.value === 'function') {
        (result as Record<string, unknown>)[name] =
          descriptor.value.bind(instance);
      }
    }

    for (const [key, value] of Object.entries(instance)) {
      if (typeof value === 'function') {
        (result as Record<string, unknown>)[key] = value.bind(instance);
      } else {
        (result as Record<string, unknown>)[key] = value;
      }
    }
  }

  return result;
}
