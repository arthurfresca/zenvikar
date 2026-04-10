import { defaultLocale } from "@zenvikar/config";

export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <h1 className="text-4xl font-bold mb-4">Zenvikar</h1>
      <p className="text-lg text-gray-600 mb-8">
        Multi-tenant booking platform for service businesses.
      </p>
      <p className="text-sm text-gray-400">Locale: {defaultLocale}</p>
    </main>
  );
}
