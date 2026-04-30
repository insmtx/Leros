import type { StateCreator } from "zustand";

export type PublicActions<T> = { [K in keyof T]: T[K] };

export type SliceCreator<T> = StateCreator<T, [], [], T>;
