# 🔐 GophKeeper

> Менеджер паролей и секретов с синхронизацией между устройствами

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

---

## 🚀 Быстрый старт

### Шаг 1. Клонирование репозитория

```bash
git clone https://github.com/mdflamingo/GophKeeper.git
cd GophKeeper
```

### Шаг 2. Сборка клиента

```bash
chmod +x build.sh && ./build.sh
```
После выполнения в директории `dist/` появятся готовые бинарные файлы для всех платформ.

### Шаг 3. Запуск

Выберите файл, соответствующий вашей операционной системе:
### 🍎 macOS (Intel)
```bash
chmod +x ./dist/gophkeeper_darwin_amd64
./dist/gophkeeper_darwin_amd64
```

### 🍎 macOS (Apple Silicon — M1/M2/M3)
```bash
chmod +x ./dist/gophkeeper_darwin_arm64
./dist/gophkeeper_darwin_arm64
```

### 🐧 Linux
```bash
chmod +x ./dist/gophkeeper_linux_amd64
./dist/gophkeeper_linux_amd64
```

### 🪟 Windows (запуск из командной строки)
```bash
dist\gophkeeper_windows_amd64.exe
```