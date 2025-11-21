#!/usr/bin/env node

/**
 * Script para descargar archivos de FFmpeg core localmente
 * Ejecutar: node scripts/download-ffmpeg.js
 */

import { writeFile, mkdir } from 'fs/promises';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const FFMPEG_VERSION = '0.12.6';
const BASE_URL = `https://cdn.jsdelivr.net/npm/@ffmpeg/core@${FFMPEG_VERSION}/dist/esm`;

const files = [
  { name: 'ffmpeg-core.js', required: true },
  { name: 'ffmpeg-core.wasm', required: true },
  { name: 'ffmpeg-core.worker.js', required: false } // Opcional
];

const outputDir = join(__dirname, '..', 'static', 'ffmpeg');

async function downloadFile(file) {
  const url = `${BASE_URL}/${file.name}`;
  console.log(`📥 Descargando ${file.name} desde ${url}...`);
  
  try {
    const response = await fetch(url);
    
    if (!response.ok) {
      if (!file.required) {
        console.log(`⚠️  ${file.name} no disponible (opcional, continuando...)`);
        return false;
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const buffer = await response.arrayBuffer();
    const outputPath = join(outputDir, file.name);
    
    await writeFile(outputPath, Buffer.from(buffer));
    console.log(`✅ ${file.name} descargado correctamente (${(buffer.byteLength / 1024).toFixed(2)} KB)`);
    return true;
  } catch (error) {
    console.error(`❌ Error descargando ${file.name}:`, error);
    if (file.required) {
      throw error;
    }
    return false;
  }
}

async function main() {
  console.log('🚀 Descargando archivos de FFmpeg core...');
  console.log(`📍 Directorio de salida: ${outputDir}`);
  
  try {
    // Crear directorio si no existe
    await mkdir(outputDir, { recursive: true });
    
    // Descargar todos los archivos
    for (const file of files) {
      await downloadFile(file);
    }
    
    console.log('\n✅ Todos los archivos descargados correctamente');
    console.log('💡 Los archivos están en: static/ffmpeg/');
  } catch (error) {
    console.error('\n❌ Error durante la descarga:', error);
    process.exit(1);
  }
}

main();
