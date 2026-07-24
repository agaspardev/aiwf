package assets

// ReadFile devuelve el contenido de un archivo embebido por su ruta de deploy
// (relativa a la raíz de instalación), p. ej. "harness/modes.json".
func ReadFile(deployPath string) ([]byte, error) {
	return files.ReadFile("files/" + deployPath)
}
