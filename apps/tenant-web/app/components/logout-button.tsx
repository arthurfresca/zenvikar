"use client";

import { useRouter } from "next/navigation";
import { clearTenantToken } from "@/lib/auth";

export function LogoutButton() {
  const router = useRouter();

  return (
    <button
      type="button"
      className="rounded border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100"
      onClick={() => {
        clearTenantToken();
        router.push("/login");
      }}
    >
      Log Out
    </button>
  );
}
