"use client";

export function LogoutButton() {
  return (
    <a
      href="/logout"
      className="rounded border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
    >
      Log Out
    </a>
  );
}
