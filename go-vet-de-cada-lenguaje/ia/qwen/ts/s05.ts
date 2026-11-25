import fs from 'fs/promises';
import { URL } from 'url';

async function descargarAdminsAPI(apiUrl: string, outputPath: string): Promise<void> {
    try {
        const response = await fetch(apiUrl);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const usuarios = await response.json();
        const admins = usuarios.filter(usuario => usuario.rol === 'admin');

        await fs.writeFile(outputPath, JSON.stringify(admins, null, 2));
        console.log(`Se han guardado los usuarios admin en ${outputPath}`);
    } catch (error) {
        console.error('Hubo un error:', error);
    }
}

// Ejecución del ejemplo
// descargarAdminsAPI(new URL('https://api.example.com/usuarios').href, './admins.json');
