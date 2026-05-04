"use client";

import { useEffect } from "react";

interface NotificationsProps {
  booking: string | undefined;
  bookingError: string | undefined;
  accentColor: string;
  primaryColor: string;
  successText: string;
  genericErrorText: string;
  slotTakenText: string;
  serverErrorText: string;
}

export function Notifications({
  booking,
  bookingError,
  accentColor,
  primaryColor,
  successText,
  genericErrorText,
  slotTakenText,
  serverErrorText,
}: NotificationsProps) {
  useEffect(() => {
    if (booking || bookingError) {
      const url = new URL(window.location.href);
      url.searchParams.delete("booking");
      url.searchParams.delete("bookingError");
      window.history.replaceState({}, "", url.pathname + url.search);
    }
  }, []);

  if (!booking && !bookingError) return null;

  if (booking === "created") {
    return (
      <div
        className="flex items-center gap-3 rounded-2xl border p-4 text-sm font-medium"
        style={{
          borderColor: accentColor + "40",
          backgroundColor: accentColor + "12",
          color: primaryColor,
        }}
      >
        <svg
          className="h-5 w-5 flex-shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        {successText}
      </div>
    );
  }

  const errorText =
    bookingError === "slot_taken"
      ? slotTakenText
      : bookingError === "server_error"
      ? serverErrorText
      : genericErrorText;

  return (
    <div className="flex items-center gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-medium text-red-700">
      <svg
        className="h-5 w-5 flex-shrink-0"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={2}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
      {errorText}
    </div>
  );
}
