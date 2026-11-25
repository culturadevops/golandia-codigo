import { writeFile } from "node:fs/promises";

interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

const API_URL = process.env.USERS_API_URL ?? "https://api.example.com/users";
const OUTPUT_FILE = process.env.OUTPUT_FILE ?? "admins.json";

function isUser(value: unknown): value is User {
  if (typeof value !== "object" || value === null) return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.id === "number" &&
    typeof v.name === "string" &&
    typeof v.email === "string" &&
    typeof v.role === "string"
  );
}

async function fetchUsers(url: string, timeoutMs = 10_000): Promise<User[]> {
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
    signal: AbortSignal.timeout(timeoutMs),
  });

  if (!response.ok) {
    throw new Error(`Error HTTP ${response.status}: ${response.statusText}`);
  }

  const data: unknown = await response.json();

  if (!Array.isArray(data)) {
    throw new Error("La respuesta de la API no es un array de usuarios");
  }

  return data.filter(isUser);
}

export async function saveAdmins(
  url: string = API_URL,
  outputPath: string = OUTPUT_FILE,
): Promise<User[]> {
  const users = await fetchUsers(url);
  const admins = users.filter((user) => user.role.toLowerCase() === "admin");

  await writeFile(outputPath, JSON.stringify(admins, null, 2), "utf-8");
  console.log(`Guardados ${admins.length} administradores en ${outputPath}`);

  return admins;
}

saveAdmins().catch((error: unknown) => {
  console.error("Error:", error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
