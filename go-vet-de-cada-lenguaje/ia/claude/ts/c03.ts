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
    throw new Error("La respuesta de la API no es un arreglo de usuarios");
  }

  return data as User[];
}

function filterAdmins(users: User[]): User[] {
  return users.filter((user) => user.role?.toLowerCase() === "admin");
}

async function saveAdmins(admins: User[], path: string): Promise<void> {
  await writeFile(path, JSON.stringify(admins, null, 2), "utf-8");
}

async function main(): Promise<void> {
  try {
    const users = await fetchUsers(API_URL);
    const admins = filterAdmins(users);
    await saveAdmins(admins, OUTPUT_FILE);
    console.log(`Guardados ${admins.length} de ${users.length} usuarios en ${OUTPUT_FILE}`);
  } catch (error) {
    console.error("Fallo al procesar usuarios:", error instanceof Error ? error.message : error);
    process.exitCode = 1;
  }
}

await main();
