import { writeFile } from "node:fs/promises";

interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

const API_URL = process.env.USERS_API_URL ?? "https://api.example.com/users";
const OUTPUT_FILE = process.env.OUTPUT_FILE ?? "admins.json";

async function fetchUsers(url: string): Promise<User[]> {
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
    signal: AbortSignal.timeout(10_000),
  });

  if (!response.ok) {
    throw new Error(`Error HTTP ${response.status}: ${response.statusText}`);
  }

  const data: unknown = await response.json();

  if (!Array.isArray(data)) {
    throw new Error("La respuesta de la API no es un array de usuarios");
  }

  return data as User[];
}

function filterAdmins(users: User[]): User[] {
  return users.filter((user) => user.role?.toLowerCase() === "admin");
}

async function saveAdmins(url: string, outputPath: string): Promise<void> {
  const users = await fetchUsers(url);
  const admins = filterAdmins(users);

  await writeFile(outputPath, JSON.stringify(admins, null, 2), "utf-8");

  console.log(`Guardados ${admins.length} de ${users.length} usuarios en ${outputPath}`);
}

saveAdmins(API_URL, OUTPUT_FILE).catch((error: unknown) => {
  console.error("Error:", error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
