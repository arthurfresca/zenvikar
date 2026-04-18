type FetchApiInit = RequestInit & {
  path: string;
};

function getServerApiBaseUrls(): string[] {
  const candidates = [
    process.env.API_INTERNAL_URL,
    process.env.API_PUBLIC_URL,
    process.env.NEXT_PUBLIC_API_URL,
    "http://host.docker.internal:8081",
  ];

  return [...new Set(candidates.map((value) => value?.trim()).filter(Boolean) as string[])];
}

export async function fetchServerApi({ path, ...init }: FetchApiInit): Promise<Response> {
  const bases = getServerApiBaseUrls();
  let lastError: unknown;

  for (const base of bases) {
    try {
      return await fetch(new URL(path, base).toString(), init);
    } catch (error) {
      lastError = error;
    }
  }

  throw lastError instanceof Error
    ? lastError
    : new Error("Failed to reach API from booking-web server");
}
