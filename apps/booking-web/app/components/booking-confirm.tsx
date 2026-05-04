"use client";

import { useRef, useState } from "react";

interface BookingConfirmLabels {
  confirm: string;
  cancel: string;
  specialist: string;
  service: string;
  time: string;
  price: string;
}

export interface BookingConfirmButtonProps {
  timeLabel: string;
  specialistName: string;
  serviceName: string;
  price: string;
  tenantSlug: string;
  serviceMemberId: string;
  startTime: string;
  returnTo: string;
  primaryColor: string;
  primaryTextColor: string;
  accentColor: string;
  labels: BookingConfirmLabels;
}

export function BookingConfirmButton({
  timeLabel,
  specialistName,
  serviceName,
  price,
  tenantSlug,
  serviceMemberId,
  startTime,
  returnTo,
  primaryColor,
  primaryTextColor,
  accentColor,
  labels,
}: BookingConfirmButtonProps) {
  const [open, setOpen] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="w-full min-h-11 rounded-xl border px-3 py-3 text-center text-sm font-medium transition hover:opacity-90 active:scale-[0.97]"
        style={{
          borderColor: accentColor + "55",
          backgroundColor: accentColor + "12",
          color: primaryColor,
        }}
      >
        {timeLabel}
      </button>

      {open ? (
        <div
          className="fixed inset-0 z-50 flex items-end sm:items-center justify-center sm:p-4"
          style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
          role="dialog"
          aria-modal="true"
          onClick={(e) => {
            if (e.target === e.currentTarget) setOpen(false);
          }}
        >
          <div className="w-full rounded-t-3xl sm:max-w-sm sm:rounded-3xl bg-white p-6 shadow-2xl">
            <div
              className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-2xl"
              style={{ backgroundColor: primaryColor + "18" }}
            >
              <svg
                className="h-6 w-6"
                style={{ color: primaryColor }}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
            </div>
            <h2 className="mb-5 text-center text-lg font-semibold text-gray-900">
              {labels.confirm}
            </h2>
            <dl className="mb-6 divide-y divide-gray-100">
              {(
                [
                  [labels.specialist, specialistName],
                  [labels.service, serviceName],
                  [labels.time, timeLabel],
                  [labels.price, price],
                ] as [string, string][]
              ).map(([label, value]) => (
                <div key={label} className="flex items-center justify-between gap-4 py-2.5">
                  <dt className="text-sm text-gray-500">{label}</dt>
                  <dd className="text-sm font-semibold text-gray-900">{value}</dd>
                </div>
              ))}
            </dl>
            <form ref={formRef} action="/book" method="POST" className="flex flex-col gap-3">
              <input type="hidden" name="tenantSlug" value={tenantSlug} />
              <input type="hidden" name="serviceMemberId" value={serviceMemberId} />
              <input type="hidden" name="startTime" value={startTime} />
              <input type="hidden" name="returnTo" value={returnTo} />
              <button
                type="submit"
                className="min-h-11 w-full rounded-xl px-4 py-3 text-sm font-semibold transition hover:opacity-90 active:scale-[0.98]"
                style={{ backgroundColor: primaryColor, color: primaryTextColor }}
              >
                {labels.confirm}
              </button>
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="min-h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm font-medium text-gray-700 transition hover:bg-gray-100 active:scale-[0.98]"
              >
                {labels.cancel}
              </button>
            </form>
          </div>
        </div>
      ) : null}
    </>
  );
}
