# COOLIFY DEPLOYMENT FIX - CRITICAL

## THE PROBLEM

Your Coolify deployment is using **Dockerfile mode** which only deploys the app container.
Redis and PostgreSQL are NOT being deployed, so environment variables don't work.

## THE SOLUTION

You MUST change Coolify to use **Docker Compose** mode.

---

## STEPS TO FIX IN COOLIFY

### Step 1: Change Build Pack to Docker Compose

1. Go to your application in Coolify
2. Click on **"General"** or **"Configuration"** tab
3. Find **"Build Pack"** setting
4. Change from **"Dockerfile"** to **"Docker Compose"**
5. Set **"Docker Compose File"** to: `docker-compose.coolify.yml`

### Step 2: Add Environment Variables

In Coolify's **Environment Variables** section, add these variables:

**IMPORTANT:** Do NOT check "Build Variable?" for any of these!

```
API_KEY=sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1
ALLOWED_DOMAINS=https://capcut.ogtemplate.com/,https://ogtemplate.com/
DATABASE_URL=postgres://compressor:compressor_secure_pw_9x8c7v6b5n4m3@db:5432/compression?sslmode=disable
WORDPRESS_API_URL=https://capcut.ogtemplate.com/wp-json/wp/v2
WORDPRESS_USERNAME=vps
WORDPRESS_APP_PASSWORD=bisf lAxw AsTk Jm2t ytUb 3ENg
POSTGRES_PASSWORD=compressor_secure_pw_9x8c7v6b5n4m3
```

### Step 3: Deploy

Click **"Redeploy"** and wait for all 3 containers to start:
- app (your Go API)
- db (PostgreSQL)
- redis

---

## WHY THIS FIXES IT

- **Before**: Coolify deployed only the Dockerfile → no database → errors
- **After**: Coolify uses docker-compose.coolify.yml → deploys app + database + redis → works!

---

## ALTERNATIVE: If You Can't Change to Docker Compose

If Coolify won't let you change to Docker Compose:

1. Create separate PostgreSQL service in Coolify
2. Create separate Redis service in Coolify  
3. Update DATABASE_URL to point to those services
4. Keep Dockerfile deployment for the app only

But **Docker Compose is much easier** and will work immediately.
