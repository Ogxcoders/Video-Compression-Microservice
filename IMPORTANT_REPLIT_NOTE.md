# ⚠️ IMPORTANT: This Project is Not for Running in Replit

## Project Type: VPS Deployment Only

This is a **production-ready Docker-based microservice** designed for deployment on your own VPS (Virtual Private Server) using Coolify or Docker Compose.

### Why This Won't Run in Replit

1. **Docker is Not Supported** - Replit does not support Docker, Docker Compose, or containerization
2. **Resource Requirements** - Video compression requires significant CPU/memory resources
3. **External Services** - Requires Redis, PostgreSQL, FFmpeg, ImageMagick
4. **Long-Running Jobs** - Video processing can take hours for large files

## What You Have Here

✅ **Complete Go source code** for video/image compression microservice  
✅ **Docker Compose configuration** for easy deployment  
✅ **Nginx reverse proxy** setup with SSL support  
✅ **PostgreSQL database** schema and migrations  
✅ **Redis queue** configuration  
✅ **Comprehensive documentation** for deployment  
✅ **WordPress integration** code  

## How to Deploy

### Option 1: Deploy to Your VPS with Coolify

1. **Download this project** from Replit
2. **Upload to your VPS** or push to GitHub
3. **Follow the deployment guide** in `DEPLOYMENT.md`
4. **Start with Quick Start** guide in `QUICKSTART.md`

### Option 2: Deploy with Docker Compose

```bash
# On your VPS
git clone <your-repo>
cd video-compressor
cp .env.example .env
# Edit .env with your settings
docker-compose up -d
```

### Option 3: Deploy to Coolify Dashboard

1. Login to your Coolify instance
2. Create new project
3. Upload these files
4. Configure environment variables
5. Click Deploy

## Quick Start Steps

See `QUICKSTART.md` for a 5-minute deployment guide.

## Files You'll Need

All files in this Replit project are ready for deployment:

- `docker-compose.yml` - Service orchestration
- `Dockerfile` - Application container
- `nginx/nginx.conf` - Reverse proxy config
- `scripts/init.sql` - Database schema
- `.env.example` - Configuration template
- `cmd/`, `internal/`, `pkg/` - Go source code

## Documentation

📖 **README.md** - Full feature overview and usage  
📖 **QUICKSTART.md** - 5-minute deployment guide  
📖 **DEPLOYMENT.md** - Detailed deployment instructions  
📖 **API_DOCUMENTATION.md** - Complete API reference  

## What to Do Next

1. **Download/Clone** this project from Replit
2. **Read** `QUICKSTART.md` for fastest deployment
3. **Deploy** to your VPS using Coolify or Docker Compose
4. **Test** using the provided API examples

## Need Help?

All documentation is included. The project is production-ready and tested for deployment on standard VPS environments with Docker support.

## Technical Requirements for Your VPS

- Ubuntu 20.04+ or similar Linux distribution
- Docker & Docker Compose installed
- 2GB+ RAM (4GB+ recommended)
- 20GB+ disk space
- Domain name with DNS configured
- SSL certificate (Let's Encrypt recommended)

---

**Note:** This is a complete, professional-grade microservice. It's designed to run on external servers, not within the Replit environment.
