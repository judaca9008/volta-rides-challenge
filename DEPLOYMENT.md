# Deployment Guide - Fly.io

Este proyecto está configurado para deployment rápido en Fly.io con HTTPS automático y URL pública.

## 🚀 Quick Deploy (3 comandos)

### Paso 1: Instalar Fly CLI

**macOS:**
```bash
brew install flyctl
```

**Linux/WSL:**
```bash
curl -L https://fly.io/install.sh | sh
```

**Windows (PowerShell):**
```powershell
iwr https://fly.io/install.ps1 -useb | iex
```

### Paso 2: Autenticarse

```bash
fly auth signup
# O si ya tienes cuenta:
fly auth login
```

### Paso 3: Deploy

```bash
# En el directorio del proyecto
fly launch --now

# Fly.io te preguntará:
# - App name: [presiona Enter para usar "volta-router" o escribe otro]
# - Region: [presiona Enter para usar Miami/mia - óptimo para LATAM]
# - PostgreSQL/Redis: [presiona "n" - no necesitamos por ahora]
```

**¡Listo!** Tu app estará en: `https://volta-router.fly.dev`

---

## 📊 Post-Deployment

### Cargar Test Data

```bash
# Obtén tu URL
export APP_URL=$(fly status --json | jq -r '.Hostname')

# Carga test data
curl -X POST https://$APP_URL/volta-router/v1/transactions/load

# Verifica que funciona
curl https://$APP_URL/volta-router/v1/processors | jq
```

### Test Routing

```bash
# Test routing decision para Brasil
curl -X POST https://$APP_URL/volta-router/v1/route \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}' | jq

# Test con failover ranking
curl -X POST "https://$APP_URL/volta-router/v1/route?failover=true" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100, "currency": "BRL", "country": "BR"}' | jq
```

---

## 🔧 Comandos Útiles

### Ver logs en tiempo real
```bash
fly logs
```

### Ver status de la app
```bash
fly status
```

### Abrir la app en el navegador
```bash
fly open
```

### Ver dashboard de monitoreo
```bash
fly dashboard
```

### Escalar la app
```bash
# Aumentar memoria
fly scale memory 512

# Agregar más instancias
fly scale count 2
```

### SSH a la instancia
```bash
fly ssh console
```

---

## 🔄 Re-deployment (Después de cambios)

```bash
# Deploy nueva versión
fly deploy

# O con nombre específico
fly deploy --app volta-router
```

---

## 🌍 URL Pública

Tu API estará disponible en:
- **Base URL**: `https://volta-router.fly.dev`
- **Health**: `https://volta-router.fly.dev/health`
- **API Docs**: Ver README.md para endpoints completos

### Endpoints Públicos

```bash
# Health check
https://volta-router.fly.dev/health

# Cargar test data
POST https://volta-router.fly.dev/volta-router/v1/transactions/load

# Routing decision
POST https://volta-router.fly.dev/volta-router/v1/route

# Processor health
GET https://volta-router.fly.dev/volta-router/v1/processors

# Routing stats
GET https://volta-router.fly.dev/volta-router/v1/routing/stats
```

---

## 💰 Costos

**Free Tier incluye:**
- ✅ 3 apps shared-cpu-1x (256MB RAM)
- ✅ 160GB bandwidth
- ✅ HTTPS automático
- ✅ Auto-scaling
- ✅ Monitoreo incluido

**Esta app usa:**
- 1 shared CPU
- 256MB RAM
- Auto-stop cuando no hay tráfico (gratis)
- Auto-start cuando llega request

**Costo mensual: $0** (dentro del free tier)

---

## 🐛 Troubleshooting

### App no responde
```bash
# Ver logs
fly logs

# Reiniciar
fly apps restart
```

### Deploy falla
```bash
# Ver detalles del error
fly deploy --verbose

# Verificar Dockerfile localmente
docker build -t volta-router .
docker run -p 8080:8080 volta-router
```

### Cambiar región
```bash
# Ver regiones disponibles
fly platform regions

# Mover a otra región (ej: sao = São Paulo)
fly regions set sao
```

---

## 🔐 Variables de Entorno

Si necesitas agregar env vars:

```bash
fly secrets set CUSTOM_VAR=value
```

---

## 🗑️ Eliminar la App

```bash
fly apps destroy volta-router
```

---

## 📚 Documentación Adicional

- [Fly.io Docs](https://fly.io/docs/)
- [Go on Fly.io](https://fly.io/docs/languages-and-frameworks/golang/)
- [Fly.io Pricing](https://fly.io/docs/about/pricing/)

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
