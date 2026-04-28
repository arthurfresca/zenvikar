"use client";

import { useState } from "react";

type Props = {
  defaultValueCents: number;
  currency: string;
  locale: string;
};

export function CurrencyInput({ defaultValueCents, currency, locale }: Props) {
  const intlLocale = locale === "pt" ? "pt-BR" : "en-US";
  const symbol = new Intl.NumberFormat(intlLocale, { style: "currency", currency, maximumFractionDigits: 0 })
    .format(0)
    .replace(/[\d\s.,]/g, "")
    .trim();

  const [display, setDisplay] = useState((defaultValueCents / 100).toFixed(2));
  const cents = Math.round(parseFloat(display || "0") * 100) || 0;

  return (
    <div className="relative">
      <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-sm text-gray-400 dark:text-slate-500">
        {symbol}
      </span>
      <input
        type="number"
        step="0.01"
        min="0"
        value={display}
        onChange={(e) => setDisplay(e.target.value)}
        className="w-full rounded-xl border border-gray-300 bg-white py-2.5 pl-8 pr-4 text-sm text-gray-900 outline-none transition placeholder:text-gray-400 focus:border-gray-500 focus:ring-2 focus:ring-gray-100 dark:border-white/10 dark:bg-slate-950 dark:text-white dark:focus:border-white/25 dark:focus:ring-white/5"
      />
      <input type="hidden" name="priceCents" value={isNaN(cents) ? "0" : cents.toString()} />
    </div>
  );
}
