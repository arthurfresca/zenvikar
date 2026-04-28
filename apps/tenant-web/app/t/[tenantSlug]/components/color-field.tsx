"use client";

import { useState } from "react";

export function ColorField({ name, defaultValue }: { name: string; defaultValue: string }) {
  const [value, setValue] = useState(defaultValue || "#000000");

  return (
    <div className="group relative inline-flex">
      {/* Color swatch */}
      <div
        className="h-11 w-11 rounded-xl border-2 border-white shadow-sm ring-1 ring-gray-200 dark:border-slate-800 dark:ring-white/15"
        style={{ backgroundColor: value }}
      />
      {/* Native color picker overlaid invisibly so clicking the swatch opens it */}
      <input
        type="color"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
        aria-label={name}
      />
      {/* Hidden form field */}
      <input type="hidden" name={name} value={value} />
      {/* Hex tooltip on hover */}
      <div className="pointer-events-none absolute bottom-full left-1/2 mb-2 -translate-x-1/2 whitespace-nowrap rounded-lg bg-gray-900 px-2.5 py-1 font-mono text-xs text-white opacity-0 transition-opacity group-hover:opacity-100 dark:bg-zinc-700">
        {value}
        <div className="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900 dark:border-t-zinc-700" />
      </div>
    </div>
  );
}
